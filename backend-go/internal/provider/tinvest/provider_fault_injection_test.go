package tinvest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestCorporateActionsSlowProviderHonorsCallerDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider, err := newCorporateActionProvider(&http.Client{}, fixedClock{now: testNow}, testToken, server.URL)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = provider.CorporateActions(ctx, testQuery("SBER"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("slow dependency did not propagate caller deadline: %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("caller deadline was not honored promptly: %s", elapsed)
	}
}

type injectedReadFailureBody struct {
	read bool
}

func (body *injectedReadFailureBody) Read(p []byte) (int, error) {
	if body.read {
		return 0, errors.New("injected provider body read failure")
	}
	body.read = true
	prefix := []byte("{\\\"dividends\\\":[")
	copy(p, prefix)
	return len(prefix), nil
}

func (*injectedReadFailureBody) Close() error { return nil }

func TestCorporateActionsMidBodyReadFailureFailsClosed(t *testing.T) {
	provider, err := newCorporateActionProvider(
		&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &injectedReadFailureBody{}}, nil
		})},
		fixedClock{now: testNow},
		testToken,
		"https://example.invalid/rest",
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = provider.CorporateActions(context.Background(), testQuery("SBER"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("mid-body transport failure did not fail closed as provider unavailable: %v", err)
	}
}

func TestCorporateActionsConnectionFailureDoesNotLeakTransportDetails(t *testing.T) {
	const transportSecret = "transport-secret-must-not-leak"
	provider, err := newCorporateActionProvider(
		&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New(transportSecret)
		})},
		fixedClock{now: testNow},
		testToken,
		"https://example.invalid/rest",
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = provider.CorporateActions(context.Background(), testQuery("SBER"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("connection failure classified incorrectly: %v", err)
	}
	if err == nil || containsProviderFaultSecret(err.Error(), transportSecret) {
		t.Fatalf("raw transport detail leaked through provider error: %v", err)
	}
}

func containsProviderFaultSecret(haystack, needle string) bool {
	if needle == "" {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

var _ io.ReadCloser = (*injectedReadFailureBody)(nil)
