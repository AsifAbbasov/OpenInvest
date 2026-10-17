#!/usr/bin/env bash
# Bounded, collect-all coverage-guided smoke checks; no database or provider calls.
set -u
cd "$(dirname "$0")/../backend-go"
failed=0
while read -r package target; do
  printf 'NATIVE_FUZZ_START package=%s target=%s\n' "$package" "$target"
  if timeout 90s go test "$package" -run='^$' -fuzz="^${target}$" -fuzztime=2s -parallel=1 -timeout=15s; then
    printf 'NATIVE_FUZZ_RESULT target=%s result=PASS\n' "$target"
  else
    printf 'NATIVE_FUZZ_RESULT target=%s result=FAIL\n' "$target"
    failed=1
  fi
done <<'TARGETS'
./internal/auth FuzzAccessTokenByteMutation
./internal/decimal FuzzDecimalFixedScaleAlgebra
./internal/decimal FuzzFromString
./internal/httpapi FuzzParseMoney
./internal/httpapi FuzzPaginationCursorScope
./internal/httpapi FuzzLosslessPasswordJSON
./internal/httpapi FuzzTransactionRouteUUID
./internal/httpapi FuzzDecodeStrictJSON
./internal/importer FuzzReviewCSVStructuralBounds
./internal/importer FuzzReviewCSV
./internal/importer FuzzValidateCSVHeaderShape
./internal/position FuzzStatefulWACLedgerSequence
./internal/verticalslice FuzzXIRRCoefficients
./internal/verticalslice FuzzCorporateActionProjectionIntegrity
./internal/verticalslice FuzzPortfolioXIRR
./internal/verticalslice FuzzParseBusinessDate
./internal/verticalslice FuzzBuildPortfolioReturnProjection
./internal/verticalslice FuzzXIRRTwoTermProjection
TARGETS
exit "$failed"
