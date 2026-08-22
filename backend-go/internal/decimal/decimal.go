package decimal

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	Scale     = 8
	Precision = 28
)

var scaleFactor = big.NewInt(100000000)

type Decimal struct {
	value *big.Int
}

func Zero() Decimal {
	return Decimal{value: big.NewInt(0)}
}

func FromString(input string) (Decimal, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Zero(), errors.New("decimal is empty")
	}

	sign := 1
	if strings.HasPrefix(trimmed, "-") {
		sign = -1
		trimmed = strings.TrimPrefix(trimmed, "-")
	} else if strings.HasPrefix(trimmed, "+") {
		trimmed = strings.TrimPrefix(trimmed, "+")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 || parts[0] == "" {
		return Zero(), fmt.Errorf("invalid decimal %q", input)
	}

	whole := parts[0]
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > Scale {
		return Zero(), fmt.Errorf("decimal %q exceeds %d fractional digits", input, Scale)
	}
	for _, char := range whole + fraction {
		if char < '0' || char > '9' {
			return Zero(), fmt.Errorf("invalid decimal %q", input)
		}
	}

	fraction = fraction + strings.Repeat("0", Scale-len(fraction))
	canonicalDigits := strings.TrimLeft(whole+fraction, "0")
	if len(canonicalDigits) > Precision {
		return Zero(), fmt.Errorf("decimal %q exceeds NUMERIC(%d,%d) precision", input, Precision, Scale)
	}
	unscaled := new(big.Int)
	if _, ok := unscaled.SetString(whole+fraction, 10); !ok {
		return Zero(), fmt.Errorf("invalid decimal %q", input)
	}
	if sign < 0 {
		unscaled.Neg(unscaled)
	}

	return Decimal{value: unscaled}, nil
}

func Must(input string) Decimal {
	value, err := FromString(input)
	if err != nil {
		panic(err)
	}
	return value
}

func (d Decimal) String() string {
	value := new(big.Int).Set(d.value)
	sign := ""
	if value.Sign() < 0 {
		sign = "-"
		value.Abs(value)
	}

	text := value.Text(10)
	if len(text) <= Scale {
		text = strings.Repeat("0", Scale-len(text)+1) + text
	}

	whole := text[:len(text)-Scale]
	fraction := text[len(text)-Scale:]
	return sign + whole + "." + fraction
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(d.String())), nil
}

func (d Decimal) Equal(other Decimal) bool {
	return d.value.Cmp(other.value) == 0
}

func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{value: new(big.Int).Add(d.value, other.value)}
}

func (d Decimal) Sub(other Decimal) Decimal {
	return Decimal{value: new(big.Int).Sub(d.value, other.value)}
}

func (d Decimal) Mul(other Decimal) Decimal {
	product := new(big.Int).Mul(d.value, other.value)
	return Decimal{value: divHalfEven(product, scaleFactor)}
}

func (d Decimal) Div(other Decimal) (Decimal, error) {
	if other.value.Sign() == 0 {
		return Zero(), errors.New("divide by zero")
	}
	scaled := new(big.Int).Mul(d.value, scaleFactor)
	return Decimal{value: divHalfEven(scaled, other.value)}, nil
}

func (d Decimal) IsNegative() bool {
	return d.value.Sign() < 0
}

func (d Decimal) IsZero() bool {
	return d.value.Sign() == 0
}

func (d Decimal) IsPositive() bool {
	return d.value.Sign() > 0
}

// FitsStorage reports whether the normalized value fits the canonical PostgreSQL
// NUMERIC(28,8) storage contract. Arithmetic can grow beyond the ingress precision,
// so persistence-bound derived values must re-check this invariant.
func (d Decimal) FitsStorage() bool {
	if d.value == nil {
		return false
	}
	value := new(big.Int).Abs(new(big.Int).Set(d.value))
	return len(value.Text(10)) <= Precision
}

func divHalfEven(numerator *big.Int, denominator *big.Int) *big.Int {
	quotient, remainder := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	if remainder.Sign() == 0 {
		return quotient
	}

	absRemainder := new(big.Int).Abs(remainder)
	absDenominator := new(big.Int).Abs(denominator)
	doubleRemainder := new(big.Int).Mul(absRemainder, big.NewInt(2))
	comparison := doubleRemainder.Cmp(absDenominator)
	if comparison < 0 {
		return quotient
	}
	if comparison == 0 && quotient.Bit(0) == 0 {
		return quotient
	}

	if numerator.Sign()*denominator.Sign() >= 0 {
		return quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Sub(quotient, big.NewInt(1))
}
