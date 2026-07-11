"use client";

import { useState } from "react";

import { createPortfolio } from "@/common/api/openinvest";

type CreatePortfolioFormProps = {
  accessToken: string;
  onCreated: () => void;
};

export function CreatePortfolioForm({ accessToken, onCreated }: CreatePortfolioFormProps) {
  const [name, setName] = useState("Long-term RUB portfolio");
  const [status, setStatus] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setStatus(null);
    const result = await createPortfolio({ name: name.trim(), baseCurrency: "RUB" }, { accessToken });
    setIsSubmitting(false);
    if (!result.ok) {
      setStatus(result.message);
      return;
    }
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
