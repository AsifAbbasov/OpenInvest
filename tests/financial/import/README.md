# Stage 3.6 Import Test Vectors

These fixtures define the initial broker-file import vectors for the Stage 3.6 reconciliation
slice.

Scope:

- user-supplied CSV only;
- RUB only;
- no broker API;
- no credentials;
- no parser-side append;
- no raw-row persistence.

The importer tests use equivalent inline vectors for fast unit coverage. These files are kept as
reviewable canonical examples for future CLI/API fixture tests.
