package contract

import (
	"context"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPIContractIsValidAndCoversApprovedEndpoints(t *testing.T) {
	loader := openapi3.NewLoader()
	document, err := loader.LoadFromFile(filepath.Join("..", "openapi.yaml"))
	if err != nil {
		t.Fatalf("load OpenAPI document: %v", err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatalf("validate OpenAPI document: %v", err)
	}
	if document.OpenAPI != "3.1.0" {
		t.Fatalf("OpenAPI version = %q, want 3.1.0", document.OpenAPI)
	}

	for _, endpoint := range []struct {
		path   string
		method string
	}{
		{"/v1/telemetry/batches", http.MethodPost},
		{"/v1/live/fleet", http.MethodGet},
		{"/v1/live/devices/{deviceId}", http.MethodGet},
		{"/v1/history/devices/{deviceId}", http.MethodGet},
		{"/healthz", http.MethodGet},
		{"/readyz", http.MethodGet},
		{"/metrics", http.MethodGet},
	} {
		operation := requireOperation(t, document, endpoint.path, endpoint.method)
		for status, response := range operation.Responses.Map() {
			if response == nil || response.Value == nil || response.Value.Headers["X-Request-ID"] == nil {
				t.Fatalf("%s response %s must include X-Request-ID", operation.OperationID, status)
			}
		}
	}
}

func TestIngestionContractDocumentsAtomicAdmission(t *testing.T) {
	document := loadContract(t)
	operation := requireOperation(t, document, "/v1/telemetry/batches", http.MethodPost)

	requireSecurity(t, operation, "DeviceAPIKey")
	requireResponses(t, operation, http.StatusAccepted, http.StatusBadRequest, http.StatusUnauthorized, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType, http.StatusTooManyRequests, http.StatusServiceUnavailable)
	requireJSONErrorResponses(t, operation, http.StatusBadRequest, http.StatusUnauthorized, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType, http.StatusTooManyRequests, http.StatusServiceUnavailable)
	if operation.Responses.Status(http.StatusTooManyRequests).Value.Headers["Retry-After"] == nil {
		t.Fatal("429 response must document Retry-After")
	}

	mediaType := operation.RequestBody.Value.Content["application/json"]
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		t.Fatal("ingestion request body must define an application/json schema")
	}
	events := mediaType.Schema.Value.Properties["events"]
	if events == nil || events.Value == nil || events.Value.MinItems != 1 || events.Value.MaxItems == nil || *events.Value.MaxItems != 500 {
		t.Fatal("ingestion events must enforce a batch size of 1 through 500")
	}
}

func TestIngestionContractAllowsOptionalCanonicalUnitsAndEnforcesRanges(t *testing.T) {
	document := loadContract(t)
	for _, rule := range []struct {
		schema string
		unit   string
		min    float64
		max    float64
	}{
		{"TemperatureEvent", "C", -40, 125},
		{"VoltageEvent", "V", 0, 48},
		{"BatteryEvent", "%", 0, 100},
		{"PressureEvent", "kPa", 0, 1000},
	} {
		schema := document.Components.Schemas[rule.schema]
		if schema == nil || schema.Value == nil {
			t.Fatalf("%s schema is missing", rule.schema)
		}
		if slices.Contains(schema.Value.Required, "unit") {
			t.Fatalf("%s must allow an omitted unit and infer its canonical unit", rule.schema)
		}
		value := schema.Value.Properties["value"]
		if value == nil || value.Value == nil || value.Value.Min == nil || value.Value.Max == nil || *value.Value.Min != rule.min || *value.Value.Max != rule.max {
			t.Fatalf("%s must constrain value to [%v, %v]", rule.schema, rule.min, rule.max)
		}
		unit := schema.Value.Properties["unit"]
		if unit == nil || unit.Value == nil || len(unit.Value.Enum) != 1 || unit.Value.Enum[0] != rule.unit {
			t.Fatalf("%s must require unit %q", rule.schema, rule.unit)
		}
	}
}

func TestTimestampContractRequiresRFC3339UTCWireFormat(t *testing.T) {
	document := loadContract(t)
	const utcPattern = "^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]+)?Z$"

	timestamp := document.Components.Schemas["EventTimestamp"]
	if timestamp == nil || timestamp.Value == nil || timestamp.Value.Format != "date-time" || timestamp.Value.Pattern != utcPattern {
		t.Fatal("event timestamps must require RFC 3339 UTC values ending in Z")
	}
	for _, parameterName := range []string{"from", "to"} {
		parameter := requireParameter(t, requireOperation(t, document, "/v1/history/devices/{deviceId}", http.MethodGet), parameterName)
		if parameter.Schema == nil || parameter.Schema.Value == nil || parameter.Schema.Value.Format != "date-time" || parameter.Schema.Value.Pattern != utcPattern {
			t.Fatalf("history %s must require RFC 3339 UTC values ending in Z", parameterName)
		}
	}
}

func TestReadContractsSeparateLiveStateFromBoundedManualHistory(t *testing.T) {
	document := loadContract(t)
	for _, path := range []string{"/v1/live/fleet", "/v1/live/devices/{deviceId}"} {
		operation := requireOperation(t, document, path, http.MethodGet)
		requireSecurity(t, operation, "OperatorBasicAuth")
		requireResponses(t, operation, http.StatusOK, http.StatusUnauthorized, http.StatusServiceUnavailable)
	}

	history := requireOperation(t, document, "/v1/history/devices/{deviceId}", http.MethodGet)
	requireSecurity(t, history, "OperatorBasicAuth")
	requireResponses(t, history, http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized, http.StatusServiceUnavailable)
	for _, name := range []string{"deviceId", "from", "to", "limit"} {
		requireRequiredParameter(t, history, name)
	}
	limit := requireParameter(t, history, "limit")
	if limit.Schema == nil || limit.Schema.Value == nil || limit.Schema.Value.Max == nil || *limit.Schema.Value.Max != 1000 {
		t.Fatal("history limit must be bounded at 1000 events")
	}
	if !strings.Contains(limit.Description, "MVP maximum is 1,000") {
		t.Fatal("history limit must identify the approved 1,000-event MVP maximum")
	}
}

func TestOperationalContractsDocumentThePrometheusTextException(t *testing.T) {
	document := loadContract(t)

	for _, path := range []string{"/healthz", "/readyz"} {
		operation := requireOperation(t, document, path, http.MethodGet)
		for status, response := range operation.Responses.Map() {
			if response == nil || response.Value == nil || response.Value.Content["application/json"] == nil {
				t.Fatalf("%s response %s must use application/json", operation.OperationID, status)
			}
		}
	}

	metrics := requireOperation(t, document, "/metrics", http.MethodGet)
	response := metrics.Responses.Status(http.StatusOK)
	if response == nil || response.Value == nil || response.Value.Content["text/plain; version=0.0.4"] == nil {
		t.Fatal("metrics must use Prometheus text exposition")
	}
	if response.Value.Content["application/json"] != nil {
		t.Fatal("metrics must not advertise application/json")
	}
}

func TestPausedReadinessContractReturnsReadinessState(t *testing.T) {
	document := loadContract(t)
	operation := requireOperation(t, document, "/readyz", http.MethodGet)
	response := operation.Responses.Status(http.StatusServiceUnavailable)
	if response == nil || response.Value == nil {
		t.Fatal("readiness must document the paused-admission response")
	}

	mediaType := response.Value.Content["application/json"]
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		t.Fatal("paused readiness must return a JSON readiness state")
	}
	for _, property := range []string{"status", "admission"} {
		if mediaType.Schema.Value.Properties[property] == nil {
			t.Fatalf("paused readiness response must include %q", property)
		}
	}
}

func loadContract(t *testing.T) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	document, err := loader.LoadFromFile(filepath.Join("..", "openapi.yaml"))
	if err != nil {
		t.Fatalf("load OpenAPI document: %v", err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatalf("validate OpenAPI document: %v", err)
	}
	return document
}

func requireOperation(t *testing.T, document *openapi3.T, path, method string) *openapi3.Operation {
	t.Helper()
	pathItem := document.Paths.Find(path)
	if pathItem == nil {
		t.Fatalf("%s is not documented", path)
	}

	var operation *openapi3.Operation
	switch method {
	case http.MethodGet:
		operation = pathItem.Get
	case http.MethodPost:
		operation = pathItem.Post
	default:
		t.Fatalf("unsupported test method %q", method)
	}
	if operation == nil {
		t.Fatalf("%s %s is not documented", method, path)
	}
	return operation
}

func requireSecurity(t *testing.T, operation *openapi3.Operation, scheme string) {
	t.Helper()
	if operation.Security == nil {
		t.Fatalf("%s must declare %s security", operation.OperationID, scheme)
	}
	for _, requirement := range *operation.Security {
		if _, found := requirement[scheme]; found {
			return
		}
	}
	t.Fatalf("%s must declare %s security", operation.OperationID, scheme)
}

func requireResponses(t *testing.T, operation *openapi3.Operation, statuses ...int) {
	t.Helper()
	for _, status := range statuses {
		response := operation.Responses.Status(status)
		if response == nil || response.Value == nil {
			t.Fatalf("%s must document %d", operation.OperationID, status)
		}
		if response.Value.Headers["X-Request-ID"] == nil {
			t.Fatalf("%s response %d must include X-Request-ID", operation.OperationID, status)
		}
	}
}

func requireJSONErrorResponses(t *testing.T, operation *openapi3.Operation, statuses ...int) {
	t.Helper()
	for _, status := range statuses {
		response := operation.Responses.Status(status)
		if response == nil || response.Value == nil || response.Value.Content["application/json"] == nil {
			t.Fatalf("%s response %d must use the JSON error envelope", operation.OperationID, status)
		}
	}
}

func requireRequiredParameter(t *testing.T, operation *openapi3.Operation, name string) {
	t.Helper()
	if !requireParameter(t, operation, name).Required {
		t.Fatalf("%s must require %q", operation.OperationID, name)
	}
}

func requireParameter(t *testing.T, operation *openapi3.Operation, name string) *openapi3.Parameter {
	t.Helper()
	for _, parameter := range operation.Parameters {
		if parameter.Value != nil && parameter.Value.Name == name {
			return parameter.Value
		}
	}
	t.Fatalf("%s must declare %q", operation.OperationID, name)
	return nil
}
