from datetime import datetime
from typing import Literal

from pydantic import BaseModel, Field


RiskLevel = Literal["LOW", "MEDIUM", "HIGH"]


class PatientFeatures(BaseModel):
    age: int = Field(ge=0, le=130)
    gender: str = "UNKNOWN"
    neighbourhood: str = "UNKNOWN"
    scholarship: bool = False
    has_hypertension: bool = False
    has_diabetes: bool = False
    has_alcoholism: bool = False
    has_handicap: bool = False


class AppointmentFeatures(BaseModel):
    scheduled_at: datetime
    appointment_at: datetime
    booking_channel: str = "UNKNOWN"
    visit_type: str = "UNKNOWN"
    doctor_specialty: str = "UNKNOWN"
    sms_received: bool = False


class HistoryFeatures(BaseModel):
    previous_appointments_count: int = Field(default=0, ge=0)
    previous_no_show_count: int = Field(default=0, ge=0)


class PredictionRequest(BaseModel):
    patient: PatientFeatures
    appointment: AppointmentFeatures
    history: HistoryFeatures = Field(default_factory=HistoryFeatures)


class PredictionResponse(BaseModel):
    no_show_probability: float
    risk_level: RiskLevel
    model_version: str
    algorithm: str
    top_factors: list[str]
