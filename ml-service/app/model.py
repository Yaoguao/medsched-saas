from __future__ import annotations

import hashlib
import json
import os
import re
from datetime import datetime, timezone
from math import exp
from pathlib import Path
from threading import RLock
from typing import Any

import joblib

from app.features import FEATURE_COLUMNS, request_to_features, risk_level
from app.schemas import ModelInfoResponse, PredictionRequest, PredictionResponse


DEFAULT_MODEL_PATH = Path("artifacts/no_show_model.joblib")
DEFAULT_ACTIVE_MANIFEST_PATH = Path("artifacts/active_model.json")


class Predictor:
    def __init__(
        self,
        model_path: Path = DEFAULT_MODEL_PATH,
        manifest_path: Path = DEFAULT_ACTIVE_MANIFEST_PATH,
    ) -> None:
        self.fallback_model_path = model_path
        self.manifest_path = manifest_path
        self.model_path = model_path
        self.artifact: dict[str, Any] | None = None
        self._lock = RLock()
        self.load()

    def load(self) -> None:
        model_path, expected_checksum = self._resolve_model_path()
        if not model_path.exists():
            if self.manifest_path.exists():
                raise ValueError(f"active model artifact does not exist: {model_path}")
            with self._lock:
                self.model_path = model_path
                self.artifact = None
            return

        candidate = joblib.load(model_path)
        self._validate_artifact(candidate)
        if expected_checksum and sha256_file(model_path) != expected_checksum:
            raise ValueError("active model checksum does not match manifest")

        with self._lock:
            self.model_path = model_path
            self.artifact = candidate

    def activate(self, model_version: str, models_dir: Path) -> None:
        if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,127}", model_version):
            raise ValueError("invalid model version")

        model_path = (models_dir / model_version / "model.joblib").resolve()
        models_root = models_dir.resolve()
        if models_root not in model_path.parents:
            raise ValueError("model artifact is outside models directory")
        if not model_path.exists():
            raise ValueError(f"model artifact does not exist: {model_path}")

        candidate = joblib.load(model_path)
        self._validate_artifact(candidate)
        if str(candidate.get("model_version")) != model_version:
            raise ValueError("artifact model version does not match requested version")
        checksum = sha256_file(model_path)

        artifact_path = os.path.relpath(model_path, self.manifest_path.parent)
        manifest = {
            "model_version": model_version,
            "artifact_path": artifact_path,
            "checksum_sha256": checksum,
            "activated_at": datetime.now(timezone.utc).isoformat(),
        }
        atomic_json_write(self.manifest_path, manifest)
        with self._lock:
            self.model_path = model_path
            self.artifact = candidate

    def predict(self, req: PredictionRequest) -> PredictionResponse:
        features = request_to_features(req)
        with self._lock:
            artifact = self.artifact
        if artifact is not None:
            pipeline = artifact["pipeline"]
            probability = float(pipeline.predict_proba(features)[0][1])
            model_version = artifact.get("model_version", "unknown")
            algorithm = artifact.get("algorithm", "unknown")
            thresholds = artifact.get("thresholds")
        else:
            probability = heuristic_probability(features.iloc[0].to_dict())
            model_version = "heuristic-001"
            algorithm = "heuristic"
            thresholds = None

        probability = min(max(probability, 0.0), 1.0)
        return PredictionResponse(
            no_show_probability=round(probability, 4),
            risk_level=risk_level(probability, thresholds),
            model_version=model_version,
            algorithm=algorithm,
            top_factors=top_factors(features.iloc[0].to_dict()),
        )

    def info(self) -> ModelInfoResponse:
        with self._lock:
            artifact = self.artifact
            model_path = self.model_path
        if artifact is None:
            return ModelInfoResponse(
                loaded=False,
                model_version="heuristic-001",
                algorithm="heuristic",
                artifact_path=str(model_path),
                metrics={},
            )

        return ModelInfoResponse(
            loaded=True,
            model_version=str(artifact.get("model_version", "unknown")),
            algorithm=str(artifact.get("algorithm", "unknown")),
            artifact_path=str(model_path),
            metrics=dict(artifact.get("metrics", {})),
        )

    def _resolve_model_path(self) -> tuple[Path, str]:
        if not self.manifest_path.exists():
            return self.fallback_model_path, ""

        manifest = json.loads(self.manifest_path.read_text(encoding="utf-8"))
        artifact_path = str(manifest.get("artifact_path", "")).strip()
        if not artifact_path:
            raise ValueError("active model manifest does not contain artifact_path")
        path = Path(artifact_path)
        if not path.is_absolute():
            path = self.manifest_path.parent / path
        return path, str(manifest.get("checksum_sha256", ""))

    @staticmethod
    def _validate_artifact(artifact: Any) -> None:
        if not isinstance(artifact, dict):
            raise ValueError("model artifact must be a dictionary")
        for key in ("pipeline", "model_version", "algorithm"):
            if key not in artifact:
                raise ValueError(f"model artifact does not contain {key}")
        if not hasattr(artifact["pipeline"], "predict_proba"):
            raise ValueError("model pipeline does not support predict_proba")


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def atomic_json_write(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temp_path = path.with_suffix(path.suffix + ".tmp")
    try:
        temp_path.write_text(json.dumps(payload, indent=2), encoding="utf-8")
        os.replace(temp_path, path)
    finally:
        temp_path.unlink(missing_ok=True)


def heuristic_probability(row: dict[str, Any]) -> float:
    score = -1.35
    score += float(row["previous_no_show_rate"]) * 2.0
    score += min(float(row["lead_time_days"]), 30.0) / 60.0
    score += 0.25 if not row["sms_received"] else -0.15
    score += min(float(row["failed_notifications"]), 3.0) * 0.12
    score += min(float(row["pending_notifications"]), 3.0) * 0.06
    score -= min(float(row["sent_notifications"]), 3.0) * 0.08
    score += 0.20 if row["has_alcoholism"] else 0.0
    score += 0.15 if row["has_handicap"] else 0.0
    score += 0.10 if row["has_hypertension"] else 0.0
    score += 0.10 if row["has_diabetes"] else 0.0
    score += 0.15 if int(row["appointment_weekday"]) >= 5 else 0.0
    return 1.0 / (1.0 + exp(-score))


def top_factors(row: dict[str, Any]) -> list[str]:
    factors: list[tuple[str, float]] = []
    factors.append(("previous_no_show_rate", float(row["previous_no_show_rate"])))
    factors.append(("lead_time_days", min(float(row["lead_time_days"]) / 30.0, 1.0)))
    factors.append(("sms_received", 0.0 if row["sms_received"] else 0.4))
    factors.append(("failed_notifications", min(float(row["failed_notifications"]) / 3.0, 1.0)))
    factors.append(("pending_notifications", min(float(row["pending_notifications"]) / 3.0, 1.0) * 0.5))
    factors.append(("has_alcoholism", 0.3 if row["has_alcoholism"] else 0.0))
    factors.append(("has_handicap", 0.2 if row["has_handicap"] else 0.0))
    factors.append(("appointment_weekday", 0.2 if int(row["appointment_weekday"]) >= 5 else 0.0))
    factors.sort(key=lambda item: item[1], reverse=True)
    selected = [name for name, weight in factors if weight > 0]
    return selected[:3] or FEATURE_COLUMNS[:3]
