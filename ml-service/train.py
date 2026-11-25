from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
from dataclasses import dataclass
from datetime import datetime, timezone
from importlib.metadata import PackageNotFoundError, version
from pathlib import Path
from typing import Any
from urllib.request import Request, urlopen

import joblib
import numpy as np
import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    accuracy_score,
    average_precision_score,
    confusion_matrix,
    f1_score,
    fbeta_score,
    precision_score,
    recall_score,
    roc_auc_score,
)
from sklearn.model_selection import train_test_split
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, StandardScaler

from app.features import (
    BOOLEAN_FEATURES,
    CATEGORICAL_FEATURES,
    NUMERIC_FEATURES,
    core_records_to_training_frame,
    dataset_to_training_frame,
)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", default="../healthcare_noshows.csv")
    parser.add_argument("--source", choices=["csv", "core"], default="csv")
    parser.add_argument("--core-url", default="http://localhost:8081")
    parser.add_argument("--core-limit", type=int, default=5000)
    parser.add_argument("--core-token", default=os.getenv("CORE_SERVICE_TOKEN", ""))
    parser.add_argument("--models-dir", default="artifacts/models")
    parser.add_argument("--active-manifest", default="artifacts/active_model.json")
    parser.add_argument("--metrics", default="artifacts/metrics.json")
    parser.add_argument("--version")
    parser.add_argument("--algorithm", choices=["xgboost", "logistic_regression"], default="xgboost")
    parser.add_argument("--split-mode", choices=["temporal_group", "random"], default="temporal_group")
    parser.add_argument("--activate", action="store_true")
    args = parser.parse_args()

    if args.source == "core":
        x, y, metadata = load_core_training_frame(
            args.core_url,
            args.core_limit,
            args.core_token,
        )
    else:
        x, y, metadata = dataset_to_training_frame(args.data)
    validate_training_target(y, minimum_rows=30 if args.split_mode == "temporal_group" else 10)
    split = split_training_data(x, y, metadata, args.split_mode)

    pipeline, algorithm = build_pipeline(split.y_train, args.algorithm)
    pipeline.fit(split.x_train, split.y_train)

    validation_probabilities = pipeline.predict_proba(split.x_validation)[:, 1]
    thresholds = select_risk_thresholds(
        split.y_validation,
        validation_probabilities,
    )
    probabilities = pipeline.predict_proba(split.x_test)[:, 1]
    predictions = (probabilities >= thresholds["medium"]).astype(int)
    trained_at = datetime.now(timezone.utc)
    model_version = args.version or default_model_version(algorithm, trained_at)
    validate_model_version(model_version)

    metrics = {
        "algorithm": algorithm,
        "model_version": model_version,
        "trained_at": trained_at.isoformat(),
        "rows": int(len(x)),
        "positive_rate": round(float(y.mean()), 4),
        "decision_threshold": thresholds["medium"],
        "risk_thresholds": thresholds,
        "roc_auc": round(float(roc_auc_score(split.y_test, probabilities)), 4),
        "average_precision": round(float(average_precision_score(split.y_test, probabilities)), 4),
        "accuracy": round(float(accuracy_score(split.y_test, predictions)), 4),
        "precision": round(float(precision_score(split.y_test, predictions, zero_division=0)), 4),
        "recall": round(float(recall_score(split.y_test, predictions, zero_division=0)), 4),
        "f1": round(float(f1_score(split.y_test, predictions, zero_division=0)), 4),
        "confusion_matrix": confusion_matrix(split.y_test, predictions).tolist(),
        "risk_summary": risk_summary(split.y_test, probabilities, thresholds),
        "split": split.summary,
    }

    artifact = {
        "pipeline": pipeline,
        "model_version": model_version,
        "algorithm": algorithm,
        "metrics": metrics,
        "thresholds": thresholds,
        "feature_schema_version": "no-show-v1",
    }

    version_dir = Path(args.models_dir) / model_version
    version_dir.mkdir(parents=True, exist_ok=False)
    model_path = version_dir / "model.joblib"
    atomic_joblib_dump(artifact, model_path)
    checksum = sha256_file(model_path)

    metrics_path = Path(args.metrics)
    metrics_path.parent.mkdir(parents=True, exist_ok=True)
    atomic_json_write(metrics_path, metrics)
    atomic_json_write(version_dir / "metrics.json", metrics)
    atomic_json_write(
        version_dir / "features.json",
        {
            "schema_version": "no-show-v1",
            "numeric": NUMERIC_FEATURES,
            "boolean": BOOLEAN_FEATURES,
            "categorical": CATEGORICAL_FEATURES,
        },
    )
    atomic_json_write(
        version_dir / "training.json",
        {
            "source": args.source,
            "rows": int(len(x)),
            "positive_rate": round(float(y.mean()), 4),
            "trained_at": trained_at.isoformat(),
            "checksum_sha256": checksum,
            "split": split.summary,
            "thresholds": thresholds,
        },
    )

    if args.activate:
        manifest_path = Path(args.active_manifest)
        artifact_path = os.path.relpath(model_path, manifest_path.parent)
        atomic_json_write(
            manifest_path,
            {
                "model_version": model_version,
                "artifact_path": artifact_path,
                "checksum_sha256": checksum,
                "activated_at": datetime.now(timezone.utc).isoformat(),
            },
        )

    result = dict(metrics)
    result["artifact_path"] = str(model_path)
    result["checksum_sha256"] = checksum
    result["feature_schema_version"] = "no-show-v1"
    result["activated"] = args.activate
    print(json.dumps(result, indent=2))


def build_pipeline(y_train: Any, requested_algorithm: str) -> tuple[Pipeline, str]:
    preprocessor = ColumnTransformer(
        transformers=[
            ("num", StandardScaler(), NUMERIC_FEATURES),
            ("bool", "passthrough", BOOLEAN_FEATURES),
            ("cat", OneHotEncoder(handle_unknown="ignore"), CATEGORICAL_FEATURES),
        ]
    )

    if requested_algorithm == "xgboost":
        ensure_xgboost_compatible()
        from xgboost import XGBClassifier

        negative = int((y_train == 0).sum())
        positive = max(int((y_train == 1).sum()), 1)
        estimator = XGBClassifier(
            n_estimators=250,
            max_depth=4,
            learning_rate=0.05,
            subsample=0.9,
            colsample_bytree=0.8,
            objective="binary:logistic",
            eval_metric="logloss",
            scale_pos_weight=negative / positive,
            random_state=42,
            n_jobs=4,
        )
        algorithm = "xgboost"
    else:
        estimator = LogisticRegression(
            max_iter=1000,
            class_weight="balanced",
            solver="lbfgs",
        )
        algorithm = "logistic_regression"

    return Pipeline(
        steps=[
            ("features", preprocessor),
            ("model", estimator),
        ]
    ), algorithm


def load_core_training_frame(
    core_url: str,
    limit: int,
    token: str,
) -> tuple[Any, Any, Any]:
    if not token:
        raise ValueError(
            "CORE_SERVICE_TOKEN or --core-token is required when --source core is used"
        )
    url = core_url.rstrip("/") + f"/api/v1/ml/training/no-show-dataset?limit={limit}"
    request = Request(url, headers={"Authorization": f"Bearer {token}"})
    with urlopen(request, timeout=10) as response:
        payload = json.loads(response.read().decode("utf-8"))
    return core_records_to_training_frame(payload.get("items", []))


def validate_training_target(y: Any, minimum_rows: int) -> None:
    if len(y) < minimum_rows:
        raise ValueError(
            f"training dataset must contain at least {minimum_rows} completed appointments"
        )
    if len(set(y.tolist())) < 2:
        raise ValueError("training dataset must contain both no-show and showed-up examples")


@dataclass(frozen=True)
class TrainingSplit:
    x_train: pd.DataFrame
    x_validation: pd.DataFrame
    x_test: pd.DataFrame
    y_train: pd.Series
    y_validation: pd.Series
    y_test: pd.Series
    summary: dict[str, Any]


def split_training_data(
    x: pd.DataFrame,
    y: pd.Series,
    metadata: pd.DataFrame,
    mode: str,
) -> TrainingSplit:
    x = x.reset_index(drop=True)
    y = y.reset_index(drop=True)
    metadata = metadata.reset_index(drop=True)
    if mode == "random":
        return random_training_split(x, y, metadata)
    return temporal_group_training_split(x, y, metadata)


def temporal_group_training_split(
    x: pd.DataFrame,
    y: pd.Series,
    metadata: pd.DataFrame,
) -> TrainingSplit:
    ordered = metadata.sort_values("appointment_at", kind="stable").index.to_numpy()
    test_count = max(int(round(len(ordered) * 0.15)), 2)
    test_indices = ordered[-test_count:]
    test_patients = set(metadata.loc[test_indices, "patient_id"])

    development_indices = np.array(
        [
            index
            for index in ordered[:-test_count]
            if metadata.at[index, "patient_id"] not in test_patients
        ],
        dtype=int,
    )
    validation_count = max(int(round(len(ordered) * 0.15)), 2)
    if len(development_indices) <= validation_count:
        raise ValueError(
            "not enough independent patient history for temporal validation split"
        )

    validation_indices = development_indices[-validation_count:]
    validation_patients = set(metadata.loc[validation_indices, "patient_id"])
    train_indices = np.array(
        [
            index
            for index in development_indices[:-validation_count]
            if metadata.at[index, "patient_id"] not in validation_patients
        ],
        dtype=int,
    )

    split = build_split(
        x,
        y,
        metadata,
        train_indices,
        validation_indices,
        test_indices,
        mode="temporal_group",
        leakage_protected=True,
    )
    validate_split(split)
    return split


def random_training_split(
    x: pd.DataFrame,
    y: pd.Series,
    metadata: pd.DataFrame,
) -> TrainingSplit:
    indices = np.arange(len(x))
    development_indices, test_indices = train_test_split(
        indices,
        test_size=0.2,
        random_state=42,
        stratify=y,
    )
    train_indices, validation_indices = train_test_split(
        development_indices,
        test_size=0.25,
        random_state=42,
        stratify=y.iloc[development_indices],
    )
    split = build_split(
        x,
        y,
        metadata,
        np.asarray(train_indices),
        np.asarray(validation_indices),
        np.asarray(test_indices),
        mode="random",
        leakage_protected=False,
    )
    validate_split(split)
    return split


def build_split(
    x: pd.DataFrame,
    y: pd.Series,
    metadata: pd.DataFrame,
    train_indices: np.ndarray,
    validation_indices: np.ndarray,
    test_indices: np.ndarray,
    mode: str,
    leakage_protected: bool,
) -> TrainingSplit:
    train_meta = metadata.iloc[train_indices]
    validation_meta = metadata.iloc[validation_indices]
    test_meta = metadata.iloc[test_indices]
    train_patients = set(train_meta["patient_id"])
    validation_patients = set(validation_meta["patient_id"])
    test_patients = set(test_meta["patient_id"])
    patient_overlap = {
        "train_validation": len(train_patients & validation_patients),
        "train_test": len(train_patients & test_patients),
        "validation_test": len(validation_patients & test_patients),
    }
    if leakage_protected and any(patient_overlap.values()):
        raise ValueError("patient leakage detected between training splits")
    used_rows = len(train_indices) + len(validation_indices) + len(test_indices)
    summary = {
        "mode": mode,
        "leakage_protected": leakage_protected,
        "rows": {
            "train": int(len(train_indices)),
            "validation": int(len(validation_indices)),
            "test": int(len(test_indices)),
            "dropped_for_group_isolation": int(len(x) - used_rows),
        },
        "patients": {
            "train": int(train_meta["patient_id"].nunique()),
            "validation": int(validation_meta["patient_id"].nunique()),
            "test": int(test_meta["patient_id"].nunique()),
        },
        "patient_overlap": patient_overlap,
        "date_ranges": {
            "train": date_range(train_meta),
            "validation": date_range(validation_meta),
            "test": date_range(test_meta),
        },
    }
    return TrainingSplit(
        x_train=x.iloc[train_indices].reset_index(drop=True),
        x_validation=x.iloc[validation_indices].reset_index(drop=True),
        x_test=x.iloc[test_indices].reset_index(drop=True),
        y_train=y.iloc[train_indices].reset_index(drop=True),
        y_validation=y.iloc[validation_indices].reset_index(drop=True),
        y_test=y.iloc[test_indices].reset_index(drop=True),
        summary=summary,
    )


def validate_split(split: TrainingSplit) -> None:
    for name, target in (
        ("train", split.y_train),
        ("validation", split.y_validation),
        ("test", split.y_test),
    ):
        if len(target) < 2:
            raise ValueError(f"{name} split is too small")
        if target.nunique() < 2:
            raise ValueError(f"{name} split must contain both target classes")


def date_range(metadata: pd.DataFrame) -> dict[str, str]:
    if metadata.empty:
        return {"from": "", "to": ""}
    return {
        "from": metadata["appointment_at"].min().isoformat(),
        "to": metadata["appointment_at"].max().isoformat(),
    }


def select_risk_thresholds(
    target: pd.Series,
    probabilities: np.ndarray,
) -> dict[str, float]:
    candidates = np.arange(0.05, 0.96, 0.01)
    medium_scores = [
        fbeta_score(target, probabilities >= threshold, beta=2, zero_division=0)
        for threshold in candidates
    ]
    medium = float(candidates[int(np.argmax(medium_scores))])

    high_candidates = candidates[candidates >= min(medium + 0.05, 0.95)]
    if len(high_candidates) == 0:
        high = min(medium + 0.05, 0.99)
    else:
        target_array = target.to_numpy()
        minimum_high_rows = max(20, int(round(len(target) * 0.01)))
        eligible: list[tuple[float, float]] = []
        for threshold in high_candidates:
            mask = probabilities >= threshold
            if int(mask.sum()) < minimum_high_rows:
                continue
            precision = float(target_array[mask].mean())
            eligible.append((float(threshold), precision))

        target_precision = 0.4
        meeting_target = [
            (threshold, precision)
            for threshold, precision in eligible
            if precision >= target_precision
        ]
        if meeting_target:
            high = min(meeting_target, key=lambda item: item[0])[0]
        elif eligible:
            high = max(eligible, key=lambda item: (item[1], -item[0]))[0]
        else:
            high = min(medium + 0.2, 0.95)

    return {
        "medium": round(min(medium, 0.9), 4),
        "high": round(min(max(high, medium + 0.05), 0.95), 4),
    }


def risk_summary(
    target: pd.Series,
    probabilities: np.ndarray,
    thresholds: dict[str, float],
) -> dict[str, dict[str, float | int]]:
    levels = np.where(
        probabilities >= thresholds["high"],
        "HIGH",
        np.where(probabilities >= thresholds["medium"], "MEDIUM", "LOW"),
    )
    summary: dict[str, dict[str, float | int]] = {}
    target_array = target.to_numpy()
    for level in ("LOW", "MEDIUM", "HIGH"):
        mask = levels == level
        count = int(mask.sum())
        summary[level] = {
            "count": count,
            "no_show_rate": round(float(target_array[mask].mean()), 4) if count else 0.0,
        }
    return summary


def ensure_xgboost_compatible() -> None:
    try:
        sklearn_version = version("scikit-learn")
        xgboost_version = version("xgboost")
    except PackageNotFoundError as exc:
        raise ImportError(str(exc)) from exc

    if version_tuple(sklearn_version) >= (1, 6, 0) and version_tuple(xgboost_version) < (2, 1, 4):
        raise ImportError(
            "xgboost>=2.1.4 is required with scikit-learn>=1.6.0"
        )


def version_tuple(raw: str) -> tuple[int, int, int]:
    parts = raw.split(".")
    parsed: list[int] = []
    for part in parts[:3]:
        digits = ""
        for char in part:
            if not char.isdigit():
                break
            digits += char
        parsed.append(int(digits or 0))
    while len(parsed) < 3:
        parsed.append(0)
    return parsed[0], parsed[1], parsed[2]


def default_model_version(algorithm: str, trained_at: datetime) -> str:
    timestamp = trained_at.strftime("%Y%m%d-%H%M%S")
    return f"no-show-{algorithm.replace('_', '-')}-{timestamp}"


def validate_model_version(model_version: str) -> None:
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,127}", model_version):
        raise ValueError(
            "model version must contain only letters, digits, dots, underscores and hyphens"
        )


def atomic_joblib_dump(artifact: dict[str, Any], path: Path) -> None:
    temp_path = path.with_suffix(path.suffix + ".tmp")
    try:
        joblib.dump(artifact, temp_path)
        os.replace(temp_path, path)
    finally:
        temp_path.unlink(missing_ok=True)


def atomic_json_write(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temp_path = path.with_suffix(path.suffix + ".tmp")
    try:
        temp_path.write_text(json.dumps(payload, indent=2), encoding="utf-8")
        os.replace(temp_path, path)
    finally:
        temp_path.unlink(missing_ok=True)


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


if __name__ == "__main__":
    main()
