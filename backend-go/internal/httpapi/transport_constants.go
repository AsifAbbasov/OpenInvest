package httpapi

import (
	"time"
)

const devSubjectID = "00000000-0000-4000-8000-000000000001"

const maxHTTPImportPayloadBytes = 2 * 1024 * 1024

const maxHTTPImportRows = 100

const authRateLimitRetryAfterSeconds = "60"

const importFreshRateRetryAfterSeconds = "60"

const importExecutionRateRetryAfterSeconds = "60"

const importCapacityRetryAfterSeconds = "1"

const defaultImportExecutionPerSubjectLimit = 12

const defaultImportExecutionGlobalLimit = 120

const defaultImportExecutionMaxSubjects = 2048

const defaultImportExecutionWindow = time.Minute

const defaultImportFreshPerSubjectLimit = 6

const defaultImportFreshGlobalLimit = 60

const defaultImportFreshMaxSubjects = 2048

const defaultImportFreshWindow = time.Minute

const defaultImportActiveCapacity = 2

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
