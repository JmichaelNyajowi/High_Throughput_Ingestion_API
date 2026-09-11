package ingestion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/example/telemetry/api/internal/ingestion/credentials"
	"github.com/example/telemetry/api/internal/platform/httpserver"
)

const maxTelemetryBodyBytes = 1 << 20

var rfc3339UTCPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?Z$`)

type validatedBatchContextKey struct{}

type measurementRule struct {
	unit string
	min  float64
	max  float64
}

var measurementRules = map[string]measurementRule{
	"temperature": {unit: "C", min: -40, max: 125},
	"voltage":     {unit: "V", min: 0, max: 48},
	"battery":     {unit: "%", min: 0, max: 100},
	"pressure":    {unit: "kPa", min: 0, max: 1000},
}

type BatchValidator struct {
	now func() time.Time
}

type ValidatedBatch struct {
	Device credentials.AuthorizedDevice
	Events []ValidatedEvent
}

type ValidatedEvent struct {
	EventID             string
	DeviceID            string
	Timestamp           time.Time
	MeasurementType     string
	Value               float64
	Unit                string
	SkipLiveAggregation bool
}

type validationError struct {
	status  int
	code    string
	message string
}

type telemetryBatchPayload struct {
	Events json.RawMessage `json:"events"`
}

type telemetryEventPayload struct {
	EventID         string       `json:"event_id"`
	DeviceID        string       `json:"device_id"`
	Timestamp       string       `json:"timestamp"`
	MeasurementType string       `json:"measurement_type"`
	Value           *float64     `json:"value"`
	Unit            optionalUnit `json:"unit"`
}

type optionalUnit struct {
	present bool
	value   string
}

func (unit *optionalUnit) UnmarshalJSON(value []byte) error {
	unit.present = true
	if bytes.Equal(value, []byte("null")) {
		return errors.New("unit must be a string")
	}
	return json.Unmarshal(value, &unit.value)
}

func NewBatchValidator(now func() time.Time) BatchValidator {
	if now == nil {
		now = time.Now
	}
	return BatchValidator{now: now}
}

// Middleware must follow DeviceAuthenticator so the batch is bound to the
// device identity established at the authentication boundary.
func (validator BatchValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		device, found := AuthorizedDeviceFromContext(r.Context())
		if !found {
			httpserver.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication failed")
			return
		}

		batch, err := validator.Validate(r, device)
		if err != nil {
			httpserver.WriteError(w, r, err.status, err.code, err.message)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), validatedBatchContextKey{}, batch)))
	})
}

// Validate applies the entire request contract before a batch can enter the
// rate-limiting or queue stages. It never performs dependency I/O.
func (validator BatchValidator) Validate(r *http.Request, device credentials.AuthorizedDevice) (ValidatedBatch, *validationError) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return ValidatedBatch{}, invalid(http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
	}
	if r.ContentLength > maxTelemetryBodyBytes {
		return ValidatedBatch{}, invalid(http.StatusRequestEntityTooLarge, "payload_too_large", "Request body exceeds 1 MB")
	}
	if r.Body == nil {
		return ValidatedBatch{}, invalidRequest()
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxTelemetryBodyBytes+1))
	if err != nil {
		return ValidatedBatch{}, invalidRequest()
	}
	if len(body) > maxTelemetryBodyBytes {
		return ValidatedBatch{}, invalid(http.StatusRequestEntityTooLarge, "payload_too_large", "Request body exceeds 1 MB")
	}
	if !utf8.Valid(body) {
		return ValidatedBatch{}, invalidRequest()
	}

	var payload telemetryBatchPayload
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil || ensureOnlyOneJSONValue(decoder) != nil {
		return ValidatedBatch{}, invalidRequest()
	}
	return validator.validateEvents(payload.Events, device)
}

func (validator BatchValidator) validateEvents(rawEvents json.RawMessage, device credentials.AuthorizedDevice) (ValidatedBatch, *validationError) {
	decoder := json.NewDecoder(bytes.NewReader(rawEvents))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('[') {
		return ValidatedBatch{}, invalidRequest()
	}

	batch := ValidatedBatch{Device: device, Events: make([]ValidatedEvent, 0, 1)}
	now := validator.now().UTC()
	for decoder.More() {
		if len(batch.Events) == 500 {
			return ValidatedBatch{}, invalidRequest()
		}

		var event telemetryEventPayload
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&event); err != nil {
			return ValidatedBatch{}, invalidRequest()
		}
		validated, err := validateEvent(event, device.ExternalID, now)
		if err != nil {
			return ValidatedBatch{}, err
		}
		batch.Events = append(batch.Events, validated)
	}
	if _, err := decoder.Token(); err != nil || len(batch.Events) == 0 || ensureOnlyOneJSONValue(decoder) != nil {
		return ValidatedBatch{}, invalidRequest()
	}
	return batch, nil
}

func validateEvent(event telemetryEventPayload, authorizedExternalID string, now time.Time) (ValidatedEvent, *validationError) {
	if !validIdentifier(event.EventID) || !validIdentifier(event.DeviceID) || event.DeviceID != authorizedExternalID {
		return ValidatedEvent{}, invalidRequest()
	}
	rule, found := measurementRules[event.MeasurementType]
	if !found || event.Value == nil || (event.Unit.present && event.Unit.value != rule.unit) || math.IsNaN(*event.Value) || math.IsInf(*event.Value, 0) || *event.Value < rule.min || *event.Value > rule.max {
		return ValidatedEvent{}, invalidRequest()
	}
	if !rfc3339UTCPattern.MatchString(event.Timestamp) {
		return ValidatedEvent{}, invalidRequest()
	}
	timestamp, err := time.Parse(time.RFC3339Nano, event.Timestamp)
	if err != nil {
		return ValidatedEvent{}, invalidRequest()
	}
	return ValidatedEvent{
		EventID:             event.EventID,
		DeviceID:            event.DeviceID,
		Timestamp:           timestamp,
		MeasurementType:     event.MeasurementType,
		Value:               *event.Value,
		Unit:                rule.unit,
		SkipLiveAggregation: timestamp.Before(now.Add(-60 * time.Second)),
	}, nil
}

func validIdentifier(value string) bool {
	return utf8.RuneCountInString(value) >= 1 && utf8.RuneCountInString(value) <= 128 && strings.TrimSpace(value) != ""
}

func ensureOnlyOneJSONValue(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return err
	}
	return nil
}

func invalidRequest() *validationError {
	return invalid(http.StatusBadRequest, "invalid_request", "Invalid telemetry batch")
}

func invalid(status int, code, message string) *validationError {
	return &validationError{status: status, code: code, message: message}
}

func ValidatedBatchFromContext(ctx context.Context) (ValidatedBatch, bool) {
	batch, found := ctx.Value(validatedBatchContextKey{}).(ValidatedBatch)
	return batch, found
}
