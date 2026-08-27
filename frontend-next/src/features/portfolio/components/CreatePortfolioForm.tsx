"use client";

import { useRef, useState } from "react";

import { createPortfolio } from "@/common/api/openinvest";
import {
  clearBrowserIdempotencyIntent,
  emptyIdempotencyIntent,
  idempotencyIntentForBrowser,
  principalScopedIdempotencyScope,
} from "@/common/api/idempotency";
import { unicodeTextValidationError } from "@/common/presentation/unicode";

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
    setStatus(null);
    const normalizedName = name.trim();
    if (normalizedName === "") {
      setStatus("Portfolio name is required.");
      return;
    }
    const nameProblem = unicodeTextValidationError(normalizedName, 100);
    if (nameProblem === "ILL_FORMED") {
      setStatus("Portfolio name contains invalid Unicode.");
      return;
    }
    if (nameProblem === "TOO_LONG") {
      setStatus("Portfolio name must be at most 100 Unicode code points.");
      return;
    }
    setIsSubmitting(true);
    const payload = { name: normalizedName, baseCurrency: "RUB" as const };
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
        <input value={name} required onChange={(event) => setName(event.target.value)} />
      </label>
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Creating…" : "Create portfolio"}
      </button>
      {status ? <p className="form-status">{status}</p> : null}
    </form>
  );
}
