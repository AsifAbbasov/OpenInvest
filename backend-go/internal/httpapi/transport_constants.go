package httpapi

import (
	"time"
)

const devSubjectID = "00000000-0000-4000-8000-000000000001"

const maxHTTPImportPayloadBytes = 2 * 1024 * 1024

const maxHTTPImportRows = 100

const authRateLimitRetryAfterSeconds = "60"

const dividendCalculatorRateLimitRetryAfterSeconds = "60"

const defaultDividendCalculatorPerKeyLimit = 20

const defaultDividendCalculatorGlobalLimit = 1200

const defaultDividendCalculatorMaxKeys = 4096

const defaultDividendCalculatorWindow = time.Minute

const defaultAuthRateLimiterMaxKeys = 2048

const defaultAuthRateLimiterGlobalLimit = 2000

const minImportReviewTokenSecretBytes = 32

const maxPaginationCursorBytes = 512

const maxImportReviewTokenBytes = 16384

const importReviewTokenVersion = 1

const importReviewTokenTTL = 15 * time.Minute
