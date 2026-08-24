package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestStage335AuthHTTPRejectsMalformedPasswordTransportBeforeService(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "register raw invalid utf8",
			path: "/api/v1/auth/register",
			body: string(append([]byte(`{"email":"investor@example.com","password":"aaaaaaaaaaaa`), append([]byte{0xff}, []byte(`","language":"en","theme":"system","timezone":"UTC"}`)...)...)),
		},
		{
			name: "login raw invalid utf8",
			path: "/api/v1/auth/login",
			body: string(append([]byte(`{"email":"investor@example.com","password":"aaaaaaaaaaaa`), append([]byte{0xff}, []byte(`"}`)...)...)),
		},
		{
			name: "register lone surrogate",
			path: "/api/v1/auth/register",
			body: `{"email":"investor@example.com","password":"aaaaaaaaaaa\uD800","language":"en","theme":"system","timezone":"UTC"}`,
		},
		{
			name: "login lone surrogate",
			path: "/api/v1/auth/login",
			body: `{"email":"investor@example.com","password":"\uD800"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
			response := authRequest(t, app, http.MethodPost, test.path, test.body, "", "")
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
			}
			requireAuthErrorCode(t, response, "VALIDATION_ERROR")
			if store.registerCalls != 0 || store.findCalls != 0 {
				t.Fatalf("malformed transport reached auth service store path: register=%d find=%d", store.registerCalls, store.findCalls)
			}
		})
	}
}

func TestStage335AuthHTTPPreservesValidSurrogatePairAndLegitimateReplacementCharacter(t *testing.T) {
	tests := []struct {
		name         string
		registerBody string
		loginBody    string
	}{
		{
			name:         "valid surrogate pair",
			registerBody: `{"email":"investor@example.com","password":"aaaaaaaaaaa\uD83D\uDE00","language":"en","theme":"system","timezone":"UTC"}`,
			loginBody:    `{"email":"investor@example.com","password":"aaaaaaaaaaa😀"}`,
		},
		{
			name:         "legitimate replacement character",
			registerBody: `{"email":"investor@example.com","password":"aaaaaaaaaaa\uFFFD","language":"en","theme":"system","timezone":"UTC"}`,
			loginBody:    `{"email":"investor@example.com","password":"aaaaaaaaaaa�"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
			registerResponse := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", test.registerBody, "", "")
			defer registerResponse.Body.Close()
			if registerResponse.StatusCode != http.StatusCreated {
				t.Fatalf("expected register status %d, got %d", http.StatusCreated, registerResponse.StatusCode)
			}
			loginResponse := authRequest(t, app, http.MethodPost, "/api/v1/auth/login", test.loginBody, "", "")
			defer loginResponse.Body.Close()
			if loginResponse.StatusCode != http.StatusOK {
				t.Fatalf("expected login status %d, got %d", http.StatusOK, loginResponse.StatusCode)
			}
		})
	}
}

func requireAuthErrorCode(t *testing.T, response *http.Response, want string) {
	t.Helper()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode auth error response: %v", err)
	}
	if !strings.EqualFold(payload.Error.Code, want) {
		t.Fatalf("expected error code %q, got %q", want, payload.Error.Code)
	}
}
