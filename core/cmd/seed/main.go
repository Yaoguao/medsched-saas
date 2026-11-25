package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"base/internal/config"
	"base/internal/domain/models"
	"base/internal/lib/logger/sl"
	"base/internal/storage/postgres"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const demoPassword = "password123"

func main() {
	cfg := config.MustLoad()
	log := sl.InitLogger(cfg.Env, os.Stdout)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := postgres.Open(ctx, cfg.StorageConfig, log)
	if err != nil {
		log.Error("failed to connect postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := postgres.AutoMigrate(db); err != nil {
		log.Error("failed to migrate database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return seed(ctx, tx)
	}); err != nil {
		log.Error("failed to seed demo data", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info(
		"demo data ready",
		slog.Int("doctors", len(doctorSeeds)),
		slog.Int("patients", len(patientSeeds)),
		slog.Int("linked_patient_accounts", linkedPatientCount()),
		slog.Int("historical_appointments", 60),
		slog.String("admin", "admin@example.com"),
		slog.String("registrator", "registrator@example.com"),
		slog.String("manager", "manager@example.com"),
		slog.String("it_support", "it@example.com"),
		slog.String("password", demoPassword),
	)
}

func seed(ctx context.Context, db *gorm.DB) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	ehrProvider, err := ensureDemoEHRProvider(ctx, db)
	if err != nil {
		return err
	}

	staff := make(map[models.UserRole]models.User)
	for _, spec := range staffSeeds {
		user, err := ensureUser(ctx, db, spec.Email, spec.Phone, spec.Role, string(passwordHash))
		if err != nil {
			return err
		}
		if _, exists := staff[spec.Role]; !exists {
			staff[spec.Role] = user
		}
	}
	if _, err := ensureUser(
		ctx,
		db,
		"available.patient@example.com",
		"+79991000999",
		models.UserRolePatient,
		string(passwordHash),
	); err != nil {
		return err
	}

	doctors := make([]models.Doctor, 0, len(doctorSeeds))
	schedules := make(map[uuid.UUID]models.Schedule, len(doctorSeeds))
	for _, spec := range doctorSeeds {
		user, err := ensureUser(ctx, db, spec.Email, spec.Phone, models.UserRoleDoctor, string(passwordHash))
		if err != nil {
			return err
		}
		doctor, err := ensureDoctor(ctx, db, user.ID, spec)
		if err != nil {
			return err
		}
		schedule, err := ensureSchedule(ctx, db, doctor.ID)
		if err != nil {
			return err
		}
		doctors = append(doctors, doctor)
		schedules[doctor.ID] = schedule
	}

	patients := make([]models.Patient, 0, len(patientSeeds))
	for index, spec := range patientSeeds {
		var userID *uuid.UUID
		if spec.LinkAccount {
			user, err := ensureUser(ctx, db, spec.Email, spec.Phone, models.UserRolePatient, string(passwordHash))
			if err != nil {
				return err
			}
			userID = &user.ID
		}
		patient, err := ensurePatient(ctx, db, userID, spec)
		if err != nil {
			return err
		}
		if err := ensureMedicalProfile(ctx, db, patient.ID, spec); err != nil {
			return err
		}
		if spec.WithEHR {
			if err := ensureEHRLink(ctx, db, ehrProvider.ID, &patient, index); err != nil {
				return err
			}
		}
		patients = append(patients, patient)
	}

	actor := staff[models.UserRoleAdmin]
	manager := staff[models.UserRoleManager]
	if err := seedHistoricalAppointments(ctx, db, actor.ID, patients, doctors, schedules); err != nil {
		return err
	}
	if err := seedCurrentWorkspace(ctx, db, actor.ID, manager.ID, patients, doctors, schedules); err != nil {
		return err
	}
	return nil
}

type staffSeed struct {
	Email string
	Phone string
	Role  models.UserRole
}

var staffSeeds = []staffSeed{
	{Email: "admin@example.com", Phone: "+79991000001", Role: models.UserRoleAdmin},
	{Email: "registrator@example.com", Phone: "+79991000002", Role: models.UserRoleRegistrator},
	{Email: "registrator2@example.com", Phone: "+79991000003", Role: models.UserRoleRegistrator},
	{Email: "manager@example.com", Phone: "+79991000004", Role: models.UserRoleManager},
	{Email: "it@example.com", Phone: "+79991000005", Role: models.UserRoleITSupport},
}

type doctorSeed struct {
	Email           string
	Phone           string
	FullName        string
	Specialty       string
	Cabinet         string
	ExperienceYears int
}

var doctorSeeds = []doctorSeed{
	{Email: "doctor.therapist@example.com", Phone: "+79992000001", FullName: "Петров Петр Петрович", Specialty: "Терапевт", Cabinet: "201", ExperienceYears: 12},
	{Email: "doctor.cardiologist@example.com", Phone: "+79992000002", FullName: "Смирнова Анна Сергеевна", Specialty: "Кардиолог", Cabinet: "305", ExperienceYears: 9},
	{Email: "doctor.neurologist@example.com", Phone: "+79992000003", FullName: "Волков Максим Олегович", Specialty: "Невролог", Cabinet: "214", ExperienceYears: 15},
	{Email: "doctor.endocrinologist@example.com", Phone: "+79992000004", FullName: "Орлова Екатерина Ильинична", Specialty: "Эндокринолог", Cabinet: "307", ExperienceYears: 8},
	{Email: "doctor.surgeon@example.com", Phone: "+79992000005", FullName: "Федоров Андрей Викторович", Specialty: "Хирург", Cabinet: "118", ExperienceYears: 18},
}

type patientSeed struct {
	Email             string
	Phone             string
	FullName          string
	Gender            string
	BirthDate         time.Time
	Neighbourhood     string
	Scholarship       bool
	ChronicConditions string
	RiskFactors       string
	Allergies         string
	Notes             string
	HasHypertension   bool
	HasDiabetes       bool
	HasAlcoholism     bool
	HasHandicap       bool
	LinkAccount       bool
	WithEHR           bool
}

var patientSeeds = []patientSeed{
	patient("patient01@example.com", "+79993000001", "Иванов Иван Иванович", "M", 1990, 1, 15, "Центральный", "Артериальная гипертензия", "Ранее пропускал приемы", "Нет", true, false, false, false, true, true),
	patient("patient02@example.com", "+79993000002", "Кузнецова Мария Андреевна", "F", 1984, 7, 9, "Северный", "Сахарный диабет 2 типа", "Требуется утреннее время", "Пенициллин", false, true, false, false, true, true),
	patient("patient03@example.com", "+79993000003", "Орлов Дмитрий Павлович", "M", 1975, 3, 22, "Южный", "Сердечно-сосудистый риск", "Нерегулярное подтверждение", "Нет", true, false, true, false, true, false),
	patient("patient04@example.com", "+79993000004", "Соколова Елена Викторовна", "F", 2001, 11, 4, "Западный", "Нет", "Первичное обращение", "Нет", false, false, false, false, true, false),
	patient("patient05@example.com", "+79993000005", "Морозов Алексей Игоревич", "M", 1968, 5, 30, "Восточный", "Гипертензия, диабет", "Ограниченная мобильность", "Аспирин", true, true, false, true, true, true),
	patient("patient06@example.com", "+79993000006", "Новикова Ольга Романовна", "F", 1996, 2, 18, "Центральный", "Мигрень", "Повторный прием", "Нет", false, false, false, false, true, false),
	patient("patient07@example.com", "+79993000007", "Лебедев Сергей Николаевич", "M", 1988, 9, 12, "Северный", "Нет", "Длительное ожидание приема", "Нет", false, false, false, false, true, true),
	patient("patient08@example.com", "+79993000008", "Козлова Наталья Игоревна", "F", 1959, 12, 3, "Южный", "Гипертензия", "Нуждается в напоминаниях", "Йод", true, false, false, false, true, true),
	patient("patient09@example.com", "+79993000009", "Васильев Артем Денисович", "M", 1993, 6, 27, "Западный", "Нет", "Часто меняет время приема", "Нет", false, false, false, false, true, false),
	patient("patient10@example.com", "+79993000010", "Павлова Ирина Михайловна", "F", 1979, 4, 8, "Восточный", "Сахарный диабет 2 типа", "Контроль терапии", "Латекс", false, true, false, false, true, true),
	patient("patient11@example.com", "+79993000011", "Семенов Кирилл Алексеевич", "M", 2003, 8, 16, "Центральный", "Нет", "Студент", "Нет", false, false, false, false, true, false),
	patient("patient12@example.com", "+79993000012", "Голубева Татьяна Олеговна", "F", 1986, 10, 25, "Северный", "Гипотиреоз", "Диспансерное наблюдение", "Нет", false, false, false, false, true, true),
	patient("patient13@example.com", "+79993000013", "Виноградов Роман Евгеньевич", "M", 1972, 1, 6, "Южный", "Гипертензия", "Не отвечает на SMS", "Нет", true, false, true, false, true, false),
	patient("patient14@example.com", "+79993000014", "Богданова Светлана Юрьевна", "F", 1998, 5, 19, "Западный", "Нет", "Первичное обращение", "Цитрусовые", false, false, false, false, true, false),
	patient("patient15@example.com", "+79993000015", "Воробьев Михаил Степанович", "M", 1965, 11, 29, "Восточный", "Диабет, нейропатия", "Необходим сопровождающий", "Нет", false, true, false, true, true, true),
	patient("patient16@example.com", "+79993000016", "Фомина Анастасия Львовна", "F", 1991, 3, 14, "Центральный", "Нет", "Предпочитает онлайн-запись", "Нет", false, false, false, false, true, false),
	patient("patient17@example.com", "+79993000017", "Тихонов Виктор Аркадьевич", "M", 1982, 7, 1, "Северный", "Гипертензия", "Нет аккаунта пациента", "Нет", true, false, false, false, false, true),
	patient("patient18@example.com", "+79993000018", "Медведева Людмила Павловна", "F", 1956, 9, 21, "Южный", "Ишемическая болезнь сердца", "Телефонная запись", "Контрастное вещество", true, false, false, true, false, true),
	patient("patient19@example.com", "+79993000019", "Егоров Денис Валерьевич", "M", 1994, 12, 10, "Западный", "Нет", "Нет аккаунта пациента", "Нет", false, false, false, false, false, false),
	patient("patient20@example.com", "+79993000020", "Николаева Полина Артемовна", "F", 2000, 4, 2, "Восточный", "Нет", "Первичный прием", "Нет", false, false, false, false, false, false),
}

func patient(
	email, phone, fullName, gender string,
	year int,
	month time.Month,
	day int,
	neighbourhood, conditions, risks, allergies string,
	hypertension, diabetes, alcoholism, handicap, linkAccount, withEHR bool,
) patientSeed {
	return patientSeed{
		Email: email, Phone: phone, FullName: fullName, Gender: gender,
		BirthDate:     time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		Neighbourhood: neighbourhood, Scholarship: year >= 2000,
		ChronicConditions: conditions, RiskFactors: risks, Allergies: allergies,
		Notes:           "Демонстрационная медицинская карта",
		HasHypertension: hypertension, HasDiabetes: diabetes,
		HasAlcoholism: alcoholism, HasHandicap: handicap,
		LinkAccount: linkAccount, WithEHR: withEHR,
	}
}

func ensureUser(
	ctx context.Context,
	db *gorm.DB,
	email, phone string,
	role models.UserRole,
	passwordHash string,
) (models.User, error) {
	var user models.User
	err := db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err == nil {
		user.Role = role
		user.Phone = phone
		user.Status = "ACTIVE"
		if user.PasswordHash == "" {
			user.PasswordHash = passwordHash
		}
		return user, db.WithContext(ctx).Save(&user).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, err
	}
	user = models.User{
		ID: uuid.New(), Role: role, Email: email, Phone: phone,
		PasswordHash: passwordHash, Status: "ACTIVE",
	}
	return user, db.WithContext(ctx).Create(&user).Error
}

func ensureDoctor(ctx context.Context, db *gorm.DB, userID uuid.UUID, spec doctorSeed) (models.Doctor, error) {
	var doctor models.Doctor
	err := db.WithContext(ctx).First(&doctor, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		doctor = models.Doctor{ID: uuid.New(), UserID: &userID}
	} else if err != nil {
		return models.Doctor{}, err
	}
	doctor.FullName = spec.FullName
	doctor.Specialty = spec.Specialty
	doctor.Cabinet = spec.Cabinet
	doctor.ExperienceYears = spec.ExperienceYears
	doctor.Status = "ACTIVE"
	if doctor.CreatedAt.IsZero() {
		err = db.WithContext(ctx).Create(&doctor).Error
	} else {
		err = db.WithContext(ctx).Save(&doctor).Error
	}
	return doctor, err
}

func ensurePatient(ctx context.Context, db *gorm.DB, userID *uuid.UUID, spec patientSeed) (models.Patient, error) {
	var patientModel models.Patient
	err := db.WithContext(ctx).First(&patientModel, "email = ?", spec.Email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		patientModel = models.Patient{ID: uuid.New()}
	} else if err != nil {
		return models.Patient{}, err
	}
	patientModel.UserID = userID
	patientModel.FullName = spec.FullName
	patientModel.Gender = spec.Gender
	patientModel.BirthDate = spec.BirthDate
	patientModel.Age = ageAt(spec.BirthDate, time.Now())
	patientModel.Phone = spec.Phone
	patientModel.Email = spec.Email
	patientModel.Neighbourhood = spec.Neighbourhood
	patientModel.Scholarship = spec.Scholarship
	if patientModel.CreatedAt.IsZero() {
		err = db.WithContext(ctx).Create(&patientModel).Error
	} else {
		err = db.WithContext(ctx).Save(&patientModel).Error
	}
	return patientModel, err
}

func ensureMedicalProfile(ctx context.Context, db *gorm.DB, patientID uuid.UUID, spec patientSeed) error {
	var profile models.PatientMedicalProfile
	err := db.WithContext(ctx).First(&profile, "patient_id = ?", patientID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = models.PatientMedicalProfile{ID: uuid.New(), PatientID: patientID}
	} else if err != nil {
		return err
	}
	profile.ChronicConditions = spec.ChronicConditions
	profile.RiskFactors = spec.RiskFactors
	profile.Allergies = spec.Allergies
	profile.Notes = spec.Notes
	profile.HasHypertension = spec.HasHypertension
	profile.HasDiabetes = spec.HasDiabetes
	profile.HasAlcoholism = spec.HasAlcoholism
	profile.HasHandicap = spec.HasHandicap
	profile.UpdatedAtProfile = time.Now().UTC()
	if profile.CreatedAt.IsZero() {
		return db.WithContext(ctx).Create(&profile).Error
	}
	return db.WithContext(ctx).Save(&profile).Error
}

func ensureDemoEHRProvider(ctx context.Context, db *gorm.DB) (models.IntegrationProvider, error) {
	var provider models.IntegrationProvider
	err := db.WithContext(ctx).First(&provider, "code = ?", "demo-ehr").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		provider = models.IntegrationProvider{
			ID:      uuid.New(),
			Code:    "demo-ehr",
			Name:    "Demo EHR",
			Type:    "MOCK",
			Status:  "CONNECTED",
			Enabled: true,
		}
		err = db.WithContext(ctx).Create(&provider).Error
	}
	return provider, err
}

func ensureEHRLink(
	ctx context.Context,
	db *gorm.DB,
	providerID uuid.UUID,
	patientModel *models.Patient,
	index int,
) error {
	syncedAt := time.Now().UTC().Add(-time.Duration(index+1) * time.Hour)
	patientModel.ExternalEHRID = fmt.Sprintf("DEMO-EHR-%04d", index+1)
	patientModel.LastEHRSyncAt = &syncedAt
	if err := db.WithContext(ctx).Save(patientModel).Error; err != nil {
		return err
	}
	var link models.ExternalEntityLink
	err := db.WithContext(ctx).First(
		&link,
		"provider_id = ? AND entity_type = ? AND internal_id = ?",
		providerID,
		"patient",
		patientModel.ID,
	).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		link = models.ExternalEntityLink{
			ID:           uuid.New(),
			ProviderID:   providerID,
			EntityType:   "patient",
			InternalID:   patientModel.ID,
			ExternalID:   patientModel.ExternalEHRID,
			LastSyncedAt: &syncedAt,
		}
		err = db.WithContext(ctx).Create(&link).Error
	} else if err == nil {
		link.ExternalID = patientModel.ExternalEHRID
		link.LastSyncedAt = &syncedAt
		err = db.WithContext(ctx).Save(&link).Error
	}
	if err != nil {
		return err
	}
	var count int64
	if err := db.WithContext(ctx).Model(&models.IntegrationEvent{}).
		Where("provider = ? AND entity_type = ? AND entity_id = ? AND action = ?",
			"demo-ehr", "patient", patientModel.ID, "PROFILE_SYNC").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&models.IntegrationEvent{
		ID: uuid.New(), Provider: "demo-ehr", EntityType: "patient",
		EntityID: patientModel.ID, Action: "PROFILE_SYNC", Status: "SUCCESS",
		Message: "Профиль синхронизирован из Demo EHR",
		AppliedFields: "chronic_conditions,risk_factors,allergies,notes," +
			"has_hypertension,has_diabetes,has_alcoholism,has_handicap",
		OccurredAt: syncedAt,
	}).Error
}

func ensureSchedule(ctx context.Context, db *gorm.DB, doctorID uuid.UUID) (models.Schedule, error) {
	startDate := dateOnly(time.Now().UTC()).AddDate(0, 0, -70)
	endDate := dateOnly(time.Now().UTC()).AddDate(0, 0, 35)
	var schedule models.Schedule
	err := db.WithContext(ctx).First(&schedule, "doctor_id = ? AND start_date = ?", doctorID, startDate).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		schedule = models.Schedule{ID: uuid.New(), DoctorID: doctorID}
	} else if err != nil {
		return models.Schedule{}, err
	}
	schedule.StartDate = startDate
	schedule.EndDate = endDate
	schedule.WorkingDays = "1,2,3,4,5"
	schedule.WorkStartTime = "09:00"
	schedule.WorkEndTime = "18:00"
	schedule.Status = "ACTIVE"
	if schedule.CreatedAt.IsZero() {
		err = db.WithContext(ctx).Create(&schedule).Error
	} else {
		err = db.WithContext(ctx).Save(&schedule).Error
	}
	return schedule, err
}

func seedHistoricalAppointments(
	ctx context.Context,
	db *gorm.DB,
	actorID uuid.UUID,
	patients []models.Patient,
	doctors []models.Doctor,
	schedules map[uuid.UUID]models.Schedule,
) error {
	today := dateOnly(time.Now().UTC())
	for index := 0; index < 60; index++ {
		appointmentDay := businessDayOffset(today, -(65 - index))
		doctor := doctors[index%len(doctors)]
		patientModel := patients[(index*7+index/5)%len(patients)]
		startAt := appointmentDay.Add(time.Duration(9+(index%7)) * time.Hour)
		slot, err := ensureTimeSlot(ctx, db, schedules[doctor.ID].ID, doctor.ID, startAt)
		if err != nil {
			return err
		}
		noShow := index%6 == 0 || (patientSeeds[(index*7+index/5)%len(patientSeeds)].HasAlcoholism && index%4 == 0)
		status := models.AppointmentStatusCompleted
		if noShow {
			status = models.AppointmentStatusNoShow
		}
		appointmentModel, err := ensureAppointment(ctx, db, actorID, patientModel.ID, doctor.ID, slot, appointmentSeed{
			Status: status, VisitType: visitTypeByIndex(index),
			Channel: channelByIndex(index), ReminderSent: index%4 != 0,
			ScheduledAt: startAt.AddDate(0, 0, -(2 + index%21)),
		})
		if err != nil {
			return err
		}
		notificationStatus := models.NotificationStatusDelivered
		if index%9 == 0 {
			notificationStatus = models.NotificationStatusFailed
		}
		if err := ensureNotification(ctx, db, appointmentModel, notificationStatus, "history"); err != nil {
			return err
		}
		probability := historicalProbability(noShow, index)
		if err := ensurePrediction(ctx, db, appointmentModel.ID, probability); err != nil {
			return err
		}
	}
	return nil
}

func seedCurrentWorkspace(
	ctx context.Context,
	db *gorm.DB,
	actorID, managerID uuid.UUID,
	patients []models.Patient,
	doctors []models.Doctor,
	schedules map[uuid.UUID]models.Schedule,
) error {
	today := dateOnly(time.Now().UTC())
	futureSlots := make(map[uuid.UUID][]models.TimeSlot, len(doctors))
	for doctorIndex, doctor := range doctors {
		for slotIndex := 0; slotIndex < 18; slotIndex++ {
			day := businessDayOffset(today, 1+slotIndex/6)
			startAt := day.Add(time.Duration(9+(slotIndex%6)) * time.Hour)
			slot, err := ensureTimeSlot(ctx, db, schedules[doctor.ID].ID, doctor.ID, startAt)
			if err != nil {
				return err
			}
			if slotIndex == 17 && doctorIndex%2 == 0 {
				slot.Status = models.SlotStatusBlocked
				slot.IsBlocked = true
				slot.BlockReason = "Совещание отделения"
				if err := db.WithContext(ctx).Save(&slot).Error; err != nil {
					return err
				}
			}
			futureSlots[doctor.ID] = append(futureSlots[doctor.ID], slot)
		}
	}

	for index := 0; index < 14; index++ {
		doctor := doctors[index%len(doctors)]
		slot := futureSlots[doctor.ID][index/len(doctors)]
		patientModel := patients[(index*3+1)%len(patients)]
		status := models.AppointmentStatusCreated
		if index%3 != 0 {
			status = models.AppointmentStatusConfirmed
		}
		appointmentModel, err := ensureAppointment(ctx, db, actorID, patientModel.ID, doctor.ID, slot, appointmentSeed{
			Status: status, VisitType: visitTypeByIndex(index + 1),
			Channel: channelByIndex(index + 2), ReminderSent: status == models.AppointmentStatusConfirmed,
			ScheduledAt: time.Now().UTC().Add(-time.Duration(index+1) * time.Hour),
		})
		if err != nil {
			return err
		}
		notifyStatus := models.NotificationStatusPending
		if status == models.AppointmentStatusConfirmed {
			notifyStatus = models.NotificationStatusSent
		}
		if err := ensureNotification(ctx, db, appointmentModel, notifyStatus, "future"); err != nil {
			return err
		}
		if index%2 == 0 {
			if err := ensurePrediction(ctx, db, appointmentModel.ID, 0.16+float64((index*13)%70)/100); err != nil {
				return err
			}
		}
	}

	primaryDoctor := doctors[0]
	primarySlots := futureSlots[primaryDoctor.ID]
	var appointmentModel models.Appointment
	if err := db.WithContext(ctx).First(&appointmentModel, "time_slot_id = ?", primarySlots[0].ID).Error; err != nil {
		return err
	}
	if err := ensureRescheduleProposal(
		ctx,
		db,
		appointmentModel.ID,
		primarySlots[5].ID,
		managerID,
		models.RescheduleProposalPending,
		"Высокий риск неявки: предложено более удобное время",
	); err != nil {
		return err
	}

	cancelDoctor := doctors[4]
	cancelSlot := futureSlots[cancelDoctor.ID][4]
	cancelPatient := patients[18]
	_, err := ensureAppointment(ctx, db, actorID, cancelPatient.ID, cancelDoctor.ID, cancelSlot, appointmentSeed{
		Status: models.AppointmentStatusCancelled, VisitType: models.VisitTypeConsultation,
		Channel: models.BookingChannelPhone, ScheduledAt: time.Now().UTC().Add(-48 * time.Hour),
	})
	return err
}

type appointmentSeed struct {
	Status       models.AppointmentStatus
	VisitType    models.VisitType
	Channel      models.BookingChannel
	ReminderSent bool
	ScheduledAt  time.Time
}

func ensureTimeSlot(
	ctx context.Context,
	db *gorm.DB,
	scheduleID, doctorID uuid.UUID,
	startAt time.Time,
) (models.TimeSlot, error) {
	var slot models.TimeSlot
	err := db.WithContext(ctx).First(&slot, "doctor_id = ? AND start_at = ?", doctorID, startAt).Error
	if err == nil {
		return slot, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.TimeSlot{}, err
	}
	slot = models.TimeSlot{
		ID: uuid.New(), ScheduleID: scheduleID, DoctorID: doctorID,
		StartAt: startAt, EndAt: startAt.Add(30 * time.Minute), Status: models.SlotStatusFree,
	}
	return slot, db.WithContext(ctx).Create(&slot).Error
}

func ensureAppointment(
	ctx context.Context,
	db *gorm.DB,
	actorID, patientID, doctorID uuid.UUID,
	slot models.TimeSlot,
	spec appointmentSeed,
) (models.Appointment, error) {
	var appointmentModel models.Appointment
	err := db.WithContext(ctx).First(&appointmentModel, "time_slot_id = ?", slot.ID).Error
	isNew := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !isNew {
		return models.Appointment{}, err
	}
	if isNew {
		appointmentModel = models.Appointment{
			ID: uuid.New(), PatientID: patientID, DoctorID: doctorID, TimeSlotID: slot.ID,
		}
	}
	appointmentModel.ScheduledAt = spec.ScheduledAt
	appointmentModel.AppointmentAt = slot.StartAt
	appointmentModel.Status = spec.Status
	appointmentModel.BookingChannel = spec.Channel
	appointmentModel.VisitType = spec.VisitType
	appointmentModel.ReminderSent = spec.ReminderSent
	applyAppointmentOutcome(&appointmentModel, spec.Status)
	if isNew {
		err = db.WithContext(ctx).Create(&appointmentModel).Error
	} else {
		err = db.WithContext(ctx).Save(&appointmentModel).Error
	}
	if err != nil {
		return models.Appointment{}, err
	}

	slot.Status = models.SlotStatusBooked
	if spec.Status == models.AppointmentStatusCancelled {
		slot.Status = models.SlotStatusFree
	}
	if err := db.WithContext(ctx).Save(&slot).Error; err != nil {
		return models.Appointment{}, err
	}
	if err := ensureStatusHistory(ctx, db, appointmentModel, actorID); err != nil {
		return models.Appointment{}, err
	}
	return appointmentModel, nil
}

func applyAppointmentOutcome(appointmentModel *models.Appointment, status models.AppointmentStatus) {
	appointmentModel.ShowedUp = nil
	appointmentModel.CompletedAt = nil
	appointmentModel.CancelledAt = nil
	appointmentModel.ConfirmedAt = nil
	switch status {
	case models.AppointmentStatusCompleted:
		showedUp := true
		completedAt := appointmentModel.AppointmentAt.Add(30 * time.Minute)
		appointmentModel.ShowedUp = &showedUp
		appointmentModel.CompletedAt = &completedAt
	case models.AppointmentStatusNoShow:
		showedUp := false
		completedAt := appointmentModel.AppointmentAt.Add(30 * time.Minute)
		appointmentModel.ShowedUp = &showedUp
		appointmentModel.CompletedAt = &completedAt
	case models.AppointmentStatusCancelled:
		cancelledAt := appointmentModel.ScheduledAt.Add(time.Hour)
		appointmentModel.CancelledAt = &cancelledAt
	case models.AppointmentStatusConfirmed:
		confirmedAt := appointmentModel.ScheduledAt.Add(15 * time.Minute)
		appointmentModel.ConfirmedAt = &confirmedAt
	}
}

func ensureStatusHistory(
	ctx context.Context,
	db *gorm.DB,
	appointmentModel models.Appointment,
	actorID uuid.UUID,
) error {
	var count int64
	if err := db.WithContext(ctx).Model(&models.AppointmentStatusHistory{}).
		Where("appointment_id = ? AND new_status = ?", appointmentModel.ID, appointmentModel.Status).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&models.AppointmentStatusHistory{
		ID: uuid.New(), AppointmentID: appointmentModel.ID,
		NewStatus: appointmentModel.Status, ChangedBy: &actorID,
		Reason:    "Актуальные демонстрационные данные",
		ChangedAt: appointmentModel.ScheduledAt,
	}).Error
}

func ensureNotification(
	ctx context.Context,
	db *gorm.DB,
	appointmentModel models.Appointment,
	status models.NotificationStatus,
	scope string,
) error {
	externalID := fmt.Sprintf("seed-%s-%s", scope, appointmentModel.ID)
	var notification models.Notification
	err := db.WithContext(ctx).First(&notification, "external_id = ?", externalID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		notification = models.Notification{ID: uuid.New(), ExternalID: externalID}
	} else if err != nil {
		return err
	}
	notification.AppointmentID = appointmentModel.ID
	notification.PatientID = appointmentModel.PatientID
	notification.Channel = notificationChannelByAppointment(appointmentModel)
	notification.Type = models.NotificationTypeReminder
	notification.Status = status
	notification.ScheduledAt = appointmentModel.AppointmentAt.Add(-24 * time.Hour)
	notification.Provider = "seed"
	notification.ErrorMessage = ""
	if status == models.NotificationStatusFailed {
		notification.ErrorMessage = "Временная ошибка демонстрационного провайдера"
	}
	if status == models.NotificationStatusSent || status == models.NotificationStatusDelivered {
		sentAt := notification.ScheduledAt.Add(2 * time.Minute)
		notification.SentAt = &sentAt
	}
	if notification.CreatedAt.IsZero() {
		return db.WithContext(ctx).Create(&notification).Error
	}
	return db.WithContext(ctx).Save(&notification).Error
}

func ensurePrediction(ctx context.Context, db *gorm.DB, appointmentID uuid.UUID, probability float64) error {
	if probability > 0.99 {
		probability = 0.99
	}
	var prediction models.PredictionResult
	err := db.WithContext(ctx).First(&prediction, "appointment_id = ?", appointmentID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		prediction = models.PredictionResult{ID: uuid.New(), AppointmentID: appointmentID}
	} else if err != nil {
		return err
	}
	now := time.Now().UTC()
	prediction.Probability = probability
	prediction.RiskLevel = predictionRiskLevel(probability)
	prediction.ModelVersion = "seed-baseline-2026"
	prediction.RequestedAt = now
	prediction.ReceivedAt = now
	if prediction.CreatedAt.IsZero() {
		return db.WithContext(ctx).Create(&prediction).Error
	}
	return db.WithContext(ctx).Save(&prediction).Error
}

func ensureRescheduleProposal(
	ctx context.Context,
	db *gorm.DB,
	appointmentID, slotID, actorID uuid.UUID,
	status models.RescheduleProposalStatus,
	reason string,
) error {
	var proposal models.AppointmentRescheduleProposal
	err := db.WithContext(ctx).
		First(&proposal, "appointment_id = ? AND proposed_time_slot_id = ?", appointmentID, slotID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		proposal = models.AppointmentRescheduleProposal{
			ID: uuid.New(), AppointmentID: appointmentID, ProposedTimeSlotID: slotID,
		}
	} else if err != nil {
		return err
	}
	proposal.RequestedBy = &actorID
	proposal.Reason = reason
	proposal.Status = status
	if proposal.CreatedAt.IsZero() {
		return db.WithContext(ctx).Create(&proposal).Error
	}
	return db.WithContext(ctx).Save(&proposal).Error
}

func historicalProbability(noShow bool, index int) float64 {
	if noShow {
		return 0.68 + float64(index%25)/100
	}
	return 0.08 + float64(index%38)/100
}

func predictionRiskLevel(probability float64) models.PredictionRiskLevel {
	switch {
	case probability >= 0.74:
		return models.PredictionRiskHigh
	case probability >= 0.25:
		return models.PredictionRiskMedium
	default:
		return models.PredictionRiskLow
	}
}

func channelByIndex(index int) models.BookingChannel {
	channels := []models.BookingChannel{
		models.BookingChannelWeb,
		models.BookingChannelRegistry,
		models.BookingChannelPhone,
	}
	return channels[index%len(channels)]
}

func visitTypeByIndex(index int) models.VisitType {
	types := []models.VisitType{
		models.VisitTypePrimary,
		models.VisitTypeRepeat,
		models.VisitTypeConsultation,
		models.VisitTypeDiagnostic,
	}
	return types[index%len(types)]
}

func notificationChannelByAppointment(appointmentModel models.Appointment) models.NotificationChannel {
	switch appointmentModel.BookingChannel {
	case models.BookingChannelWeb:
		return models.NotificationChannelPush
	case models.BookingChannelPhone:
		return models.NotificationChannelSMS
	default:
		return models.NotificationChannelEmail
	}
}

func ageAt(birthDate, now time.Time) int {
	age := now.Year() - birthDate.Year()
	birthday := time.Date(now.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(birthday) {
		age--
	}
	return age
}

func linkedPatientCount() int {
	count := 0
	for _, spec := range patientSeeds {
		if spec.LinkAccount {
			count++
		}
	}
	return count
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func businessDayOffset(start time.Time, offset int) time.Time {
	if offset == 0 {
		return start
	}
	step := 1
	if offset < 0 {
		step = -1
	}
	current := start
	for moved := 0; moved != offset; moved += step {
		current = current.AddDate(0, 0, step)
		for current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			current = current.AddDate(0, 0, step)
		}
	}
	return current
}
