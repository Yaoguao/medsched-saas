package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
}

func NewClient(baseURL string, timeout time.Duration, enabled bool) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://localhost:8082"
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		enabled: enabled,
	}
}

func (c *Client) Schedule(ctx context.Context, req ScheduleNotificationRequest) error {
	if c == nil || !c.enabled {
		return nil
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/notify", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("notifier returned status %d", resp.StatusCode)
	}
	return nil
}

type ScheduleNotificationRequest struct {
	ID            uuid.UUID         `json:"id,omitempty"`
	ExternalID    string            `json:"external_id,omitempty"`
	AppointmentID uuid.UUID         `json:"appointment_id,omitempty"`
	PatientID     uuid.UUID         `json:"patient_id,omitempty"`
	Channel       string            `json:"channel"`
	Recipient     string            `json:"recipient,omitempty"`
	Subject       string            `json:"subject,omitempty"`
	Message       string            `json:"message"`
	SendAt        time.Time         `json:"send_at"`
	Variables     map[string]string `json:"variables,omitempty"`
}
