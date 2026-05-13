package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://localhost:8000"
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) PredictNoShow(ctx context.Context, req NoShowPredictionRequest) (NoShowPredictionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return NoShowPredictionResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/predict/no-show", bytes.NewReader(body))
	if err != nil {
		return NoShowPredictionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return NoShowPredictionResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return NoShowPredictionResponse{}, fmt.Errorf("ml service returned status %d", resp.StatusCode)
	}

	var prediction NoShowPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return NoShowPredictionResponse{}, err
	}
	return prediction, nil
}

type NoShowPredictionRequest struct {
	Patient     PatientFeatures     `json:"patient"`
	Appointment AppointmentFeatures `json:"appointment"`
	History     HistoryFeatures     `json:"history"`
}

type PatientFeatures struct {
	Age             int    `json:"age"`
	Gender          string `json:"gender"`
	Neighbourhood   string `json:"neighbourhood"`
	Scholarship     bool   `json:"scholarship"`
	HasHypertension bool   `json:"has_hypertension"`
	HasDiabetes     bool   `json:"has_diabetes"`
	HasAlcoholism   bool   `json:"has_alcoholism"`
	HasHandicap     bool   `json:"has_handicap"`
}

type AppointmentFeatures struct {
	ScheduledAt     time.Time `json:"scheduled_at"`
	AppointmentAt   time.Time `json:"appointment_at"`
	BookingChannel  string    `json:"booking_channel"`
	VisitType       string    `json:"visit_type"`
	DoctorSpecialty string    `json:"doctor_specialty"`
	SMSReceived     bool      `json:"sms_received"`
}

type HistoryFeatures struct {
	PreviousAppointmentsCount int `json:"previous_appointments_count"`
	PreviousNoShowCount       int `json:"previous_no_show_count"`
}

type NoShowPredictionResponse struct {
	NoShowProbability float64  `json:"no_show_probability"`
	RiskLevel         string   `json:"risk_level"`
	ModelVersion      string   `json:"model_version"`
	Algorithm         string   `json:"algorithm"`
	TopFactors        []string `json:"top_factors"`
}
