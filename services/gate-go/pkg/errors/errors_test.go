package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"errors"

	"github.com/gin-gonic/gin"
	apperrors "github.com/rohanpatel2002/ironclad/services/gate-go/pkg/errors"
)

func TestAppError_ErrorString(t *testing.T) {
	err := apperrors.New(http.StatusBadRequest, apperrors.ErrInvalidRequest, "bad data")
	if err.Error() != "bad data" {
		t.Errorf("expected 'bad data', got %s", err.Error())
	}
}

func TestRespond_WithAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := apperrors.New(http.StatusForbidden, apperrors.ErrForbidden, "access denied")
	apperrors.Respond(c, appErr)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp apperrors.AppError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != apperrors.ErrForbidden || resp.Message != "access denied" {
		t.Errorf("unexpected response body: %+v", resp)
	}
}

func TestRespond_WithGenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	genericErr := errors.New("something blew up")
	apperrors.Respond(c, genericErr)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var resp apperrors.AppError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != apperrors.ErrInternal || resp.Message != "An unexpected error occurred" {
		t.Errorf("unexpected response body: %+v", resp)
	}
}
