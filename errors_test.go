package centreon

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		HTTPStatus: http.StatusForbidden,
		Code:       42,
		Message:    "access denied",
	}
	got := err.Error()
	want := "centreon API error (HTTP 403): access denied"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAPIError_ErrorsAs(t *testing.T) {
	var base error = &APIError{HTTPStatus: http.StatusInternalServerError, Message: "boom"}

	// errors.AsType (Go 1.26)
	apiErr, ok := errors.AsType[*APIError](base)
	if !ok {
		t.Fatal("errors.AsType should match *APIError")
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want 500", apiErr.HTTPStatus)
	}

	// errors.AsType (Go 1.26)
	got, ok := errors.AsType[*APIError](base)
	if !ok {
		t.Fatal("errors.AsType should match *APIError")
	}
	if got.Message != "boom" {
		t.Errorf("Message = %q, want %q", got.Message, "boom")
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{Resource: "host", ID: 42}
	got := err.Error()
	want := "centreon: host with ID 42 not found"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNotFoundError_ErrorsAs(t *testing.T) {
	var base error = &NotFoundError{Resource: "service", ID: 7}

	got, ok := errors.AsType[*NotFoundError](base)
	if !ok {
		t.Fatal("errors.AsType should match *NotFoundError")
	}
	if got.Resource != "service" || got.ID != 7 {
		t.Errorf("got Resource=%q ID=%d, want service/7", got.Resource, got.ID)
	}
}

func TestParseError_JSONBody(t *testing.T) {
	body := `{"code":42,"message":"invalid parameter"}`
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	err := parseError(resp)
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatal("expected *APIError")
	}
	if apiErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d, want 400", apiErr.HTTPStatus)
	}
	if apiErr.Code != 42 {
		t.Errorf("Code = %d, want 42", apiErr.Code)
	}
	if apiErr.Message != "invalid parameter" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "invalid parameter")
	}
}

func TestParseError_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	err := parseError(resp)
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatal("expected *APIError")
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want 500", apiErr.HTTPStatus)
	}
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Internal Server Error")
	}
}

func TestParseError_NonJSONBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("<html>Bad Gateway</html>")),
	}

	err := parseError(resp)
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatal("expected *APIError")
	}
	if apiErr.HTTPStatus != http.StatusBadGateway {
		t.Errorf("HTTPStatus = %d, want 502", apiErr.HTTPStatus)
	}
	if apiErr.Message != "<html>Bad Gateway</html>" {
		t.Errorf("Message = %q, want raw body", apiErr.Message)
	}
}

func TestAPIError_JSONMarshal(t *testing.T) {
	err := &APIError{HTTPStatus: http.StatusBadRequest, Code: 1, Message: "bad"}
	data, e := json.Marshal(err)
	if e != nil {
		t.Fatal(e)
	}
	// HTTPStatus should be omitted (json:"-")
	if strings.Contains(string(data), "HTTPStatus") {
		t.Errorf("JSON should not contain HTTPStatus, got %s", data)
	}
}
