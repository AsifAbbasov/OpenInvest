package moexiss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type parsedQuote struct {
	price decimal.Decimal
	asOf  time.Time
}

type issResponse struct {
	MarketData  *issTable `json:"marketdata"`
	DataVersion *issTable `json:"dataversion"`
}

type issTable struct {
	Columns *[]string            `json:"columns"`
	Data    *[][]json.RawMessage `json:"data"`
}

func parseResponse(body []byte, ticker string, moscow *time.Location) (parsedQuote, error) {
	var payload issResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&payload); err != nil {
		return parsedQuote{}, providerDataError("malformed JSON")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return parsedQuote{}, err
	}
	if payload.MarketData == nil {
		return parsedQuote{}, providerDataError("marketdata block is missing")
	}
	if payload.DataVersion == nil {
		return parsedQuote{}, providerDataError("dataversion block is missing")
	}

	versionIndexes, err := requiredIndexes(payload.DataVersion, "dataversion", "trade_date")
	if err != nil {
		return parsedQuote{}, err
	}
	if payload.DataVersion.Data == nil {
		return parsedQuote{}, providerDataError("dataversion.data is missing")
	}
	versionRows := *payload.DataVersion.Data
	if len(versionRows) != 1 {
		return parsedQuote{}, providerDataError("dataversion must contain exactly one row")
	}
	versionRow := versionRows[0]
	if payload.DataVersion.Columns == nil || len(versionRow) != len(*payload.DataVersion.Columns) {
		return parsedQuote{}, providerDataError("dataversion row shape is invalid")
	}
	tradeDate, err := rawString(versionRow[versionIndexes["trade_date"]])
	if err != nil {
		return parsedQuote{}, providerDataError("dataversion trade_date is invalid")
	}
	if _, err := time.Parse("2006-01-02", tradeDate); err != nil {
		return parsedQuote{}, providerDataError("dataversion trade_date is invalid")
	}

	marketIndexes, err := requiredIndexes(payload.MarketData, "marketdata", "SECID", "BOARDID", "LAST", "TIME")
	if err != nil {
		return parsedQuote{}, err
	}
	if payload.MarketData.Data == nil {
		return parsedQuote{}, providerDataError("marketdata.data is missing")
	}
	marketRows := *payload.MarketData.Data
	if len(marketRows) == 0 {
		return parsedQuote{}, verticalslice.ErrMarketQuoteNotFound
	}
	if len(marketRows) != 1 {
		return parsedQuote{}, providerDataError("marketdata must contain exactly one row")
	}
	marketRow := marketRows[0]
	if payload.MarketData.Columns == nil || len(marketRow) != len(*payload.MarketData.Columns) {
		return parsedQuote{}, providerDataError("marketdata row shape is invalid")
	}

	secid, err := rawString(marketRow[marketIndexes["SECID"]])
	if err != nil || secid != ticker {
		return parsedQuote{}, providerDataError("marketdata SECID does not match requested ticker")
	}
	board, err := rawString(marketRow[marketIndexes["BOARDID"]])
	if err != nil || board != boardID {
		return parsedQuote{}, providerDataError("marketdata BOARDID is invalid")
	}

	last := bytes.TrimSpace(marketRow[marketIndexes["LAST"]])
	if bytes.Equal(last, []byte("null")) {
		return parsedQuote{}, verticalslice.ErrMarketQuoteNotFound
	}
	if len(last) == 0 || !json.Valid(last) {
		return parsedQuote{}, providerDataError("marketdata LAST is invalid")
	}
	price, err := decimal.FromString(string(last))
	if err != nil || !price.FitsStorage() || price.IsNegative() {
		return parsedQuote{}, providerDataError("marketdata LAST is not a canonical non-negative decimal")
	}

	marketTime, err := rawString(marketRow[marketIndexes["TIME"]])
	if err != nil {
		return parsedQuote{}, providerDataError("marketdata TIME is invalid")
	}
	asOf, err := time.ParseInLocation("2006-01-02 15:04:05", tradeDate+" "+marketTime, moscow)
	if err != nil {
		return parsedQuote{}, providerDataError("trade date/time is invalid")
	}

	return parsedQuote{price: price, asOf: asOf.UTC()}, nil
}

func requiredIndexes(table *issTable, block string, required ...string) (map[string]int, error) {
	if table.Columns == nil {
		return nil, providerDataError(block + ".columns is missing")
	}
	columns := *table.Columns
	wanted := make(map[string]struct{}, len(required))
	for _, name := range required {
		wanted[name] = struct{}{}
	}
	indexes := make(map[string]int, len(required))
	for index, name := range columns {
		if _, requiredName := wanted[name]; !requiredName {
			continue
		}
		if _, duplicate := indexes[name]; duplicate {
			return nil, providerDataError(block + " contains duplicate required column " + name)
		}
		indexes[name] = index
	}
	for _, name := range required {
		if _, ok := indexes[name]; !ok {
			return nil, providerDataError(block + " is missing required column " + name)
		}
	}
	return indexes, nil
}

func rawString(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	return value, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	}
	return providerDataError("response contains trailing JSON content")
}

func providerDataError(message string) error {
	return fmt.Errorf("%w: %s", verticalslice.ErrMarketQuoteProviderData, message)
}
