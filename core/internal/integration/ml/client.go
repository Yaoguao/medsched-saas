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
	adminToken string
}

func NewClient(baseURL string, timeout time.Duration, adminToken ...string) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://localhost:8000"
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	token := ""
	if len(adminToken) > 0 {
		token = strings.TrimSpace(adminToken[0])
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		adminToken: token,
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

func (c *Client) ModelInfo(ctx context.Context) (ModelInfoResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/model/info", nil)
	if err != nil {
		return ModelInfoResponse{}, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ModelInfoResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ModelInfoResponse{}, fmt.Errorf("ml service returned status %d", resp.StatusCode)
	}

	var info ModelInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return ModelInfoResponse{}, err
	}
	return info, nil
}

func (c *Client) ActivateModel(ctx context.Context, version string) (ModelInfoResponse, error) {
	body, err := json.Marshal(map[string]string{"model_version": version})
	if err != nil {
		return ModelInfoResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/model/activate",
		bytes.NewReader(body),
	)
	if err != nil {
		return ModelInfoResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.adminToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.adminToken)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ModelInfoResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ModelInfoResponse{}, fmt.Errorf("ml service returned status %d", resp.StatusCode)
	}
	var info ModelInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return ModelInfoResponse{}, err
	}
	return info, nil
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
	ScheduledAt          time.Time `json:"scheduled_at"`
	AppointmentAt        time.Time `json:"appointment_at"`
	BookingChannel       string    `json:"booking_channel"`
	VisitType            string    `json:"visit_type"`
	DoctorSpecialty      string    `json:"doctor_specialty"`
	SMSReceived          bool      `json:"sms_received"`
	NotificationCount    int       `json:"notification_count"`
	SentNotifications    int       `json:"sent_notifications"`
	FailedNotifications  int       `json:"failed_notifications"`
	PendingNotifications int       `json:"pending_notifications"`
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

type ModelInfoResponse struct {
	Loaded       bool           `json:"loaded"`
	ModelVersion string         `json:"model_version"`
	Algorithm    string         `json:"algorithm"`
	ArtifactPath string         `json:"artifact_path"`
	Metrics      map[string]any `json:"metrics"`
}
