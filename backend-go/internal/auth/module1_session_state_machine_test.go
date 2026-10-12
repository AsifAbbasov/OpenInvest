package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func module1StateMachineFixture(t *testing.T) (*Service, *memoryStore, string, string) {
	t.Helper()
	now := fixedClock{}.Now()
	user := StoredUser{
		ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		InvestmentSubjectID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		Email: "state-machine@example.test",
		Language: LanguageEN,
		Theme: ThemeSystem,
		Timezone: "UTC",
		PrivacyMode: true,
		CreatedAt: now,
	}
	refresh := "state-machine-refresh-seed"
	csrf := "state-machine-csrf-seed"
	record := SessionRecord{
		SessionID: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		UserID: user.ID,
		RefreshTokenHash: tokenHash(refresh),
		CSRFTokenHash: tokenHash(csrf),
		ExpiresAt: now.Add(time.Hour),
		Now: now,
	}
	store := &memoryStore{
		user: user,
		sessions: map[string]SessionRecord{record.RefreshTokenHash: record},
	}
	return newTestService(t, store), store, refresh, csrf
}

func TestModule1RefreshLogoutStateMachineDeterministic(t *testing.T) {
	service, store, refresh, csrf := module1StateMachineFixture(t)
	originalRefresh, originalCSRF := refresh, csrf

	first, err := service.Refresh(context.Background(), refresh, csrf)
	if err != nil {
		t.Fatalf("MODEL_ALLOWED refresh unexpectedly denied: %v", err)
	}
	refresh, csrf = first.RefreshToken, first.Session.CSRFToken
	if len(store.sessions) != 1 {
		t.Fatalf("expected exactly one active service-store session after rotation, got %d", len(store.sessions))
	}
	if _, err := service.Refresh(context.Background(), originalRefresh, originalCSRF); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("IMPLEMENTATION_GRANTED_WHEN_MODEL_DENIES replay err=%v", err)
	}
	if len(store.sessions) != 1 {
		t.Fatalf("replay changed current session state unexpectedly: %d", len(store.sessions))
	}
	revoked, err := service.Logout(context.Background(), refresh, csrf, false)
	if err != nil || !revoked {
		t.Fatalf("MODEL_ALLOWED logout failed revoked=%v err=%v", revoked, err)
	}
	if len(store.sessions) != 0 {
		t.Fatalf("logout left active service-store session count=%d", len(store.sessions))
	}
	if _, err := service.Refresh(context.Background(), refresh, csrf); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("IMPLEMENTATION_GRANTED_WHEN_MODEL_DENIES post-logout refresh err=%v", err)
	}
}

func FuzzModule1RefreshLogoutStateMachine(f *testing.F) {
	f.Add([]byte{0, 1, 2, 0})
	f.Add([]byte{3, 0, 4, 2})
	f.Add([]byte{0, 0, 1, 5, 2})
	f.Fuzz(func(t *testing.T, commands []byte) {
		if len(commands) > 16 {
			commands = commands[:16]
		}
		service, store, currentRefresh, currentCSRF := module1StateMachineFixture(t)
		originalRefresh, originalCSRF := currentRefresh, currentCSRF
		alive := true

		for step, raw := range commands {
			cmd := raw % 6
			switch cmd {
			case 0: // refresh current
				result, err := service.Refresh(context.Background(), currentRefresh, currentCSRF)
				if alive {
					if err != nil {
						t.Fatalf("step=%d cmd=refresh MODEL_ALLOWED implementation_denied=%v", step, err)
					}
					currentRefresh, currentCSRF = result.RefreshToken, result.Session.CSRFToken
				} else if !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("step=%d cmd=refresh IMPLEMENTATION_GRANTED_WHEN_MODEL_DENIES err=%v", step, err)
				}
			case 1: // replay original
				result, err := service.Refresh(context.Background(), originalRefresh, originalCSRF)
				originalStillCurrent := alive && currentRefresh == originalRefresh && currentCSRF == originalCSRF
				if originalStillCurrent {
					if err != nil {
						t.Fatalf("step=%d cmd=replay-as-current MODEL_ALLOWED denied=%v", step, err)
					}
					currentRefresh, currentCSRF = result.RefreshToken, result.Session.CSRFToken
				} else if !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("step=%d cmd=replay IMPLEMENTATION_GRANTED_WHEN_MODEL_DENIES err=%v", step, err)
				}
			case 2: // logout current
				revoked, err := service.Logout(context.Background(), currentRefresh, currentCSRF, false)
				if alive {
					if err != nil || !revoked {
						t.Fatalf("step=%d cmd=logout MODEL_ALLOWED revoked=%v err=%v", step, revoked, err)
					}
					alive = false
				} else if !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("step=%d cmd=logout IMPLEMENTATION_GRANTED_WHEN_MODEL_DENIES err=%v", step, err)
				}
			case 3: // wrong csrf refresh
				_, err := service.Refresh(context.Background(), currentRefresh, "wrong-csrf")
				if !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("step=%d cmd=wrong-csrf-refresh expected invalid session got=%v", step, err)
				}
			case 4: // wrong csrf logout
				_, err := service.Logout(context.Background(), currentRefresh, "wrong-csrf", false)
				if !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("step=%d cmd=wrong-csrf-logout expected invalid session got=%v", step, err)
				}
			case 5: // random invalid refresh
				_, err := service.Refresh(context.Background(), "invalid-refresh", currentCSRF)
				if !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("step=%d cmd=invalid-refresh expected invalid session got=%v", step, err)
				}
			}

			expectedSessions := 0
			if alive {
				expectedSessions = 1
			}
			if len(store.sessions) != expectedSessions {
				t.Fatalf("step=%d MODEL_STATE_MISMATCH alive=%v expected_sessions=%d actual=%d", step, alive, expectedSessions, len(store.sessions))
			}
		}
	})
}
