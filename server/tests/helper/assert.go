package helper

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/stretchr/testify/assert"
)

func AssertResponse(t *testing.T, wantCode int, wantRes any, w *httptest.ResponseRecorder) {

	if wantCode >= 400 || wantCode >= 500 {

		var actualRes domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &actualRes)
		assert.NoError(t, err)

		if diff := cmp.Diff(wantRes, actualRes); diff != "" {
			t.Errorf("response value is mismatch (-want +actual):\n%s", diff)
		}

	} else {

		var resp domain.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		if diff := cmp.Diff(wantRes, resp); diff != "" {

			t.Errorf("response value is mismatch (-want +actual):\n%s", diff)
			assert.Equal(t, wantRes, resp)
		}
	}
}
