package httpapi

import (
	"fmt"
	"os"
	"testing"
)

func TestM11V2RepeatedProviderConcurrencyCampaign(t *testing.T) {
	if os.Getenv("OPENINVEST_M11_V2_EXPENSIVE") != "1" { t.Skip("specialized Module 11 V2 campaign only") }
	const identicalIterations = 100
	for i := 0; i < identicalIterations; i++ {
		t.Run(fmt.Sprintf("identical-%03d", i), TestModule10RemediationIdenticalConcurrentCoalescing)
	}
	const cancellationIterations = 25
	for i := 0; i < cancellationIterations; i++ {
		t.Run(fmt.Sprintf("cancel-%03d", i), TestModule10RemediationCoalescerErrorAndCancellationRecovery)
	}
	t.Logf("M11_V2_PROVIDER_IDENTICAL_CONCURRENCY_ITERATIONS=%d", identicalIterations)
	t.Logf("M11_V2_PROVIDER_CANCELLATION_RECOVERY_ITERATIONS=%d", cancellationIterations)
}
