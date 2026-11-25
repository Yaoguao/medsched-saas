import hmac
import os
from pathlib import Path

from fastapi import FastAPI, Header, HTTPException, status

from app.model import DEFAULT_ACTIVE_MANIFEST_PATH, DEFAULT_MODEL_PATH, Predictor
from app.schemas import (
    ModelActivationRequest,
    ModelInfoResponse,
    PredictionRequest,
    PredictionResponse,
)


app = FastAPI(title="MedSched ML Service", version="0.1.0")
predictor = Predictor(
    Path(os.getenv("ML_MODEL_PATH", str(DEFAULT_MODEL_PATH))),
    Path(os.getenv("ML_ACTIVE_MANIFEST_PATH", str(DEFAULT_ACTIVE_MANIFEST_PATH))),
)


@app.get("/health")
def health() -> dict[str, str | bool]:
    info = predictor.info()
    return {
        "status": "ok",
        "model_loaded": info.loaded,
        "model_version": info.model_version,
        "algorithm": info.algorithm,
    }


@app.get("/model/info", response_model=ModelInfoResponse)
def model_info() -> ModelInfoResponse:
    return predictor.info()


@app.post("/model/reload", response_model=ModelInfoResponse)
def reload_model(authorization: str = Header(default="")) -> ModelInfoResponse:
    require_admin_token(authorization)
    try:
        predictor.load()
    except Exception as exc:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"model reload failed: {exc}",
        ) from exc
    return predictor.info()


@app.post("/model/activate", response_model=ModelInfoResponse)
def activate_model(
    req: ModelActivationRequest,
    authorization: str = Header(default=""),
) -> ModelInfoResponse:
    require_admin_token(authorization)
    try:
        predictor.activate(
            req.model_version,
            Path(os.getenv("ML_MODELS_DIR", "artifacts/models")),
        )
    except Exception as exc:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"model activation failed: {exc}",
        ) from exc
    return predictor.info()


@app.post("/predict/no-show", response_model=PredictionResponse)
def predict_no_show(req: PredictionRequest) -> PredictionResponse:
    return predictor.predict(req)


def require_admin_token(authorization: str) -> None:
    expected = os.getenv("ML_ADMIN_TOKEN", "")
    if not expected:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="model management is disabled",
        )
    prefix = "Bearer "
    actual = authorization[len(prefix):].strip() if authorization.startswith(prefix) else ""
    if not actual or not hmac.compare_digest(actual, expected):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="unauthorized",
        )
