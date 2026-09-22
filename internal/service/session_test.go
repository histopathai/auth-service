package service

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/histopathai/auth-service/internal/infrastructure/storage/memory"
)

// A burst of concurrent requests on the same session (e.g. many image/annotation
// fetches fired at once) must not corrupt or spuriously expire it. Run with
// -race: before the fix, ValidateSession/ExtendSession mutated the *model.Session
// returned by Get in place with no lock, racing on ExpiresAt/LastUsedAt.
func TestValidateSession_ConcurrentUseDoesNotExpireOrRace(t *testing.T) {
	repo := memory.NewInMemorySessionRepository(0)
	logger := slog.Default()
	svc := NewSessionService(repo, AuthService{}, logger)

	sessionID, err := svc.CreateSession(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	const concurrency = 200
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			if _, err := svc.ValidateAndExtend(context.Background(), sessionID); err != nil {
				t.Errorf("session unexpectedly invalid under concurrent use: %v", err)
			}
		}()
	}
	wg.Wait()

	if _, err := svc.ValidateSession(context.Background(), sessionID); err != nil {
		t.Fatalf("session should still be valid after the burst, got: %v", err)
	}
}
