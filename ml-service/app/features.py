from __future__ import annotations

from datetime import datetime
from typing import Any

import pandas as pd

from app.schemas import PredictionRequest


FEATURE_COLUMNS = [
    "age",
    "gender",
    "neighbourhood",
    "scholarship",
    "has_hypertension",
    "has_diabetes",
    "has_alcoholism",
    "has_handicap",
    "sms_received",
    "notification_count",
    "sent_notifications",
    "failed_notifications",
    "pending_notifications",
    "lead_time_days",
    "appointment_weekday",
    "appointment_hour",
    "previous_appointments_count",
    "previous_no_show_count",
    "previous_no_show_rate",
    "booking_channel",
    "visit_type",
    "doctor_specialty",
]

NUMERIC_FEATURES = [
    "age",
    "notification_count",
    "sent_notifications",
    "failed_notifications",
    "pending_notifications",
    "lead_time_days",
    "appointment_weekday",
    "appointment_hour",
    "previous_appointments_count",
    "previous_no_show_count",
    "previous_no_show_rate",
]

CATEGORICAL_FEATURES = [
    "gender",
    "neighbourhood",
    "booking_channel",
    "visit_type",
    "doctor_specialty",
]

BOOLEAN_FEATURES = [
    "scholarship",
    "has_hypertension",
    "has_diabetes",
    "has_alcoholism",
    "has_handicap",
    "sms_received",
]


def request_to_features(req: PredictionRequest) -> pd.DataFrame:
    scheduled_at = req.appointment.scheduled_at
    appointment_at = req.appointment.appointment_at
    lead_time_days = max((appointment_at - scheduled_at).days, 0)
    previous_count = req.history.previous_appointments_count
    previous_no_show_count = req.history.previous_no_show_count
    previous_no_show_rate = (
        previous_no_show_count / previous_count if previous_count > 0 else 0.0
    )

    row = {
        "age": req.patient.age,
        "gender": normalize_text(req.patient.gender),
        "neighbourhood": normalize_text(req.patient.neighbourhood),
        "scholarship": req.patient.scholarship,
        "has_hypertension": req.patient.has_hypertension,
        "has_diabetes": req.patient.has_diabetes,
        "has_alcoholism": req.patient.has_alcoholism,
        "has_handicap": req.patient.has_handicap,
        "sms_received": req.appointment.sms_received,
        "notification_count": req.appointment.notification_count,
        "sent_notifications": req.appointment.sent_notifications,
        "failed_notifications": req.appointment.failed_notifications,
        "pending_notifications": req.appointment.pending_notifications,
        "lead_time_days": lead_time_days,
        "appointment_weekday": appointment_at.weekday(),
        "appointment_hour": appointment_at.hour,
        "previous_appointments_count": previous_count,
        "previous_no_show_count": previous_no_show_count,
        "previous_no_show_rate": previous_no_show_rate,
        "booking_channel": normalize_text(req.appointment.booking_channel),
        "visit_type": normalize_text(req.appointment.visit_type),
        "doctor_specialty": normalize_text(req.appointment.doctor_specialty),
    }
    return pd.DataFrame([row], columns=FEATURE_COLUMNS)


def dataset_to_training_frame(
    csv_path: str,
) -> tuple[pd.DataFrame, pd.Series, pd.DataFrame]:
    df = pd.read_csv(csv_path)
    scheduled = pd.to_datetime(df["ScheduledDay"], errors="coerce")
    appointment = pd.to_datetime(df["AppointmentDay"], errors="coerce")
    lead_time = (appointment - scheduled).dt.days.clip(lower=0).fillna(0)

    training = pd.DataFrame(
        {
            "age": df["Age"].clip(lower=0),
            "gender": df["Gender"].map(normalize_text).fillna("UNKNOWN"),
            "neighbourhood": df["Neighbourhood"].map(normalize_text).fillna("UNKNOWN"),
            "scholarship": to_bool(df["Scholarship"]),
            "has_hypertension": to_bool(df["Hipertension"]),
            "has_diabetes": to_bool(df["Diabetes"]),
            "has_alcoholism": to_bool(df["Alcoholism"]),
            "has_handicap": to_bool(df["Handcap"]),
            "sms_received": to_bool(df["SMS_received"]),
            "notification_count": to_bool(df["SMS_received"]).astype(int),
            "sent_notifications": to_bool(df["SMS_received"]).astype(int),
            "failed_notifications": 0,
            "pending_notifications": 0,
            "lead_time_days": lead_time,
            "appointment_weekday": appointment.dt.weekday.fillna(0),
            "appointment_hour": appointment.dt.hour.fillna(0),
            "previous_appointments_count": 0,
            "previous_no_show_count": 0,
            "previous_no_show_rate": 0.0,
            "booking_channel": "UNKNOWN",
            "visit_type": "UNKNOWN",
            "doctor_specialty": "UNKNOWN",
        },
        columns=FEATURE_COLUMNS,
    )

    showed_up = to_bool(df["Showed_up"])
    target = (~showed_up).astype(int)
    metadata = training_metadata(df["PatientId"], appointment)
    return training, target, metadata


def core_records_to_training_frame(
    records: list[dict[str, Any]],
) -> tuple[pd.DataFrame, pd.Series, pd.DataFrame]:
    df = pd.DataFrame(records)
    if df.empty:
        raise ValueError("core training dataset is empty")

    training = pd.DataFrame(
        {
            "age": df["age"].clip(lower=0),
            "gender": df["gender"].map(normalize_text).fillna("UNKNOWN"),
            "neighbourhood": df["neighbourhood"].map(normalize_text).fillna("UNKNOWN"),
            "scholarship": to_bool(df["scholarship"]),
            "has_hypertension": to_bool(df["has_hypertension"]),
            "has_diabetes": to_bool(df["has_diabetes"]),
            "has_alcoholism": to_bool(df["has_alcoholism"]),
            "has_handicap": to_bool(df["has_handicap"]),
            "sms_received": to_bool(df["sms_received"]),
            "notification_count": numeric_column(df, "notification_count"),
            "sent_notifications": numeric_column(df, "sent_notifications"),
            "failed_notifications": numeric_column(df, "failed_notifications"),
            "pending_notifications": numeric_column(df, "pending_notifications"),
            "lead_time_days": df["lead_time_days"].clip(lower=0),
            "appointment_weekday": df["appointment_weekday"].fillna(0),
            "appointment_hour": df["appointment_hour"].fillna(0),
            "previous_appointments_count": df["previous_appointments_count"].clip(lower=0),
            "previous_no_show_count": df["previous_no_show_count"].clip(lower=0),
            "previous_no_show_rate": df["previous_no_show_rate"].fillna(0.0),
            "booking_channel": df["booking_channel"].map(normalize_text).fillna("UNKNOWN"),
            "visit_type": df["visit_type"].map(normalize_text).fillna("UNKNOWN"),
            "doctor_specialty": df["doctor_specialty"].map(normalize_text).fillna("UNKNOWN"),
        },
        columns=FEATURE_COLUMNS,
    )
    target = to_bool(df["no_show"]).astype(int)
    metadata = training_metadata(df["patient_id"], df["appointment_at"])
    return training, target, metadata


def training_metadata(patient_ids: pd.Series, appointment_times: pd.Series) -> pd.DataFrame:
    metadata = pd.DataFrame(
        {
            "patient_id": patient_ids.map(str),
            "appointment_at": pd.to_datetime(
                appointment_times,
                errors="coerce",
                utc=True,
            ),
        }
    )
    if metadata["patient_id"].str.strip().eq("").any():
        raise ValueError("training dataset contains empty patient_id")
    if metadata["appointment_at"].isna().any():
        raise ValueError("training dataset contains invalid appointment_at")
    return metadata


def to_bool(series: pd.Series) -> pd.Series:
    return series.map(_bool_value).astype(bool)


def numeric_column(df: pd.DataFrame, name: str) -> pd.Series:
    if name not in df:
        return pd.Series(0, index=df.index)
    return pd.to_numeric(df[name], errors="coerce").fillna(0).clip(lower=0)


def normalize_text(value: Any) -> str:
    if value is None:
        return "UNKNOWN"
    text = str(value).strip()
    return text.upper() if text else "UNKNOWN"


def risk_level(
    probability: float,
    thresholds: dict[str, float] | None = None,
) -> str:
    thresholds = thresholds or {"medium": 0.35, "high": 0.65}
    if probability >= thresholds["high"]:
        return "HIGH"
    if probability >= thresholds["medium"]:
        return "MEDIUM"
    return "LOW"


def _bool_value(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    text = str(value).strip().lower()
    return text in {"true", "1", "yes", "y", "sim"}
