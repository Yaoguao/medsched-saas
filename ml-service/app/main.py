from pathlib import Path

from fastapi import FastAPI

from app.model import DEFAULT_MODEL_PATH, Predictor
from app.schemas import PredictionRequest, PredictionResponse


app = FastAPI(title="MedSched ML Service", version="0.1.0")
predictor = Predictor(Path(DEFAULT_MODEL_PATH))


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/predict/no-show", response_model=PredictionResponse)
def predict_no_show(req: PredictionRequest) -> PredictionResponse:
    return predictor.predict(req)
