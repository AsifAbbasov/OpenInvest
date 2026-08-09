"use client";

import { useRef, useState } from "react";

import { createPortfolio } from "@/common/api/openinvest";
import { emptyIdempotencyIntent, idempotencyIntentFor } from "@/common/api/idempotency";

type CreatePortfolioFormProps = {
  accessToken: string;
  onCreated: () => void;
};

export function CreatePortfolioForm({ accessToken, onCreated }: CreatePortfolioFormProps) {
  const idempotencyIntentRef = useRef(emptyIdempotencyIntent);
  const [name, setName] = useState("Long-term RUB portfolio");
  const [status, setStatus] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setStatus(null);
    const payload = { name: name.trim(), baseCurrency: "RUB" as const };
    const intent = JSON.stringify(payload);
    idempotencyIntentRef.current = idempotencyIntentFor(idempotencyIntentRef.current, intent, () => crypto.randomUUID());
    const result = await createPortfolio(
      payload,
      { accessToken, idempotencyKey: idempotencyIntentRef.current.key ?? undefined },
    );
    setIsSubmitting(false);
    if (!result.ok) {
      setStatus(result.message);
      return;
    }
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
