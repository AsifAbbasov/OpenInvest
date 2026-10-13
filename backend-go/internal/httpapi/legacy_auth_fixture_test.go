package httpapi

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"net/http"
	"sync"
	"testing"
)

var fixtureServices sync.Map

// fixtureLegacyLogin seeds a legacy account directly, then exercises real HTTP login.
// It never calls the public register endpoint or changes ownership admission.
func fixtureLegacyLogin(t *testing.T, app *fiber.App, method, path, body, refresh, csrf string) *http.Response {
	t.Helper()
	var request registerRequestDTO
	if err := decodeStrictJSON([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	value, ok := fixtureServices.Load(app)
	if !ok {
		t.Fatal("missing auth fixture service")
	}
	service := value.(*auth.Service)
	if _, err := service.Register(context.Background(), auth.RegistrationRequest{Email: request.Email, Password: string(request.Password), Language: request.Language, Theme: request.Theme, Timezone: request.Timezone}); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]string{"email": request.Email, "password": string(request.Password)})
	return authRequest(t, app, http.MethodPost, "/api/v1/auth/login", string(payload), "", "")
}
func seedLegacyFixture(t *testing.T, app *fiber.App, body string) {
	t.Helper()
	r := fixtureLegacyLogin(t, app, "", "", body, "", "")
	r.Body.Close()
}
