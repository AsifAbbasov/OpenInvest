<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

Next.js APIs and conventions may differ from model training data. Before changing Web code, read
the relevant guide in `node_modules/next/dist/docs/` and follow current deprecation notices.

This application is presentation-only under `docs/ADR/ADR-007-use-nextjs-for-web-frontend.md`.
Do not add business logic, business Route Handlers, database access, financial calculations,
external-source integrations, LocalStorage business persistence, or mobile code. All business data
comes through the OpenAPI-defined Go API.
<!-- END:nextjs-agent-rules -->
