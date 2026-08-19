package centreon

import (
	"encoding/json"
	"errors"
	"fmt"
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

func TestAPIError_IsRouteNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want bool
	}{
		{
			// 25.10 emits code 0 for a routing 404.
			name: "routing 404 on 25.10 (code 0)",
			err:  &APIError{HTTPStatus: http.StatusNotFound, Code: 0, Message: `No route found for "GET https://host/centreon/api/latest/x"`},
			want: true,
		},
		{
			// 24.10 emits code 404 for the same routing 404, so code cannot be
			// required in the predicate.
			name: "routing 404 on 24.10 (code 404 with prefix)",
			err:  &APIError{HTTPStatus: http.StatusNotFound, Code: 404, Message: `No route found for "GET https://host/centreon/api/latest/x"`},
			want: true,
		},
		{
			// The 25.10-only detail GET on 24.10: the path exists only for
			// DELETE/PATCH, so a GET yields a "method not allowed" dressed as a
			// routing 404. From the caller's view the endpoint is unavailable.
			name: "detail GET method not allowed on 24.10",
			err:  &APIError{HTTPStatus: http.StatusNotFound, Code: 404, Message: `No route found for "GET https://host/x": Method Not Allowed (Allow: DELETE, PATCH)`},
			want: true,
		},
		{
			name: "resource 404 (Host not found)",
			err:  &APIError{HTTPStatus: http.StatusNotFound, Code: 404, Message: "Host not found"},
			want: false,
		},
		{
			name: "empty-body 404 (code 0 but no prefix)",
			err:  &APIError{HTTPStatus: http.StatusNotFound, Code: 0, Message: "Not Found"},
			want: false,
		},
		{
			name: "404 code 0 with unrelated message",
			err:  &APIError{HTTPStatus: http.StatusNotFound, Code: 0, Message: "something else"},
			want: false,
		},
		{
			name: "prefix present but not 404",
			err:  &APIError{HTTPStatus: http.StatusInternalServerError, Code: 0, Message: "No route found for x"},
			want: false,
		},
		{
			name: "nil receiver",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsRouteNotFound(); got != tt.want {
				t.Errorf("IsRouteNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRouteNotFound_Helper(t *testing.T) {
	routing := &APIError{HTTPStatus: http.StatusNotFound, Code: 0, Message: `No route found for "GET https://host/x"`}
	resource := &APIError{HTTPStatus: http.StatusNotFound, Code: 404, Message: "Host not found"}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"routing *APIError", routing, true},
		{"resource *APIError", resource, false},
		{"wrapped routing *APIError", fmt.Errorf("call failed: %w", routing), true},
		{"plain error", errors.New("boom"), false},
		{"nil error", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRouteNotFound(tt.err); got != tt.want {
				t.Errorf("IsRouteNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseError_RouteNotFound(t *testing.T) {
	// Routing 404: body code 0 + "No route found for" -> IsRouteNotFound true.
	routing := parseError(&http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"No route found for \"GET https://host/centreon/api/latest/x\""}`)),
	})
	rErr, ok := errors.AsType[*APIError](routing)
	if !ok {
		t.Fatal("expected *APIError for routing 404")
	}
	if rErr.Code != 0 {
		t.Errorf("routing Code = %d, want 0", rErr.Code)
	}
	if !rErr.IsRouteNotFound() {
		t.Error("routing 404 should report IsRouteNotFound() == true")
	}

	// 24.10 routing 404: body code is 404 (not 0), yet the message prefix still
	// identifies it, so IsRouteNotFound must be true despite the nonzero code.
	routing2410 := parseError(&http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"code":404,"message":"No route found for \"GET https://host/x\": Method Not Allowed (Allow: DELETE, PATCH)"}`)),
	})
	r2410Err, ok := errors.AsType[*APIError](routing2410)
	if !ok {
		t.Fatal("expected *APIError for 24.10 routing 404")
	}
	if r2410Err.Code != 404 {
		t.Errorf("24.10 routing Code = %d, want 404", r2410Err.Code)
	}
	if !r2410Err.IsRouteNotFound() {
		t.Error("24.10 routing 404 (code 404 with prefix) should report IsRouteNotFound() == true")
	}

	// Resource 404: body code 404 -> IsRouteNotFound false.
	resource := parseError(&http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"code":404,"message":"Host not found"}`)),
	})
	resErr, ok := errors.AsType[*APIError](resource)
	if !ok {
		t.Fatal("expected *APIError for resource 404")
	}
	if resErr.Code != 404 {
		t.Errorf("resource Code = %d, want 404", resErr.Code)
	}
	if resErr.IsRouteNotFound() {
		t.Error("resource 404 should report IsRouteNotFound() == false")
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
