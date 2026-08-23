"use client";

import { useRef, useState } from "react";

import { createPortfolio } from "@/common/api/openinvest";
import {
  clearBrowserIdempotencyIntent,
  emptyIdempotencyIntent,
  idempotencyIntentForBrowser,
  principalScopedIdempotencyScope,
} from "@/common/api/idempotency";

type CreatePortfolioFormProps = {
  accessToken: string;
  principalId: string;
  onCreated: () => void;
};

const idempotencyConflictMessage = "Idempotency-Key is already bound to another request";

export function CreatePortfolioForm({ accessToken, principalId, onCreated }: CreatePortfolioFormProps) {
  const idempotencyIntentRef = useRef(emptyIdempotencyIntent);
  const [name, setName] = useState("Long-term RUB portfolio");
  const [status, setStatus] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const retryScope = principalScopedIdempotencyScope(principalId, "portfolio-create");

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setStatus(null);
    const payload = { name: name.trim(), baseCurrency: "RUB" as const };
    const intent = JSON.stringify(payload);
    idempotencyIntentRef.current = await idempotencyIntentForBrowser(
      idempotencyIntentRef.current,
      intent,
      retryScope,
    );
    const result = await createPortfolio(
      payload,
      { accessToken, idempotencyKey: idempotencyIntentRef.current.key ?? undefined },
    );
    setIsSubmitting(false);
    if (!result.ok) {
      if (result.status === 409 && result.message === idempotencyConflictMessage) {
        // After a reload the browser deliberately retries the unresolved technical key without
        // persisting the business payload. A server conflict proves the current payload is a new
        // intent, so the old retry key can now be safely abandoned for this principal only.
        await clearBrowserIdempotencyIntent(retryScope);
        idempotencyIntentRef.current = emptyIdempotencyIntent;
      }
      setStatus(result.message);
      return;
    }
    await clearBrowserIdempotencyIntent(retryScope);
    idempotencyIntentRef.current = emptyIdempotencyIntent;
    setStatus("Portfolio created. Loading portfolio data from Go API…");
    onCreated();
  }

  return (
    <form className="panel form-grid" onSubmit={submit}>
      <div>
        <p className="eyebrow">First portfolio</p>
        <h2>Create a RUB portfolio</h2>
        <p className="muted">
          The web layer sends the request to the Go API. Portfolio ownership, idempotency, and ledger
          rules remain on the backend.
        </p>
      </div>
      <label>
        Portfolio name
        <input value={name} maxLength={100} required onChange={(event) => setName(event.target.value)} />
      </label>
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Creating…" : "Create portfolio"}
      </button>
      {status ? <p className="form-status">{status}</p> : null}
    </form>
  );
}
