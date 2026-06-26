export default function HomePage() {
  return (
    <main className="shell">
      <p className="eyebrow">Personal Capital Operating System</p>
      <h1>OpenInvest</h1>
      <p className="summary">
        The Next.js presentation-layer skeleton is running. Business data and calculations will
        come exclusively from the OpenAPI-defined Go API in later approved stages.
      </p>
      <dl className="status" aria-label="Architecture status">
        <div>
          <dt>Web</dt>
          <dd>Next.js App Router</dd>
        </div>
        <div>
          <dt>Business API</dt>
          <dd>Go remains canonical</dd>
        </div>
        <div>
          <dt>Current scope</dt>
          <dd>Architecture skeleton only</dd>
        </div>
      </dl>
    </main>
  );
}
