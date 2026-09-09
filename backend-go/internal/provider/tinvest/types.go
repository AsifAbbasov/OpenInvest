package tinvest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type corporateActionRequest struct {
	InstrumentID string `json:"instrumentId"`
	From         string `json:"from"`
	To           string `json:"to"`
}

type moneyValue struct {
	Currency string     `json:"currency"`
	Units    protoInt64 `json:"units"`
	Nano     int32      `json:"nano"`
}

type dividendsResponse struct {
	Dividends []dividend `json:"dividends"`
}

type dividend struct {
	DividendNet  *moneyValue `json:"dividendNet"`
	PaymentDate  string      `json:"paymentDate"`
	RecordDate   string      `json:"recordDate"`
	DividendType string      `json:"dividendType"`
}

type bondCouponsResponse struct {
	Events []coupon `json:"events"`
}

type coupon struct {
	CouponDate   string      `json:"couponDate"`
	CouponNumber protoInt64  `json:"couponNumber"`
	FixDate      string      `json:"fixDate"`
	PayOneBond   *moneyValue `json:"payOneBond"`
}

// protoInt64 accepts protobuf JSON's canonical quoted int64 representation and
// an exact unquoted JSON integer. It never parses through float64.
type protoInt64 int64

func (value *protoInt64) UnmarshalJSON(data []byte) error {
	if value == nil {
		return fmt.Errorf("tinvest: nil proto int64 destination")
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return fmt.Errorf("tinvest: invalid int64")
	}

	var raw string
	if data[0] == '"' {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("tinvest: invalid quoted int64")
		}
	} else {
		raw = string(data)
	}
	if raw == "" || raw[0] == '+' {
		return fmt.Errorf("tinvest: invalid int64")
	}
	if len(raw) > 1 && raw[0] == '0' {
		return fmt.Errorf("tinvest: non-canonical int64")
	}
	if len(raw) > 2 && raw[0] == '-' && raw[1] == '0' {
		return fmt.Errorf("tinvest: non-canonical int64")
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("tinvest: invalid int64")
	}
	*value = protoInt64(parsed)
	return nil
}
