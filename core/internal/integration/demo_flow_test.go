package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"base/internal/domain/models"
	mlclient "base/internal/integration/ml"
	"base/internal/repository"
	"base/internal/service"
	postgresstorage "base/internal/storage/postgres"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDemoFlowCreateUserAppointmentAndMLPrediction(t *testing.T) {
	//if os.Getenv("RUN_INTEGRATION") != "1" {
	//	t.Skip("set RUN_INTEGRATION=1 and TEST_DATABASE_DSN to run integration demo flow")
	//}

	dsn := "host=localhost user=medcoredb password=medcoredb dbname=medcoredb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is required, example: postgres://user:pass@localhost:5432/db?sslmode=disable")
	}

	ctx := context.Background()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		t.Fatalf("connect test postgres: %v", err)
	}
	if err := postgresstorage.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate test postgres: %v", err)
	}

	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/predict/no-show" {
			http.NotFound(w, r)
			return
		}
		t.Logf("ML fake server received request: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"no_show_probability": 0.73,
			"risk_level":          "HIGH",
			"model_version":       "integration-test-model",
			"algorithm":           "fake-ml",
			"top_factors":         []string{"previous_no_show_rate", "lead_time_days"},
		})
	}))
	defer mlServer.Close()

	svc := service.New(
		repository.NewStore(db),
		service.WithMLClient(mlclient.NewClient(mlServer.URL, 2*time.Second)),
	)

	suffix := time.Now().UnixNano()
	adminEmail := uniqueEmail("flow-admin", suffix)
	doctorEmail := uniqueEmail("flow-doctor", suffix)
	patientEmail := uniqueEmail("flow-patient", suffix)
	adminPhone := uniquePhone(suffix, 0)
	patientPhone := uniquePhone(suffix, 1)
	doctorPhone := uniquePhone(suffix, 2)

	t.Logf("step 1: create admin actor user: %s", adminEmail)
	adminUser, err := svc.CreateStaffUser(ctx, service.CreateStaffUserRequest{
		Role:     models.UserRoleAdmin,
		Email:    adminEmail,
		Phone:    adminPhone,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	t.Logf("step 2: create patient user through public registration: %s", patientEmail)
	patientUser, err := svc.CreateUser(ctx, service.CreateUserRequest{
		Email:    patientEmail,
		Phone:    patientPhone,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create patient user: %v", err)
	}

	t.Logf("step 3: create patient card and medical profile, then link user")
	patient, err := svc.CreatePatient(ctx, models.Patient{
		FullName:      "Интеграционный Пациент",
		Gender:        "M",
		BirthDate:     time.Date(1991, 2, 10, 0, 0, 0, 0, time.UTC),
		Age:           35,
		Phone:         patientPhone,
		Email:         patientEmail,
		Neighbourhood: "TEST",
		Scholarship:   false,
		MedicalProfile: &models.PatientMedicalProfile{
			ChronicConditions: "hypertension",
			RiskFactors:       "integration no-show risk",
			Allergies:         "none",
			Notes:             "created by integration flow",
			HasHypertension:   true,
			HasAlcoholism:     true,
		},
	})
	if err != nil {
		t.Fatalf("create patient: %v", err)
	}
	patient, err = svc.LinkPatientUser(ctx, patient.ID, patientUser.ID)
	if err != nil {
		t.Fatalf("link patient user: %v", err)
	}
	if patient.UserID == nil || *patient.UserID != patientUser.ID {
		t.Fatalf("patient user link was not saved")
	}
	t.Logf("patient linked: patient_id=%s user_id=%s", patient.ID, patientUser.ID)

	t.Logf("step 4: create doctor user and doctor profile")
	doctorUser, err := svc.CreateStaffUser(ctx, service.CreateStaffUserRequest{
		Role:     models.UserRoleDoctor,
		Email:    doctorEmail,
		Phone:    doctorPhone,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create doctor user: %v", err)
	}
	doctor, err := svc.CreateDoctor(ctx, models.Doctor{
		ID:              uuid.New(),
		FullName:        "Интеграционный Врач",
		Specialty:       "Терапевт",
		Cabinet:         "T-101",
		ExperienceYears: 7,
		Status:          "ACTIVE",
	})
	if err != nil {
		t.Fatalf("create doctor: %v", err)
	}
	doctor, err = svc.LinkDoctorUser(ctx, doctor.ID, doctorUser.ID)
	if err != nil {
		t.Fatalf("link doctor user: %v", err)
	}
	t.Logf("doctor linked: doctor_id=%s user_id=%s", doctor.ID, doctorUser.ID)

	t.Logf("step 5: create schedule and free time slot")
	startDate := dateOnly(time.Now()).AddDate(0, 0, 1)
	schedule, err := svc.CreateSchedule(ctx, models.Schedule{
		ID:            uuid.New(),
		DoctorID:      doctor.ID,
		StartDate:     startDate,
		EndDate:       startDate.AddDate(0, 0, 7),
		WorkingDays:   "1,2,3,4,5",
		WorkStartTime: "09:00",
		WorkEndTime:   "18:00",
		Status:        "ACTIVE",
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	slotStart := startDate.Add(10 * time.Hour)
	slot, err := svc.CreateTimeSlot(ctx, models.TimeSlot{
		ID:         uuid.New(),
		ScheduleID: schedule.ID,
		StartAt:    slotStart,
		EndAt:      slotStart.Add(30 * time.Minute),
		Status:     models.SlotStatusFree,
	})
	if err != nil {
		t.Fatalf("create slot: %v", err)
	}
	t.Logf("slot ready: slot_id=%s start=%s", slot.ID, slot.StartAt.Format(time.RFC3339))

	t.Logf("step 6: create appointment")
	appointment, err := svc.CreateAppointment(ctx, service.CreateAppointmentRequest{
		PatientID:      patient.ID,
		TimeSlotID:     slot.ID,
		BookingChannel: models.BookingChannelRegistry,
		VisitType:      models.VisitTypePrimary,
		ChangedBy:      &adminUser.ID,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}
	t.Logf("appointment created: appointment_id=%s status=%s doctor_id=%s patient_id=%s", appointment.ID, appointment.Status, appointment.DoctorID, appointment.PatientID)

	t.Logf("step 7: request ML no-show prediction through core service")
	prediction, err := svc.PredictAppointmentNoShow(ctx, appointment.ID)
	if err != nil {
		t.Fatalf("predict appointment no-show: %v", err)
	}
	t.Logf("prediction saved: prediction_id=%s probability=%.2f risk=%s model=%s", prediction.ID, prediction.Probability, prediction.RiskLevel, prediction.ModelVersion)

	if prediction.AppointmentID != appointment.ID {
		t.Fatalf("prediction appointment mismatch: got %s want %s", prediction.AppointmentID, appointment.ID)
	}
	if prediction.RiskLevel != models.PredictionRiskHigh {
		t.Fatalf("expected high risk prediction, got %s", prediction.RiskLevel)
	}

	loadedAppointment, err := svc.GetAppointment(ctx, appointment.ID)
	if err != nil {
		t.Fatalf("load appointment after prediction: %v", err)
	}
	if loadedAppointment.PredictionResult == nil {
		t.Fatalf("appointment has no prediction result after ML calculation")
	}
	t.Logf("flow completed: appointment_id=%s prediction_result_id=%s notifications=%d", loadedAppointment.ID, loadedAppointment.PredictionResult.ID, len(loadedAppointment.Notifications))
}

func uniqueEmail(prefix string, suffix int64) string {
	return prefix + "-" + time.Unix(0, suffix).Format("20060102150405.000000000") + "@example.com"
}

func uniquePhone(suffix int64, offset int64) string {
	return fmt.Sprintf("+79%09d", (suffix/1000+offset)%1_000_000_000)
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
