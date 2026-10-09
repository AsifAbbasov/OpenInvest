package httpapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func m11V2TransactionBody() string {
	return `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"100.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-08","settlementDate":null,"note":null}`
}

func TestM11V2HTTPParserRoutingHostileMatrix(t *testing.T) {
	validID := "00000000-0000-4000-8000-000000000002"
	type testCase struct {
		name       string
		path       string
		body       []byte
		content    string
		headerMode string
		wantStatus int
		wantCalls  int
		flexible   bool
	}
	cases := []testCase{
		{name:"valid", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantStatus:http.StatusCreated, wantCalls:1},
		{name:"uppercase uuid", path:"/api/v1/portfolios/"+strings.ToUpper(validID)+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantStatus:http.StatusCreated, wantCalls:1},
		{name:"malformed uuid", path:"/api/v1/portfolios/not-a-uuid/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"percent malformed uuid", path:"/api/v1/portfolios/%6Eot-a-uuid/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"encoded slash", path:"/api/v1/portfolios/"+validID+"%2Fextra/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantCalls:0, flexible:true},
		{name:"repeated slash", path:"/api/v1/portfolios//"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantCalls:0, flexible:true},
		{name:"dot segment", path:"/api/v1/portfolios/../"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", wantCalls:0, flexible:true},
		{name:"empty body", path:"/api/v1/portfolios/"+validID+"/transactions", body:nil, content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"trailing json", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()+" {}"), content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"multiple json", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()+m11V2TransactionBody()), content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"unknown field", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(strings.Replace(m11V2TransactionBody(), `"note":null`, `"note":null,"unexpected":true`, 1)), content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"malformed date", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(strings.Replace(m11V2TransactionBody(), "2026-10-08", "2026-02-30", 1)), content:"application/json", wantStatus:http.StatusBadRequest, wantCalls:0},
		{name:"mixed case idempotency", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", headerMode:"mixed", wantStatus:http.StatusCreated, wantCalls:1},
		{name:"duplicate idempotency", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json", headerMode:"duplicate", wantCalls:1, flexible:true},
		{name:"json charset", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"application/json; charset=utf-8", wantStatus:http.StatusCreated, wantCalls:1},
		{name:"text plain observation", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"text/plain", wantCalls:1, flexible:true},
		{name:"missing content type observation", path:"/api/v1/portfolios/"+validID+"/transactions", body:[]byte(m11V2TransactionBody()), content:"", wantCalls:1, flexible:true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &stage374HTTPStore{}
			app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
			req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader(tc.body))
			if tc.content != "" { req.Header.Set("Content-Type", tc.content) }
			switch tc.headerMode {
			case "mixed":
				req.Header.Set("iDeMpOtEnCy-KeY", "m11-v2-mixed-key-0001")
			case "duplicate":
				req.Header.Add("Idempotency-Key", "m11-v2-duplicate-key-a-0001")
				req.Header.Add("Idempotency-Key", "m11-v2-duplicate-key-b-0001")
			default:
				req.Header.Set("Idempotency-Key", "m11-v2-parser-"+strings.ReplaceAll(tc.name, " ", "-")+"-0001")
			}
			resp, err := app.Test(req)
			if err != nil { t.Fatal(err) }
			defer resp.Body.Close()
			if !tc.flexible && resp.StatusCode != tc.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("status=%d want=%d body=%s", resp.StatusCode, tc.wantStatus, string(body))
			}
			if tc.flexible && resp.StatusCode >= 500 {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("hostile case produced 5xx status=%d body=%s", resp.StatusCode, string(body))
			}
			if tc.flexible {
				if store.appendCalls > tc.wantCalls {
					t.Fatalf("downstream amplification calls=%d max=%d", store.appendCalls, tc.wantCalls)
				}
			} else if store.appendCalls != tc.wantCalls {
				t.Fatalf("downstream calls=%d want=%d", store.appendCalls, tc.wantCalls)
			}
			t.Logf("M11_V2_PARSER_CASE=%s STATUS=%d CALLS=%d", tc.name, resp.StatusCode, store.appendCalls)
		})
	}
	t.Logf("M11_V2_HTTP_PARSER_ROUTING_CASES=%d", len(cases))
}
