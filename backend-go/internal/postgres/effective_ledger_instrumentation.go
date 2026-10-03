package postgres

import (
	"context"
	"sync/atomic"
)

// effectiveLedgerInstrumentation is an opt-in, context-scoped test seam. It is absent from
// ordinary requests and records authoritative effective-ledger materializations only.
type effectiveLedgerInstrumentation struct {
	materializations atomic.Int64
}

type effectiveLedgerInstrumentationContextKey struct{}

func withEffectiveLedgerInstrumentation(ctx context.Context, instrumentation *effectiveLedgerInstrumentation) context.Context {
	if instrumentation == nil {
		return ctx
	}
	return context.WithValue(ctx, effectiveLedgerInstrumentationContextKey{}, instrumentation)
}

func countEffectiveLedgerMaterialization(ctx context.Context) {
	if ctx == nil {
		return
	}
	if instrumentation, _ := ctx.Value(effectiveLedgerInstrumentationContextKey{}).(*effectiveLedgerInstrumentation); instrumentation != nil {
		instrumentation.materializations.Add(1)
	}
}
