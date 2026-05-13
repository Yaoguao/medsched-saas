from __future__ import annotations

from math import exp
from pathlib import Path
from typing import Any

import joblib

from app.features import FEATURE_COLUMNS, request_to_features, risk_level
from app.schemas import PredictionRequest, PredictionResponse


DEFAULT_MODEL_PATH = Path("artifacts/no_show_model.joblib")


class Predictor:
    def __init__(self, model_path: Path = DEFAULT_MODEL_PATH) -> None:
        self.model_path = model_path
        self.artifact: dict[str, Any] | None = None
        self.load()

    def load(self) -> None:
        if self.model_path.exists():
            self.artifact = joblib.load(self.model_path)
        else:
            self.artifact = None

    def predict(self, req: PredictionRequest) -> PredictionResponse:
        features = request_to_features(req)
        if self.artifact is not None:
            pipeline = self.artifact["pipeline"]
            probability = float(pipeline.predict_proba(features)[0][1])
            model_version = self.artifact.get("model_version", "unknown")
            algorithm = self.artifact.get("algorithm", "unknown")
        else:
            probability = heuristic_probability(features.iloc[0].to_dict())
            model_version = "heuristic-001"
            algorithm = "heuristic"

        probability = min(max(probability, 0.0), 1.0)
        return PredictionResponse(
            no_show_probability=round(probability, 4),
            risk_level=risk_level(probability),
            model_version=model_version,
            algorithm=algorithm,
            top_factors=top_factors(features.iloc[0].to_dict()),
        )


def heuristic_probability(row: dict[str, Any]) -> float:
    score = -1.35
    score += float(row["previous_no_show_rate"]) * 2.0
    score += min(float(row["lead_time_days"]), 30.0) / 60.0
    score += 0.25 if not row["sms_received"] else -0.15
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
    factors.append(("has_alcoholism", 0.3 if row["has_alcoholism"] else 0.0))
    factors.append(("has_handicap", 0.2 if row["has_handicap"] else 0.0))
    factors.append(("appointment_weekday", 0.2 if int(row["appointment_weekday"]) >= 5 else 0.0))
    factors.sort(key=lambda item: item[1], reverse=True)
    selected = [name for name, weight in factors if weight > 0]
    return selected[:3] or FEATURE_COLUMNS[:3]
