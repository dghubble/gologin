package gologin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultErrorHandler(t *testing.T) {
	const expectedMessage = "some error"
	rec := httptest.NewRecorder()
	// should pass through errors and codes
	DefaultErrorHandler.ServeHTTP(rec, fmt.Errorf(expectedMessage), http.StatusBadRequest)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected code %v, got %v", http.StatusBadRequest, rec.Code)
	}
	if rec.Body.String() != expectedMessage+"\n" {
		t.Errorf("expected error message %v, got %v", expectedMessage+"\n", rec.Body.String())
	}
}
