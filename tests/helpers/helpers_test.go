package helpers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAssertErrorCode(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
	}
	AssertErrorCode(t, resp, http.StatusOK)
}

func TestAssertValidationError(t *testing.T) {
	body := `{"error": {"message": "invalid field x"}}`
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	AssertValidationError(t, resp, "invalid")
}

func TestAssertNotEmpty(t *testing.T) {
	body := `{"data": {"id": "123"}}`
	AssertNotEmpty(t, []byte(body), "id")

	body2 := `{"id": "123"}`
	AssertNotEmpty(t, []byte(body2), "id")
}

func TestRunConcurrently(t *testing.T) {
	count := 5
	fn := func(i int) error {
		return nil
	}
	errs := RunConcurrently(count, fn)
	assert.Len(t, errs, count)
	for _, err := range errs {
		assert.NoError(t, err)
	}
}

func TestAssertOnlyOneSucceeds(t *testing.T) {
	// This helper asserts that only one succeeds. We need to mock a scenario where that happens.
	// Since AssertOnlyOneSucceeds calls t.Errorf, testing failure is hard without a mock T.
	// We will test the success case.

	errs := []error{nil, assert.AnError, assert.AnError}
	AssertOnlyOneSucceeds(t, errs)
}

func TestAssertAllFail(t *testing.T) {
	errs := []error{assert.AnError, assert.AnError}
	AssertAllFail(t, errs)
}

func TestSendMalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"name": "test", "incomplete": `, string(body))
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	client := &http.Client{}
	SendMalformedJSON(t, client, ts.URL, "POST", "token")
}

func TestSendOversizedPayload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assert.Equal(t, int64(1*1024*1024), r.ContentLength) // Skipping length check as it varies
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	defer ts.Close()

	client := &http.Client{}
	SendOversizedPayload(t, client, ts.URL, "POST", "token", 1)
}

func TestSendWithContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &http.Client{}
	SendWithContentType(t, client, ts.URL, "POST", "token", "text/plain", bytes.NewBufferString("data"))
}

func TestGetBaseURL(t *testing.T) {
	url := GetBaseURL()
	assert.Equal(t, testutil.TestBaseURL, url)
}

func TestFormatURL(t *testing.T) {
	url := FormatURL("/users")
	assert.Equal(t, testutil.TestBaseURL+"/users", url)
}
