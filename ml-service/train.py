from __future__ import annotations

import argparse
import json
from importlib.metadata import PackageNotFoundError, version
from pathlib import Path
from typing import Any
from urllib.request import urlopen

import joblib
from sklearn.compose import ColumnTransformer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    accuracy_score,
    average_precision_score,
    f1_score,
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
    parser.add_argument("--out", default="artifacts/no_show_model.joblib")
    parser.add_argument("--metrics", default="artifacts/metrics.json")
    parser.add_argument("--version", default="xgboost-001")
    args = parser.parse_args()

    if args.source == "core":
        x, y = load_core_training_frame(args.core_url, args.core_limit)
    else:
        x, y = dataset_to_training_frame(args.data)
    validate_training_target(y)
    x_train, x_test, y_train, y_test = train_test_split(
        x,
        y,
        test_size=0.2,
        random_state=42,
        stratify=y,
    )

    pipeline, algorithm = build_pipeline(y_train)
    pipeline.fit(x_train, y_train)

    probabilities = pipeline.predict_proba(x_test)[:, 1]
    predictions = (probabilities >= 0.5).astype(int)
    model_version = args.version
    if algorithm != "xgboost" and args.version == "xgboost-001":
        model_version = "logistic-regression-001"

    metrics = {
        "algorithm": algorithm,
        "model_version": model_version,
        "rows": int(len(x)),
        "positive_rate": round(float(y.mean()), 4),
        "roc_auc": round(float(roc_auc_score(y_test, probabilities)), 4),
        "average_precision": round(float(average_precision_score(y_test, probabilities)), 4),
        "accuracy": round(float(accuracy_score(y_test, predictions)), 4),
        "precision": round(float(precision_score(y_test, predictions, zero_division=0)), 4),
        "recall": round(float(recall_score(y_test, predictions, zero_division=0)), 4),
        "f1": round(float(f1_score(y_test, predictions, zero_division=0)), 4),
    }

    artifact = {
        "pipeline": pipeline,
        "model_version": model_version,
        "algorithm": algorithm,
        "metrics": metrics,
    }

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    joblib.dump(artifact, out_path)

    metrics_path = Path(args.metrics)
    metrics_path.parent.mkdir(parents=True, exist_ok=True)
    metrics_path.write_text(json.dumps(metrics, indent=2), encoding="utf-8")
    print(json.dumps(metrics, indent=2))


def build_pipeline(y_train: Any) -> tuple[Pipeline, str]:
    preprocessor = ColumnTransformer(
        transformers=[
            ("num", StandardScaler(), NUMERIC_FEATURES),
            ("bool", "passthrough", BOOLEAN_FEATURES),
            ("cat", OneHotEncoder(handle_unknown="ignore"), CATEGORICAL_FEATURES),
        ]
    )

    try:
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
    except Exception:
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


def load_core_training_frame(core_url: str, limit: int) -> tuple[Any, Any]:
    url = core_url.rstrip("/") + f"/api/v1/ml/training/no-show-dataset?limit={limit}"
    with urlopen(url, timeout=10) as response:
        payload = json.loads(response.read().decode("utf-8"))
    return core_records_to_training_frame(payload.get("items", []))


def validate_training_target(y: Any) -> None:
    if len(y) < 10:
        raise ValueError("training dataset must contain at least 10 completed appointments")
    if len(set(y.tolist())) < 2:
        raise ValueError("training dataset must contain both no-show and showed-up examples")


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


if __name__ == "__main__":
    main()
