package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0338ExpiryBoundaryIsInclusive(t *testing.T) {
	boundary := time.Unix(1_800_000_000, 0).UTC()
	if !expiredAt(boundary, boundary) {
		t.Fatal("expires_at == decision time must be expired")
	}
	if expiredAt(boundary.Add(time.Nanosecond), boundary) {
		t.Fatal("expires_at after decision time must remain authoritative")
	}
}

func TestStage0338ExpiredReplayLookupIsNotFoundAndKeyCanBeReclaimed(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})

	key := "stage-03-38-reclaim-key-0001"
	path := "/api/v1/portfolios"
	request := verticalslice.CreatePortfolioRequest{Name: "Retention generation", BaseCurrency: verticalslice.RUB}
	firstContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "38383838383838383838383838383838"}
	firstBody := []byte(`{"generation":"old"}`)
	builderCalls := 0

	firstPortfolio, firstArtifact, err := service.CreatePortfolioWithReplay(
		ctx,
		firstContext,
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       append([]byte(nil), firstBody...),
				RequestID:  firstContext.RequestID,
				TraceID:    firstContext.TraceID,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("create old generation: %v", err)
	}

	var oldCommandID string
	var requestHash string
	var oldCreatedAt time.Time
	if err := store.db.QueryRowContext(ctx, `
		SELECT id, request_hash, created_at
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, subjectID, path, key).Scan(&oldCommandID, &requestHash, &oldCreatedAt); err != nil {
		t.Fatalf("query old generation: %v", err)
	}

	lookupCommand := verticalslice.CommandContext{
		SubjectID:      subjectID,
		IdempotencyKey: key,
		RequestHash:    requestHash,
		RequestPath:    path,
	}
	replayed, found, err := store.LookupReplayArtifact(ctx, lookupCommand, "POST")
	if err != nil || !found {
		t.Fatalf("lookup before expiry: found=%t err=%v", found, err)
	}
	if !bytes.Equal(replayed.Body, firstArtifact.Body) {
		t.Fatalf("pre-expiry replay changed: got=%s want=%s", replayed.Body, firstArtifact.Body)
	}

	if _, err := store.db.ExecContext(ctx, `
		UPDATE investment.command_deduplication
		SET expires_at = clock_timestamp()
		WHERE id = $1
	`, oldCommandID); err != nil {
		t.Fatalf("expire old generation: %v", err)
	}

	if _, found, err := store.LookupReplayArtifact(ctx, lookupCommand, "POST"); err != nil || found {
		t.Fatalf("expired read-only replay must be unavailable: found=%t err=%v", found, err)
	}

	secondContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "39393939393939393939393939393939"}
	secondBody := []byte(`{"generation":"new"}`)
	secondPortfolio, secondArtifact, err := service.CreatePortfolioWithReplay(
		ctx,
		secondContext,
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       append([]byte(nil), secondBody...),
				RequestID:  secondContext.RequestID,
				TraceID:    secondContext.TraceID,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("create new generation: %v", err)
	}
	if firstPortfolio.ID == secondPortfolio.ID {
		t.Fatalf("fresh generation reused old business identity %s", firstPortfolio.ID)
	}
	if builderCalls != 2 {
		t.Fatalf("expected one builder call per admitted generation, got %d", builderCalls)
	}
	if !bytes.Equal(secondArtifact.Body, secondBody) {
		t.Fatalf("new generation replay body mismatch: %s", secondArtifact.Body)
	}

	var newCommandID string
	var newCreatedAt time.Time
	var expirySeconds int64
	var terminalStatus sql.NullString
	if err := store.db.QueryRowContext(ctx, `
		SELECT id, created_at, EXTRACT(EPOCH FROM (expires_at - created_at))::bigint, terminal_status
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, subjectID, path, key).Scan(&newCommandID, &newCreatedAt, &expirySeconds, &terminalStatus); err != nil {
		t.Fatalf("query new generation: %v", err)
	}
	if newCommandID == oldCommandID {
		t.Fatalf("fresh generation retained old command identity %s", oldCommandID)
	}
	if !newCreatedAt.After(oldCreatedAt) {
		t.Fatalf("fresh generation created_at was not advanced: old=%s new=%s", oldCreatedAt, newCreatedAt)
	}
	if expirySeconds != 24*60*60 {
		t.Fatalf("fresh generation expiry interval = %d seconds, want 86400", expirySeconds)
	}
	if !terminalStatus.Valid || terminalStatus.String != "success" {
		t.Fatalf("fresh generation did not complete successfully: %+v", terminalStatus)
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 2, subjectID)
	assertStage0332Count(t, store, `
		SELECT count(*)
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, 1, subjectID, path, key)
}

func TestStage0338ConcurrentPostExpiryRetryCreatesOneNewBusinessEffect(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})

	key := "stage-03-38-post-expiry-concurrent-key"
	path := "/api/v1/portfolios"
	request := verticalslice.CreatePortfolioRequest{Name: "Concurrent generation", BaseCurrency: verticalslice.RUB}
	firstContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "40404040404040404040404040404040"}
	if _, _, err := service.CreatePortfolioWithReplay(
		ctx, firstContext, subjectID, key, path, request,
		stage0332ArtifactBuilder(firstContext, []byte(`{"generation":"old"}`)),
	); err != nil {
		t.Fatalf("seed old generation: %v", err)
	}

	if _, err := store.db.ExecContext(ctx, `
		UPDATE investment.command_deduplication
		SET expires_at = clock_timestamp() - interval '1 microsecond'
		WHERE principal_id = $1 AND idempotency_key = $2
	`, subjectID, key); err != nil {
		t.Fatalf("expire seed generation: %v", err)
	}

	type outcome struct {
		artifact verticalslice.CommandReplayArtifact
		err      error
	}
	results := make(chan outcome, 2)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)
	var builderCalls atomic.Int32
	traceIDs := [2]string{
		"41414141414141414141414141414141",
		"42424242424242424242424242424242",
	}

	for i := 0; i < 2; i++ {
		index := i
		go func() {
			ready.Done()
			<-start
			requestContext := verticalslice.RequestContext{
				RequestID: uuid.NewString(),
				TraceID:   traceIDs[index],
			}
			_, artifact, err := service.CreatePortfolioWithReplay(
				ctx, requestContext, subjectID, key, path, request,
				func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
					builderCalls.Add(1)
					return verticalslice.CommandReplayArtifact{
						StatusCode: 201,
						Body:       []byte(`{"generation":"post-expiry"}`),
						RequestID:  requestContext.RequestID,
						TraceID:    requestContext.TraceID,
					}, nil
				},
			)
			results <- outcome{artifact: artifact, err: err}
		}()
	}
	ready.Wait()
	close(start)

	var artifacts []verticalslice.CommandReplayArtifact
	for i := 0; i < 2; i++ {
		result := <-results
		if result.err != nil {
			t.Fatalf("post-expiry concurrent request %d: %v", i+1, result.err)
		}
		artifacts = append(artifacts, result.artifact)
	}
	if builderCalls.Load() != 1 {
		t.Fatalf("post-expiry business response builder ran %d times, want 1", builderCalls.Load())
	}
	if !bytes.Equal(artifacts[0].Body, artifacts[1].Body) ||
		artifacts[0].RequestID != artifacts[1].RequestID ||
		artifacts[0].TraceID != artifacts[1].TraceID {
		t.Fatalf("concurrent post-expiry callers did not converge on one artifact: %+v %+v", artifacts[0], artifacts[1])
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 2, subjectID)
}

func TestStage0338CommandRaceStraddlingExpiryCannotMixGenerations(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})

	key := "stage-03-38-straddle-command-key"
	path := "/api/v1/portfolios"
	request := verticalslice.CreatePortfolioRequest{Name: "Straddle command", BaseCurrency: verticalslice.RUB}
	oldContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "43434343434343434343434343434343"}
	oldArtifactBody := []byte(`{"generation":"old-artifact"}`)
	if _, _, err := service.CreatePortfolioWithReplay(
		ctx, oldContext, subjectID, key, path, request,
		stage0332ArtifactBuilder(oldContext, oldArtifactBody),
	); err != nil {
		t.Fatalf("seed old generation: %v", err)
	}

	var oldID string
	var oldHash string
	var expiresAt time.Time
	if err := store.db.QueryRowContext(ctx, `
		UPDATE investment.command_deduplication
		SET expires_at = clock_timestamp() + interval '2 seconds'
		WHERE principal_id = $1 AND idempotency_key = $2
		RETURNING id, request_hash, expires_at
	`, subjectID, key).Scan(&oldID, &oldHash, &expiresAt); err != nil {
		t.Fatalf("prepare old generation expiry: %v", err)
	}

	oldCommand := verticalslice.CommandContext{
		SubjectID:      subjectID,
		IdempotencyKey: key,
		RequestHash:    oldHash,
		RequestPath:    path,
	}
	blocker, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin blocker tx: %v", err)
	}
	if err := lockReplayCommandScope(ctx, blocker, oldCommand, "POST"); err != nil {
		_ = blocker.Rollback()
		t.Fatalf("lock exact replay scope: %v", err)
	}

	type delayedResult struct {
		reservation replayReservation
		err         error
	}
	delayed := make(chan delayedResult, 1)
	startedBeforeExpiry := time.Now().UTC()
	go func() {
		tx, beginErr := store.db.BeginTx(ctx, &sql.TxOptions{})
		if beginErr != nil {
			delayed <- delayedResult{err: beginErr}
			return
		}
		defer rollback(tx)
		reservation, reserveErr := reserveReplayCommand(ctx, tx, oldCommand, "POST")
		if reserveErr == nil {
			reserveErr = tx.Commit()
		}
		delayed <- delayedResult{reservation: reservation, err: reserveErr}
	}()

	waitForStage0338AdvisoryWaiter(t, store)
	if !startedBeforeExpiry.Before(expiresAt) {
		_ = blocker.Rollback()
		t.Fatalf("delayed request did not start before expiry: start=%s expiry=%s", startedBeforeExpiry, expiresAt)
	}
	waitForStage0338DatabaseTime(t, store, expiresAt)

	newCommand := oldCommand
	newCommand.RequestHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	newReservation, err := reserveReplayCommand(ctx, blocker, newCommand, "POST")
	if err != nil {
		_ = blocker.Rollback()
		t.Fatalf("establish new generation: %v", err)
	}
	newArtifact := verticalslice.CommandReplayArtifact{
		StatusCode: 201,
		Body:       []byte(`{"generation":"new-artifact"}`),
		RequestID:  uuid.NewString(),
		TraceID:    "44444444444444444444444444444444",
	}
	if err := completeReplayCommand(ctx, blocker, newReservation.ID, newArtifact); err != nil {
		_ = blocker.Rollback()
		t.Fatalf("complete new generation: %v", err)
	}
	if err := blocker.Commit(); err != nil {
		t.Fatalf("commit new generation: %v", err)
	}

	result := <-delayed
	if !errors.Is(result.err, ErrIdempotencyConflict) {
		t.Fatalf("delayed pre-expiry request must observe new generation conflict, got reservation=%+v err=%v", result.reservation, result.err)
	}

	var currentID string
	var currentHash string
	var currentBody []byte
	if err := store.db.QueryRowContext(ctx, `
		SELECT id, request_hash, response_body
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, subjectID, path, key).Scan(&currentID, &currentHash, &currentBody); err != nil {
		t.Fatalf("query winning generation: %v", err)
	}
	if currentID == oldID || currentID != newReservation.ID {
		t.Fatalf("winning generation identity mismatch: old=%s new=%s current=%s", oldID, newReservation.ID, currentID)
	}
	if currentHash != newCommand.RequestHash {
		t.Fatalf("winning generation hash mixed: got=%s want=%s", currentHash, newCommand.RequestHash)
	}
	if !bytes.Equal(currentBody, newArtifact.Body) || bytes.Equal(currentBody, oldArtifactBody) {
		t.Fatalf("winning generation artifact mixed: %s", currentBody)
	}
}

func TestStage0338RefreshBlockedAcrossExpiryHasNoContainmentAuthority(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	fixture := seedStage0338RevokedSessionFamily(t, store, 2*time.Second)

	blocker := lockStage0338UserRefreshes(t, store, fixture.userID)
	staleNow := fixture.expiresAt.Add(-time.Second)
	next := auth.SessionRecord{
		SessionID:        uuid.NewString(),
		RefreshTokenHash: "stage338-next-refresh-" + uuid.NewString(),
		CSRFTokenHash:    "stage338-next-csrf-" + uuid.NewString(),
		ExpiresAt:        fixture.expiresAt.Add(30 * 24 * time.Hour),
		Now:              staleNow,
	}

	type result struct {
		err error
	}
	outcome := make(chan result, 1)
	go func() {
		_, err := store.RotateSession(ctx, fixture.refreshHash, fixture.csrfHash, next, staleNow)
		outcome <- result{err: err}
	}()

	waitForStage0338AdvisoryWaiter(t, store)
	waitForStage0338DatabaseTime(t, store, fixture.expiresAt)
	if err := blocker.Commit(); err != nil {
		t.Fatalf("release refresh blocker: %v", err)
	}

	got := <-outcome
	if !errors.Is(got.err, auth.ErrInvalidSession) {
		t.Fatalf("post-expiry refresh must be rejected, got %v", got.err)
	}
	assertStage0338SessionState(t, store, fixture.descendantID, "active")
	assertStage0332Count(t, store, `SELECT count(*) FROM identity.sessions WHERE id = $1`, 0, next.SessionID)
}

func TestStage0338LogoutBlockedAcrossExpiryHasNoUserRevocationAuthority(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	fixture := seedStage0338RevokedSessionFamily(t, store, 2*time.Second)

	blocker := lockStage0338UserRefreshes(t, store, fixture.userID)
	staleNow := fixture.expiresAt.Add(-time.Second)

	type result struct {
		revoked bool
		err     error
	}
	outcome := make(chan result, 1)
	go func() {
		revoked, err := store.RevokeSession(ctx, fixture.refreshHash, fixture.csrfHash, true, staleNow)
		outcome <- result{revoked: revoked, err: err}
	}()

	waitForStage0338AdvisoryWaiter(t, store)
	waitForStage0338DatabaseTime(t, store, fixture.expiresAt)
	if err := blocker.Commit(); err != nil {
		t.Fatalf("release logout blocker: %v", err)
	}

	got := <-outcome
	if got.revoked || !errors.Is(got.err, auth.ErrInvalidSession) {
		t.Fatalf("post-expiry logout must have zero revocation authority: revoked=%t err=%v", got.revoked, got.err)
	}
	assertStage0338SessionState(t, store, fixture.descendantID, "active")
}

func TestStage0338BoundedCleanupAndIndexes(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	principalID := uuid.NewString()
	for i := 0; i < operationalRetentionCleanupBatchSize+1; i++ {
		if _, err := store.db.ExecContext(ctx, `
			INSERT INTO investment.command_deduplication (
				id, principal_id, method, canonical_path, idempotency_key, request_hash,
				created_at, expires_at
			)
			VALUES ($1, $2, 'POST', '/stage338/cleanup', $3, $4,
				clock_timestamp() - interval '49 hours',
				clock_timestamp() - interval '25 hours')
		`, uuid.NewString(), principalID, "cleanup-key-"+uuid.NewString(), "hash-"+uuid.NewString()); err != nil {
			t.Fatalf("seed expired command %d: %v", i, err)
		}
	}
	unexpiredCommandID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO investment.command_deduplication (
			id, principal_id, method, canonical_path, idempotency_key, request_hash,
			created_at, expires_at
		)
		VALUES ($1, $2, 'POST', '/stage338/cleanup', $3, $4,
			clock_timestamp(), clock_timestamp() + interval '24 hours')
	`, unexpiredCommandID, principalID, "cleanup-live-"+uuid.NewString(), "hash-"+uuid.NewString()); err != nil {
		t.Fatalf("seed unexpired command: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM investment.command_deduplication WHERE principal_id = $1`, principalID)
	})

	firstDeleted := runStage0338CommandCleanup(t, store)
	if firstDeleted != operationalRetentionCleanupBatchSize {
		t.Fatalf("first command cleanup deleted %d, want %d", firstDeleted, operationalRetentionCleanupBatchSize)
	}
	assertStage0332Count(t, store, `
		SELECT count(*) FROM investment.command_deduplication
		WHERE principal_id = $1 AND expires_at <= clock_timestamp()
	`, 1, principalID)
	secondDeleted := runStage0338CommandCleanup(t, store)
	if secondDeleted != 1 {
		t.Fatalf("second command cleanup deleted %d, want 1", secondDeleted)
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.command_deduplication WHERE id = $1`, 1, unexpiredCommandID)

	userID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.users (id, email_normalized)
		VALUES ($1, $2)
	`, userID, "stage338-cleanup-"+uuid.NewString()+"@example.com"); err != nil {
		t.Fatalf("seed cleanup user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM identity.users WHERE id = $1`, userID)
	})
	for i := 0; i < operationalRetentionCleanupBatchSize+1; i++ {
		sessionID := uuid.NewString()
		if _, err := store.db.ExecContext(ctx, `
			INSERT INTO identity.sessions (
				id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
				session_state, expires_at, created_at, last_rotated_at
			)
			VALUES ($1, $2, $1, $3, $4, 'active',
				clock_timestamp() - interval '25 hours',
				clock_timestamp() - interval '49 hours',
				clock_timestamp() - interval '49 hours')
		`, sessionID, userID, "cleanup-refresh-"+uuid.NewString(), "cleanup-csrf-"+uuid.NewString()); err != nil {
			t.Fatalf("seed expired session %d: %v", i, err)
		}
	}
	unexpiredSessionID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at
		)
		VALUES ($1, $2, $1, $3, $4, 'active',
			clock_timestamp() + interval '24 hours',
			clock_timestamp(), clock_timestamp())
	`, unexpiredSessionID, userID, "cleanup-refresh-"+uuid.NewString(), "cleanup-csrf-"+uuid.NewString()); err != nil {
		t.Fatalf("seed unexpired session: %v", err)
	}

	firstSessionDeleted := runStage0338SessionCleanup(t, store)
	if firstSessionDeleted != operationalRetentionCleanupBatchSize {
		t.Fatalf("first session cleanup deleted %d, want %d", firstSessionDeleted, operationalRetentionCleanupBatchSize)
	}
	assertStage0332Count(t, store, `
		SELECT count(*) FROM identity.sessions
		WHERE user_id = $1 AND expires_at <= clock_timestamp()
	`, 1, userID)
	secondSessionDeleted := runStage0338SessionCleanup(t, store)
	if secondSessionDeleted != 1 {
		t.Fatalf("second session cleanup deleted %d, want 1", secondSessionDeleted)
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM identity.sessions WHERE id = $1`, 1, unexpiredSessionID)

	assertStage0332Count(t, store, `
		SELECT count(*) FROM pg_indexes
		WHERE schemaname = 'investment'
			AND tablename = 'command_deduplication'
			AND indexname = 'command_deduplication_expires_id_idx'
	`, 1)
	assertStage0332Count(t, store, `
		SELECT count(*) FROM pg_indexes
		WHERE schemaname = 'identity'
			AND tablename = 'sessions'
			AND indexname = 'sessions_expires_id_idx'
	`, 1)
}

type stage0338SessionFixture struct {
	userID       string
	refreshHash  string
	csrfHash     string
	descendantID string
	expiresAt    time.Time
}

func seedStage0338RevokedSessionFamily(t *testing.T, store *Store, ttl time.Duration) stage0338SessionFixture {
	t.Helper()
	ctx := context.Background()
	userID := uuid.NewString()
	email := "stage338-expiry-" + uuid.NewString() + "@example.com"
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.users (id, email_normalized)
		VALUES ($1, $2)
	`, userID, email); err != nil {
		t.Fatalf("insert auth expiry user: %v", err)
	}
	rootID := uuid.NewString()
	descendantID := uuid.NewString()
	refreshHash := "stage338-root-refresh-" + uuid.NewString()
	csrfHash := "stage338-root-csrf-" + uuid.NewString()
	var expiresAt time.Time
	if err := store.db.QueryRowContext(ctx, `
		SELECT clock_timestamp() + ($1::bigint * interval '1 millisecond')
	`, ttl.Milliseconds()).Scan(&expiresAt); err != nil {
		t.Fatalf("compute root expiry: %v", err)
	}
	createdAt := expiresAt.Add(-time.Hour)
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at, revoked_at
		)
		VALUES ($1, $2, $1, $3, $4, 'revoked', $5, $6, $6, $6)
	`, rootID, userID, refreshHash, csrfHash, expiresAt, createdAt); err != nil {
		t.Fatalf("insert revoked root: %v", err)
	}
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at
		)
		VALUES ($1, $2, $3, $4, $5, 'active',
			clock_timestamp() + interval '1 day',
			clock_timestamp(), clock_timestamp())
	`, descendantID, userID, rootID, "stage338-desc-refresh-"+uuid.NewString(), "stage338-desc-csrf-"+uuid.NewString()); err != nil {
		t.Fatalf("insert active descendant: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.events WHERE actor_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.actors WHERE id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.users WHERE id = $1`, userID)
	})
	return stage0338SessionFixture{
		userID:       userID,
		refreshHash:  refreshHash,
		csrfHash:     csrfHash,
		descendantID: descendantID,
		expiresAt:    expiresAt.UTC(),
	}
}

func lockStage0338UserRefreshes(t *testing.T, store *Store, userID string) *sql.Tx {
	t.Helper()
	tx, err := store.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin auth blocker: %v", err)
	}
	if err := lockUserRefreshes(context.Background(), tx, userID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("lock auth serialization: %v", err)
	}
	return tx
}

func waitForStage0338AdvisoryWaiter(t *testing.T, store *Store) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		if err := store.db.QueryRowContext(context.Background(), `
			SELECT count(*)
			FROM pg_locks
			WHERE locktype = 'advisory' AND NOT granted
		`).Scan(&waiting); err != nil {
			t.Fatalf("query advisory waiters: %v", err)
		}
		if waiting > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for advisory-lock waiter")
}

func waitForStage0338DatabaseTime(t *testing.T, store *Store, target time.Time) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var reached bool
		if err := store.db.QueryRowContext(context.Background(), `SELECT clock_timestamp() >= $1`, target).Scan(&reached); err != nil {
			t.Fatalf("query database time: %v", err)
		}
		if reached {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("database clock did not reach %s", target)
}

func assertStage0338SessionState(t *testing.T, store *Store, sessionID string, want string) {
	t.Helper()
	var got string
	if err := store.db.QueryRowContext(context.Background(), `
		SELECT session_state FROM identity.sessions WHERE id = $1
	`, sessionID).Scan(&got); err != nil {
		t.Fatalf("query session %s: %v", sessionID, err)
	}
	if got != want {
		t.Fatalf("session %s state = %s, want %s", sessionID, got, want)
	}
}

func drainStage0338ExpiredRows(t *testing.T, store *Store) {
	t.Helper()
	for {
		tx, err := store.db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			t.Fatalf("begin command drain: %v", err)
		}
		deleted, err := cleanupExpiredReplayCommands(context.Background(), tx)
		if err != nil {
			_ = tx.Rollback()
			t.Fatalf("drain command rows: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit command drain: %v", err)
		}
		if deleted == 0 {
			break
		}
	}
	for {
		tx, err := store.db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			t.Fatalf("begin session drain: %v", err)
		}
		deleted, err := cleanupExpiredSessions(context.Background(), tx)
		if err != nil {
			_ = tx.Rollback()
			t.Fatalf("drain session rows: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit session drain: %v", err)
		}
		if deleted == 0 {
			break
		}
	}
}

func runStage0338CommandCleanup(t *testing.T, store *Store) int64 {
	t.Helper()
	tx, err := store.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin command cleanup: %v", err)
	}
	deleted, err := cleanupExpiredReplayCommands(context.Background(), tx)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("command cleanup: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit command cleanup: %v", err)
	}
	return deleted
}

func runStage0338SessionCleanup(t *testing.T, store *Store) int64 {
	t.Helper()
	tx, err := store.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin session cleanup: %v", err)
	}
	deleted, err := cleanupExpiredSessions(context.Background(), tx)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("session cleanup: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit session cleanup: %v", err)
	}
	return deleted
}

func TestStage0338ConcurrentCleanupPassesSkipLocked(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	principalID := uuid.NewString()
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(),
			`DELETE FROM investment.command_deduplication WHERE principal_id = $1`, principalID)
	})
	for i := 0; i < 2*operationalRetentionCleanupBatchSize; i++ {
		if _, err := store.db.ExecContext(ctx, `
			INSERT INTO investment.command_deduplication (
				id, principal_id, method, canonical_path, idempotency_key, request_hash,
				created_at, expires_at
			)
			VALUES ($1, $2, 'POST', '/stage338/concurrent-cleanup', $3, $4,
				clock_timestamp() - interval '49 hours',
				clock_timestamp() - interval '25 hours')
		`, uuid.NewString(), principalID, "concurrent-cleanup-key-"+uuid.NewString(), "hash-"+uuid.NewString()); err != nil {
			t.Fatalf("seed concurrent expired command %d: %v", i, err)
		}
	}

	tx1, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin first command cleanup: %v", err)
	}
	deleted1, err := cleanupExpiredReplayCommands(ctx, tx1)
	if err != nil {
		_ = tx1.Rollback()
		t.Fatalf("first command cleanup: %v", err)
	}
	if deleted1 != operationalRetentionCleanupBatchSize {
		_ = tx1.Rollback()
		t.Fatalf("first command cleanup deleted %d, want %d", deleted1, operationalRetentionCleanupBatchSize)
	}

	tx2, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		_ = tx1.Rollback()
		t.Fatalf("begin second command cleanup: %v", err)
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
	deleted2, err := cleanupExpiredReplayCommands(timeoutCtx, tx2)
	cancel()
	if err != nil {
		_ = tx2.Rollback()
		_ = tx1.Rollback()
		t.Fatalf("second command cleanup blocked instead of SKIP LOCKED: %v", err)
	}
	if deleted2 != operationalRetentionCleanupBatchSize {
		_ = tx2.Rollback()
		_ = tx1.Rollback()
		t.Fatalf("second command cleanup deleted %d, want %d", deleted2, operationalRetentionCleanupBatchSize)
	}
	if err := tx2.Commit(); err != nil {
		_ = tx1.Rollback()
		t.Fatalf("commit second command cleanup: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("commit first command cleanup: %v", err)
	}
	assertStage0332Count(t, store, `
		SELECT count(*) FROM investment.command_deduplication
		WHERE principal_id = $1 AND expires_at <= clock_timestamp()
	`, 0, principalID)

	userID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.users (id, email_normalized)
		VALUES ($1, $2)
	`, userID, "stage338-concurrent-cleanup-"+uuid.NewString()+"@example.com"); err != nil {
		t.Fatalf("seed concurrent cleanup user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM identity.users WHERE id = $1`, userID)
	})
	for i := 0; i < 2*operationalRetentionCleanupBatchSize; i++ {
		sessionID := uuid.NewString()
		if _, err := store.db.ExecContext(ctx, `
			INSERT INTO identity.sessions (
				id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
				session_state, expires_at, created_at, last_rotated_at
			)
			VALUES ($1, $2, $1, $3, $4, 'active',
				clock_timestamp() - interval '25 hours',
				clock_timestamp() - interval '49 hours',
				clock_timestamp() - interval '49 hours')
		`, sessionID, userID, "concurrent-refresh-"+uuid.NewString(), "concurrent-csrf-"+uuid.NewString()); err != nil {
			t.Fatalf("seed concurrent expired session %d: %v", i, err)
		}
	}

	sessionTx1, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin first session cleanup: %v", err)
	}
	sessionDeleted1, err := cleanupExpiredSessions(ctx, sessionTx1)
	if err != nil {
		_ = sessionTx1.Rollback()
		t.Fatalf("first session cleanup: %v", err)
	}
	if sessionDeleted1 != operationalRetentionCleanupBatchSize {
		_ = sessionTx1.Rollback()
		t.Fatalf("first session cleanup deleted %d, want %d", sessionDeleted1, operationalRetentionCleanupBatchSize)
	}

	sessionTx2, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		_ = sessionTx1.Rollback()
		t.Fatalf("begin second session cleanup: %v", err)
	}
	timeoutCtx, cancel = context.WithTimeout(ctx, time.Second)
	sessionDeleted2, err := cleanupExpiredSessions(timeoutCtx, sessionTx2)
	cancel()
	if err != nil {
		_ = sessionTx2.Rollback()
		_ = sessionTx1.Rollback()
		t.Fatalf("second session cleanup blocked instead of SKIP LOCKED: %v", err)
	}
	if sessionDeleted2 != operationalRetentionCleanupBatchSize {
		_ = sessionTx2.Rollback()
		_ = sessionTx1.Rollback()
		t.Fatalf("second session cleanup deleted %d, want %d", sessionDeleted2, operationalRetentionCleanupBatchSize)
	}
	if err := sessionTx2.Commit(); err != nil {
		_ = sessionTx1.Rollback()
		t.Fatalf("commit second session cleanup: %v", err)
	}
	if err := sessionTx1.Commit(); err != nil {
		t.Fatalf("commit first session cleanup: %v", err)
	}
	assertStage0332Count(t, store, `
		SELECT count(*) FROM identity.sessions
		WHERE user_id = $1 AND expires_at <= clock_timestamp()
	`, 0, userID)
}

func TestStage0338RuntimeRoleCleanupNeedsNoPrivilegeExpansionAndPreservesAudit(t *testing.T) {
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if runtimeURL == "" {
		t.Skip("OPENINVEST_DATABASE_RUNTIME_TEST_URL is not set")
	}
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	principalID := uuid.NewString()
	commandID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO investment.command_deduplication (
			id, principal_id, method, canonical_path, idempotency_key, request_hash,
			created_at, expires_at
		)
		VALUES ($1, $2, 'POST', '/stage338/runtime-role', $3, $4,
			clock_timestamp() - interval '49 hours',
			clock_timestamp() - interval '25 hours')
	`, commandID, principalID, "runtime-role-key-"+uuid.NewString(), "hash-"+uuid.NewString()); err != nil {
		t.Fatalf("seed runtime-role command: %v", err)
	}

	userID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.users (id, email_normalized)
		VALUES ($1, $2)
	`, userID, "stage338-runtime-role-"+uuid.NewString()+"@example.com"); err != nil {
		t.Fatalf("seed runtime-role user: %v", err)
	}
	sessionID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at
		)
		VALUES ($1, $2, $1, $3, $4, 'active',
			clock_timestamp() - interval '25 hours',
			clock_timestamp() - interval '49 hours',
			clock_timestamp() - interval '49 hours')
	`, sessionID, userID, "runtime-role-refresh-"+uuid.NewString(), "runtime-role-csrf-"+uuid.NewString()); err != nil {
		t.Fatalf("seed runtime-role session: %v", err)
	}

	auditID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO audit.events (
			id, actor_id, action_code, target_kind, target_id, outcome,
			request_id, trace_id, occurred_at, schema_version
		)
		VALUES ($1, NULL, 'STAGE_03_38_RETENTION_PROOF', 'session', NULL, 'success',
			NULL, NULL, clock_timestamp(), 1)
	`, auditID); err != nil {
		t.Fatalf("seed audit evidence: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.events WHERE id = $1`, auditID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.users WHERE id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM investment.command_deduplication WHERE principal_id = $1`, principalID)
	})

	runtimeStore, err := Open(runtimeURL)
	if err != nil {
		t.Fatalf("open runtime-role store: %v", err)
	}
	defer runtimeStore.Close()

	var commandDelete, sessionDelete, ledgerDelete, auditDelete bool
	if err := runtimeStore.db.QueryRowContext(ctx, `
		SELECT
			has_table_privilege(current_user, 'investment.command_deduplication', 'DELETE'),
			has_table_privilege(current_user, 'identity.sessions', 'DELETE'),
			has_table_privilege(current_user, 'investment.transaction_entries', 'DELETE'),
			has_table_privilege(current_user, 'audit.events', 'DELETE')
	`).Scan(&commandDelete, &sessionDelete, &ledgerDelete, &auditDelete); err != nil {
		t.Fatalf("query runtime privileges: %v", err)
	}
	if !commandDelete || !sessionDelete {
		t.Fatalf("runtime role lacks retention cleanup DELETE: command=%t session=%t", commandDelete, sessionDelete)
	}
	if ledgerDelete || auditDelete {
		t.Fatalf("runtime role gained forbidden DELETE privilege: ledger=%t audit=%t", ledgerDelete, auditDelete)
	}

	tx, err := runtimeStore.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin runtime-role cleanup: %v", err)
	}
	commandDeleted, err := cleanupExpiredReplayCommands(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("runtime-role command cleanup: %v", err)
	}
	sessionDeleted, err := cleanupExpiredSessions(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("runtime-role session cleanup: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit runtime-role cleanup: %v", err)
	}
	if commandDeleted < 1 || sessionDeleted < 1 {
		t.Fatalf("runtime-role cleanup made insufficient progress: commands=%d sessions=%d", commandDeleted, sessionDeleted)
	}

	assertStage0332Count(t, store, `SELECT count(*) FROM investment.command_deduplication WHERE id = $1`, 0, commandID)
	assertStage0332Count(t, store, `SELECT count(*) FROM identity.sessions WHERE id = $1`, 0, sessionID)
	assertStage0332Count(t, store, `SELECT count(*) FROM audit.events WHERE id = $1`, 1, auditID)
}

func TestStage0338ExpiredExistingCommandReclamationClearsOldGenerationDeterministically(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	path := "/api/v1/portfolios"
	key := "stage-03-38-forced-reclaim-key"
	request := verticalslice.CreatePortfolioRequest{Name: "Forced reclaim", BaseCurrency: verticalslice.RUB}
	oldContext := verticalslice.RequestContext{
		RequestID: uuid.NewString(),
		TraceID:   "51515151515151515151515151515151",
	}
	oldArtifact := []byte(`{"generation":"old-forced-reclaim"}`)
	if _, _, err := service.CreatePortfolioWithReplay(
		ctx,
		oldContext,
		subjectID,
		key,
		path,
		request,
		stage0332ArtifactBuilder(oldContext, oldArtifact),
	); err != nil {
		t.Fatalf("seed completed old generation: %v", err)
	}

	var oldID string
	var oldCreatedAt time.Time
	if err := store.db.QueryRowContext(ctx, `
		UPDATE investment.command_deduplication
		SET expires_at = clock_timestamp() - interval '1 microsecond'
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
		RETURNING id, created_at
	`, subjectID, path, key).Scan(&oldID, &oldCreatedAt); err != nil {
		t.Fatalf("expire old generation: %v", err)
	}

	var beforeAdmission time.Time
	if err := store.db.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&beforeAdmission); err != nil {
		t.Fatalf("sample pre-admission database time: %v", err)
	}

	newHash := strings.Repeat("b", 64)
	command := verticalslice.CommandContext{
		SubjectID:      subjectID,
		IdempotencyKey: key,
		RequestHash:    newHash,
		RequestPath:    path,
	}
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin forced reclaim tx: %v", err)
	}
	defer rollback(tx)

	reservation, err := reserveReplayCommand(ctx, tx, command, "POST")
	if err != nil {
		t.Fatalf("reserve forced replacement generation: %v", err)
	}
	if reservation.Duplicate {
		t.Fatal("expired existing row was incorrectly returned as a duplicate")
	}
	if reservation.ID == oldID {
		t.Fatalf("reclamation reused old command UUID %s", oldID)
	}

	var (
		gotID             string
		gotHash           string
		terminalStatus    sql.NullString
		responseHash      sql.NullString
		responseVersion   sql.NullInt64
		responseStatus    sql.NullInt64
		responseBody      []byte
		responseRequestID sql.NullString
		responseTraceID   sql.NullString
		createdAt         time.Time
		expiresAt         time.Time
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			request_hash,
			terminal_status,
			response_hash,
			response_version,
			response_status,
			response_body,
			response_request_id::text,
			response_trace_id,
			created_at,
			expires_at
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, subjectID, path, key).Scan(
		&gotID,
		&gotHash,
		&terminalStatus,
		&responseHash,
		&responseVersion,
		&responseStatus,
		&responseBody,
		&responseRequestID,
		&responseTraceID,
		&createdAt,
		&expiresAt,
	); err != nil {
		t.Fatalf("inspect reclaimed generation before completion: %v", err)
	}

	var afterAdmission time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&afterAdmission); err != nil {
		t.Fatalf("sample post-admission database time: %v", err)
	}

	if gotID != reservation.ID || gotID == oldID {
		t.Fatalf("reclaimed identity mismatch: old=%s reservation=%s stored=%s", oldID, reservation.ID, gotID)
	}
	if gotHash != newHash {
		t.Fatalf("reclaimed request hash = %s, want %s", gotHash, newHash)
	}
	if terminalStatus.Valid || responseHash.Valid || responseVersion.Valid || responseStatus.Valid ||
		responseBody != nil || responseRequestID.Valid || responseTraceID.Valid {
		t.Fatalf(
			"old replay material leaked into fresh generation: terminal=%+v responseHash=%+v version=%+v status=%+v body=%v requestID=%+v traceID=%+v",
			terminalStatus,
			responseHash,
			responseVersion,
			responseStatus,
			responseBody,
			responseRequestID,
			responseTraceID,
		)
	}
	if !createdAt.After(oldCreatedAt) {
		t.Fatalf("fresh admission timestamp did not advance: old=%s new=%s", oldCreatedAt, createdAt)
	}
	if createdAt.Before(beforeAdmission) || createdAt.After(afterAdmission) {
		t.Fatalf("created_at is outside authoritative database admission window: before=%s created=%s after=%s", beforeAdmission, createdAt, afterAdmission)
	}
	if expiresAt.Sub(createdAt) != 24*time.Hour {
		t.Fatalf("fresh generation retention = %s, want 24h", expiresAt.Sub(createdAt))
	}

	newArtifact := verticalslice.CommandReplayArtifact{
		StatusCode: 201,
		Body:       []byte(`{"generation":"new-forced-reclaim"}`),
		RequestID:  uuid.NewString(),
		TraceID:    "52525252525252525252525252525252",
	}
	if err := completeReplayCommand(ctx, tx, reservation.ID, newArtifact); err != nil {
		t.Fatalf("complete reclaimed generation: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit reclaimed generation: %v", err)
	}

	replayed, found, err := store.LookupReplayArtifact(ctx, command, "POST")
	if err != nil || !found {
		t.Fatalf("lookup completed reclaimed generation: found=%t err=%v", found, err)
	}
	if !bytes.Equal(replayed.Body, newArtifact.Body) ||
		replayed.RequestID != newArtifact.RequestID ||
		replayed.TraceID != newArtifact.TraceID {
		t.Fatalf("completed reclaimed generation did not replay exactly: got=%+v want=%+v", replayed, newArtifact)
	}
	assertStage0332Count(t, store, `
		SELECT count(*) FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, 1, subjectID, path, key)
}

func TestStage0338CommandCleanupWaitsUntilAfterExactRowAcquisition(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	subjectID := uuid.NewString()
	path := "/stage338/lock-order"
	keyA := "stage-03-38-lock-order-command-a"
	keyB := "stage-03-38-lock-order-command-b"
	idA := seedStage0338ExpiredCommandScope(t, store, subjectID, path, keyA)
	idB := seedStage0338ExpiredCommandScope(t, store, subjectID, path, keyB)
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(),
			`DELETE FROM investment.command_deduplication WHERE principal_id = $1`, subjectID)
	})

	blocker, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin exact-row blocker: %v", err)
	}
	if err := blocker.QueryRowContext(ctx, `
		SELECT id
		FROM investment.command_deduplication
		WHERE id = $1
		FOR UPDATE
	`, idA).Scan(&idA); err != nil {
		_ = blocker.Rollback()
		t.Fatalf("lock exact command row A: %v", err)
	}

	appName := "stage338-command-order-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	waiterStore := openStage0338NamedStore(t, databaseURL, appName)
	waiterTx, err := waiterStore.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		_ = blocker.Rollback()
		t.Fatalf("begin waiting reservation tx: %v", err)
	}
	commandA := verticalslice.CommandContext{
		SubjectID:      subjectID,
		IdempotencyKey: keyA,
		RequestHash:    strings.Repeat("c", 64),
		RequestPath:    path,
	}
	type reserveOutcome struct {
		reservation replayReservation
		err         error
	}
	outcome := make(chan reserveOutcome, 1)
	go func() {
		reservation, reserveErr := reserveReplayCommand(ctx, waiterTx, commandA, "POST")
		if reserveErr == nil {
			reserveErr = waiterTx.Commit()
		} else {
			_ = waiterTx.Rollback()
		}
		outcome <- reserveOutcome{reservation: reservation, err: reserveErr}
	}()

	waitForStage0338ApplicationLockWaiter(t, store, appName)
	assertStage0338RowLockAvailable(t, store, `
		SELECT id
		FROM investment.command_deduplication
		WHERE id = $1
		FOR UPDATE NOWAIT
	`, idB)

	if err := blocker.Commit(); err != nil {
		t.Fatalf("release exact-row blocker: %v", err)
	}
	got := <-outcome
	if got.err != nil {
		t.Fatalf("reservation after exact-row wait failed: %v", got.err)
	}
	if got.reservation.ID == "" {
		t.Fatal("reservation after exact-row wait returned empty identity")
	}
}

func TestStage0338DistinctExpiredReservationsCompleteWithoutDeadlock(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	subjectID := uuid.NewString()
	path := "/stage338/distinct-expired"
	keyA := "stage-03-38-distinct-expired-key-a"
	keyB := "stage-03-38-distinct-expired-key-b"
	seedStage0338ExpiredCommandScope(t, store, subjectID, path, keyA)
	seedStage0338ExpiredCommandScope(t, store, subjectID, path, keyB)
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(),
			`DELETE FROM investment.command_deduplication WHERE principal_id = $1`, subjectID)
	})

	commands := []verticalslice.CommandContext{
		{
			SubjectID:      subjectID,
			IdempotencyKey: keyA,
			RequestHash:    strings.Repeat("d", 64),
			RequestPath:    path,
		},
		{
			SubjectID:      subjectID,
			IdempotencyKey: keyB,
			RequestHash:    strings.Repeat("e", 64),
			RequestPath:    path,
		},
	}

	type result struct {
		reservation replayReservation
		err         error
	}
	results := make(chan result, len(commands))
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(len(commands))

	for _, command := range commands {
		command := command
		go func() {
			ready.Done()
			<-start
			requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			tx, beginErr := store.db.BeginTx(requestCtx, &sql.TxOptions{})
			if beginErr != nil {
				results <- result{err: beginErr}
				return
			}
			reservation, reserveErr := reserveReplayCommand(requestCtx, tx, command, "POST")
			if reserveErr == nil {
				reserveErr = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
			results <- result{reservation: reservation, err: reserveErr}
		}()
	}
	ready.Wait()
	close(start)

	seen := map[string]bool{}
	for i := 0; i < len(commands); i++ {
		got := <-results
		if got.err != nil {
			t.Fatalf("distinct expired reservation %d failed/deadlocked: %v", i+1, got.err)
		}
		if got.reservation.ID == "" {
			t.Fatalf("distinct expired reservation %d returned empty identity", i+1)
		}
		if seen[got.reservation.ID] {
			t.Fatalf("distinct expired scopes reused command identity %s", got.reservation.ID)
		}
		seen[got.reservation.ID] = true
	}
}

func TestStage0338AuthCleanupWaitsUntilAfterPresentedSessionSerialization(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	userA, sessionA, refreshA, csrfA := seedStage0338ExpiredPresentedSession(t, store)
	_, sessionB, _, _ := seedStage0338ExpiredPresentedSession(t, store)

	blocker := lockStage0338UserRefreshes(t, store, userA)
	appName := "stage338-auth-order-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	waiterStore := openStage0338NamedStore(t, databaseURL, appName)

	type revokeOutcome struct {
		revoked bool
		err     error
	}
	outcome := make(chan revokeOutcome, 1)
	go func() {
		revoked, revokeErr := waiterStore.RevokeSession(
			ctx,
			refreshA,
			csrfA,
			false,
			time.Now().UTC().Add(-24*time.Hour),
		)
		outcome <- revokeOutcome{revoked: revoked, err: revokeErr}
	}()

	waitForStage0338ApplicationLockWaiter(t, store, appName)
	assertStage0338RowLockAvailable(t, store, `
		SELECT id
		FROM identity.sessions
		WHERE id = $1
		FOR UPDATE NOWAIT
	`, sessionB)

	if err := blocker.Commit(); err != nil {
		t.Fatalf("release auth user blocker: %v", err)
	}
	got := <-outcome
	if got.revoked || !errors.Is(got.err, auth.ErrInvalidSession) {
		t.Fatalf("expired presented session must reject after serialization: revoked=%t err=%v", got.revoked, got.err)
	}

	// The presented expired row may be removed by the post-decision cleanup. The key guarantee
	// is that it had no authority and unrelated cleanup did not run before the canonical lock wait.
	var remaining int
	if err := store.db.QueryRowContext(ctx, `
		SELECT count(*) FROM identity.sessions WHERE id = $1
	`, sessionA).Scan(&remaining); err != nil {
		t.Fatalf("query expired presented session after rejection: %v", err)
	}
	if remaining != 0 && remaining != 1 {
		t.Fatalf("unexpected presented-session row count %d", remaining)
	}
}

func seedStage0338ExpiredCommandScope(
	t *testing.T,
	store *Store,
	subjectID string,
	path string,
	key string,
) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := store.db.ExecContext(context.Background(), `
		INSERT INTO investment.command_deduplication (
			id, principal_id, method, canonical_path, idempotency_key, request_hash,
			created_at, expires_at
		)
		VALUES (
			$1, $2, 'POST', $3, $4, $5,
			clock_timestamp() - interval '49 hours',
			clock_timestamp() - interval '25 hours'
		)
	`, id, subjectID, path, key, strings.Repeat("a", 64)); err != nil {
		t.Fatalf("seed expired command scope %s: %v", key, err)
	}
	return id
}

func seedStage0338ExpiredPresentedSession(
	t *testing.T,
	store *Store,
) (userID string, sessionID string, refreshHash string, csrfHash string) {
	t.Helper()
	ctx := context.Background()
	userID = uuid.NewString()
	sessionID = uuid.NewString()
	refreshHash = "stage338-expired-refresh-" + uuid.NewString()
	csrfHash = "stage338-expired-csrf-" + uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.users (id, email_normalized)
		VALUES ($1, $2)
	`, userID, "stage338-expired-presented-"+uuid.NewString()+"@example.com"); err != nil {
		t.Fatalf("seed expired-presented user: %v", err)
	}
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at
		)
		VALUES (
			$1, $2, $1, $3, $4, 'active',
			clock_timestamp() - interval '1 second',
			clock_timestamp() - interval '1 hour',
			clock_timestamp() - interval '1 hour'
		)
	`, sessionID, userID, refreshHash, csrfHash); err != nil {
		t.Fatalf("seed expired presented session: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.events WHERE actor_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.actors WHERE id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.users WHERE id = $1`, userID)
	})
	return userID, sessionID, refreshHash, csrfHash
}

func openStage0338NamedStore(t *testing.T, databaseURL string, applicationName string) *Store {
	t.Helper()
	separator := "?"
	if strings.Contains(databaseURL, "?") {
		separator = "&"
	}
	store, err := Open(databaseURL + separator + "application_name=" + applicationName)
	if err != nil {
		t.Fatalf("open named Stage 3.38 store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close named Stage 3.38 store: %v", err)
		}
	})
	return store
}

func waitForStage0338ApplicationLockWaiter(t *testing.T, store *Store, applicationName string) {
	t.Helper()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		if err := store.db.QueryRowContext(context.Background(), `
			SELECT count(*)
			FROM pg_stat_activity
			WHERE application_name = $1
				AND wait_event_type = 'Lock'
		`, applicationName).Scan(&waiting); err != nil {
			t.Fatalf("query Stage 3.38 named lock waiter: %v", err)
		}
		if waiting > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for application %s to block on a PostgreSQL lock", applicationName)
}

func assertStage0338RowLockAvailable(t *testing.T, store *Store, query string, args ...any) {
	t.Helper()
	tx, err := store.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin Stage 3.38 row-lock probe: %v", err)
	}
	defer rollback(tx)
	var id string
	if err := tx.QueryRowContext(context.Background(), query, args...).Scan(&id); err != nil {
		t.Fatalf("unrelated expired exact row was already locked by global cleanup: %v", err)
	}
}

func TestStage0338NoRowAdmissionTimestampIsPostUniqueConflictWait(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	subjectID := uuid.NewString()
	path := "/stage338/mixed-version-wait"
	key := "stage-03-38-mixed-version-conflict-key"
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(),
			`DELETE FROM investment.command_deduplication WHERE principal_id = $1`, subjectID)
	})

	// This transaction deliberately does NOT take the Stage 3.38 exact advisory lock. It models
	// an older/mixed-version writer that has inserted the unique scope but not committed yet.
	mixedTx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin mixed-version transaction: %v", err)
	}
	mixedID := uuid.NewString()
	if _, err := mixedTx.ExecContext(ctx, `
		INSERT INTO investment.command_deduplication (
			id, principal_id, method, canonical_path, idempotency_key, request_hash,
			created_at, expires_at
		)
		VALUES (
			$1, $2, 'POST', $3, $4, $5,
			clock_timestamp(),
			clock_timestamp() + interval '24 hours'
		)
	`, mixedID, subjectID, path, key, strings.Repeat("1", 64)); err != nil {
		_ = mixedTx.Rollback()
		t.Fatalf("insert uncommitted mixed-version conflict: %v", err)
	}

	appName := "stage338-mixed-wait-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	waiterStore := openStage0338NamedStore(t, databaseURL, appName)
	command := verticalslice.CommandContext{
		SubjectID:      subjectID,
		IdempotencyKey: key,
		RequestHash:    strings.Repeat("2", 64),
		RequestPath:    path,
	}

	type outcome struct {
		reservation replayReservation
		err         error
	}
	resultCh := make(chan outcome, 1)
	go func() {
		tx, beginErr := waiterStore.db.BeginTx(ctx, &sql.TxOptions{})
		if beginErr != nil {
			resultCh <- outcome{err: beginErr}
			return
		}
		reservation, reserveErr := reserveReplayCommand(ctx, tx, command, "POST")
		if reserveErr == nil {
			reserveErr = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
		resultCh <- outcome{reservation: reservation, err: reserveErr}
	}()

	waitForStage0338ApplicationLockWaiter(t, store, appName)

	// Hold the unique-index wait across a measurable database-clock interval.
	time.Sleep(300 * time.Millisecond)
	var releaseBoundary time.Time
	if err := mixedTx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&releaseBoundary); err != nil {
		_ = mixedTx.Rollback()
		t.Fatalf("sample mixed-version rollback boundary: %v", err)
	}
	if err := mixedTx.Rollback(); err != nil {
		t.Fatalf("roll back mixed-version conflicting writer: %v", err)
	}

	got := <-resultCh
	if got.err != nil {
		t.Fatalf("reservation after mixed-version rollback: %v", got.err)
	}
	if got.reservation.ID == "" || got.reservation.ID == mixedID {
		t.Fatalf("fresh winner identity invalid: mixed=%s reservation=%s", mixedID, got.reservation.ID)
	}

	var createdAt time.Time
	var expiresAt time.Time
	if err := store.db.QueryRowContext(ctx, `
		SELECT created_at, expires_at
		FROM investment.command_deduplication
		WHERE id = $1
	`, got.reservation.ID).Scan(&createdAt, &expiresAt); err != nil {
		t.Fatalf("query post-wait admitted generation: %v", err)
	}
	if createdAt.Before(releaseBoundary) {
		t.Fatalf(
			"fresh admission timestamp was sampled before unique-conflict release: release=%s created=%s",
			releaseBoundary,
			createdAt,
		)
	}
	if expiresAt.Sub(createdAt) != 24*time.Hour {
		t.Fatalf("post-wait retention interval = %s, want 24h", expiresAt.Sub(createdAt))
	}
}

func TestStage0338AuthCleanupRunsAfterUserWideRevocationLocks(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	userA := seedStage0338ActiveUserWithSibling(t, store, false)
	userB := seedStage0338ActiveUserWithSibling(t, store, true)

	// Hold a row that allSessions=true must update for user A. The auth request should block in
	// its user-wide revocation BEFORE global cleanup is allowed to touch unrelated user-B expiry.
	blocker, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin user-wide row blocker: %v", err)
	}
	var blockedID string
	if err := blocker.QueryRowContext(ctx, `
		SELECT id FROM identity.sessions WHERE id = $1 FOR UPDATE
	`, userA.siblingID).Scan(&blockedID); err != nil {
		_ = blocker.Rollback()
		t.Fatalf("lock user-A sibling row: %v", err)
	}

	appName := "stage338-auth-userwide-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	waiterStore := openStage0338NamedStore(t, databaseURL, appName)
	type outcome struct {
		revoked bool
		err     error
	}
	resultCh := make(chan outcome, 1)
	go func() {
		revoked, revokeErr := waiterStore.RevokeSession(
			ctx,
			userA.presentedRefreshHash,
			userA.presentedCSRFHash,
			true,
			time.Now().UTC(),
		)
		resultCh <- outcome{revoked: revoked, err: revokeErr}
	}()

	waitForStage0338ApplicationLockWaiter(t, store, appName)

	// userB.siblingID is the only intentionally expired unrelated session after the initial drain.
	// It must remain obtainable while user A is blocked in its broader user-wide UPDATE.
	assertStage0338RowLockAvailable(t, store, `
		SELECT id
		FROM identity.sessions
		WHERE id = $1
		FOR UPDATE NOWAIT
	`, userB.siblingID)

	if err := blocker.Commit(); err != nil {
		t.Fatalf("release user-wide row blocker: %v", err)
	}
	got := <-resultCh
	if got.err != nil || !got.revoked {
		t.Fatalf("all-sessions logout after user-wide lock wait failed: revoked=%t err=%v", got.revoked, got.err)
	}
}

func TestStage0338TwoUserAllSessionsMutationsCompleteWithoutCleanupDeadlock(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	drainStage0338ExpiredRows(t, store)

	userA := seedStage0338ActiveUserWithSibling(t, store, true)
	userB := seedStage0338ActiveUserWithSibling(t, store, true)
	users := []stage0338ActiveUserWithSibling{userA, userB}

	type outcome struct {
		revoked bool
		err     error
	}
	results := make(chan outcome, 2)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)

	for _, fixture := range users {
		fixture := fixture
		go func() {
			ready.Done()
			<-start
			requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			revoked, err := store.RevokeSession(
				requestCtx,
				fixture.presentedRefreshHash,
				fixture.presentedCSRFHash,
				true,
				time.Now().UTC(),
			)
			results <- outcome{revoked: revoked, err: err}
		}()
	}
	ready.Wait()
	close(start)

	for i := 0; i < 2; i++ {
		got := <-results
		if got.err != nil || !got.revoked {
			t.Fatalf("two-user all-sessions mutation %d failed/deadlocked: revoked=%t err=%v", i+1, got.revoked, got.err)
		}
	}
}

type stage0338ActiveUserWithSibling struct {
	userID               string
	presentedID          string
	presentedRefreshHash string
	presentedCSRFHash    string
	siblingID            string
}

func seedStage0338ActiveUserWithSibling(
	t *testing.T,
	store *Store,
	expireSibling bool,
) stage0338ActiveUserWithSibling {
	t.Helper()
	ctx := context.Background()
	userID := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.users (id, email_normalized)
		VALUES ($1, $2)
	`, userID, "stage338-active-sibling-"+uuid.NewString()+"@example.com"); err != nil {
		t.Fatalf("insert Stage 3.38 active-sibling user: %v", err)
	}

	presentedID := uuid.NewString()
	presentedRefreshHash := "stage338-presented-refresh-" + uuid.NewString()
	presentedCSRFHash := "stage338-presented-csrf-" + uuid.NewString()
	if _, err := store.db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at
		)
		VALUES (
			$1, $2, $1, $3, $4, 'active',
			clock_timestamp() + interval '1 day',
			clock_timestamp() - interval '1 hour',
			clock_timestamp() - interval '1 hour'
		)
	`, presentedID, userID, presentedRefreshHash, presentedCSRFHash); err != nil {
		t.Fatalf("insert Stage 3.38 presented session: %v", err)
	}

	siblingID := uuid.NewString()
	siblingExpiry := "clock_timestamp() + interval '1 day'"
	if expireSibling {
		siblingExpiry = "clock_timestamp() - interval '1 second'"
	}
	query := `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash,
			session_state, expires_at, created_at, last_rotated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, 'active',
			` + siblingExpiry + `,
			clock_timestamp() - interval '2 hours',
			clock_timestamp() - interval '2 hours'
		)
	`
	if _, err := store.db.ExecContext(
		ctx,
		query,
		siblingID,
		userID,
		presentedID,
		"stage338-sibling-refresh-"+uuid.NewString(),
		"stage338-sibling-csrf-"+uuid.NewString(),
	); err != nil {
		t.Fatalf("insert Stage 3.38 sibling session: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.events WHERE actor_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM audit.actors WHERE id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.sessions WHERE user_id = $1`, userID)
		_, _ = store.db.ExecContext(cleanupCtx, `DELETE FROM identity.users WHERE id = $1`, userID)
	})

	return stage0338ActiveUserWithSibling{
		userID:               userID,
		presentedID:          presentedID,
		presentedRefreshHash: presentedRefreshHash,
		presentedCSRFHash:    presentedCSRFHash,
		siblingID:            siblingID,
	}
}
