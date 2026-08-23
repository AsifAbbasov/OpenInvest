package httpapi

import (
	"context"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// Existing HTTP fakes model the pre-Stage-3.33 Store interface. Keep their transaction behavior and
// expose the new exact-outcome boundary explicitly so legacy handler tests exercise the same contract.
func (store *importAPITestStore) AppendImportedTransactionsWithOutcome(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
) (verticalslice.ImportAppendOutcome, error) {
	transactions, err := store.AppendImportedTransactions(ctx, command, request)
	if err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	dates := make([]string, 0, len(request.Transactions))
	seen := map[string]struct{}{}
	for _, item := range request.Transactions {
		if _, ok := seen[item.TradeDate]; ok {
			continue
		}
		seen[item.TradeDate] = struct{}{}
		dates = append(dates, item.TradeDate)
	}
	return verticalslice.ImportAppendOutcome{
		Transactions:         transactions,
		SnapshotDatesRebuilt: dates,
	}, nil
}
