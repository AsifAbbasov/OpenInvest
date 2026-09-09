package verticalslice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

const (
	PortfolioReturnMethodologyVersion = "portfolio-xirr-v1"
	maxAmbiguousXIRRTerms             = 160
	PortfolioReturnDayCountConvention = "ACT/365"

	PortfolioReturnAvailableStatus   = "AVAILABLE"
	PortfolioReturnUnavailableStatus = "UNAVAILABLE"

	PortfolioReturnReasonIncompleteValuation      = "INCOMPLETE_VALUATION"
	PortfolioReturnReasonNoExternalContributions  = "NO_EXTERNAL_CONTRIBUTIONS"
	PortfolioReturnReasonInsufficientDateSpan     = "INSUFFICIENT_DATE_SPAN"
	PortfolioReturnReasonNoSignChange             = "NO_SIGN_CHANGE"
	PortfolioReturnReasonNonPositiveTerminalValue = "NON_POSITIVE_TERMINAL_VALUE"
	PortfolioReturnReasonAmbiguousMultipleRoots   = "AMBIGUOUS_MULTIPLE_ROOTS"
	PortfolioReturnReasonNumericalSolutionFailed  = "NUMERICAL_SOLUTION_FAILED"
)

var ErrPortfolioReturnProjectionUnavailable = errors.New("portfolio return projection is unavailable")

type PortfolioReturnStore interface {
	GetPortfolioReturns(
		ctx context.Context,
		subjectID string,
		portfolioID string,
		asOfDate string,
	) (PortfolioReturnProjection, error)
}

type PortfolioReturnCashFlow struct {
	Date   string
	Amount decimal.Decimal
}

type PortfolioReturnProjection struct {
	PortfolioID            string
	AsOfDate               string
	Status                 string
	Reason                 string
	XIRR                   *decimal.Decimal
	ExternalCashFlows      []PortfolioReturnCashFlow
	TerminalPortfolioValue *Money
	MethodologyVersion     string
	DayCountConvention     string
}

func (s *Service) GetPortfolioReturns(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (PortfolioReturnProjection, error) {
	portfolioID = strings.TrimSpace(portfolioID)
	if portfolioID == "" {
		return PortfolioReturnProjection{}, ErrNotFound
	}
	if _, err := parseBusinessDate(asOfDate); err != nil {
		return PortfolioReturnProjection{}, fmt.Errorf("%w: asOfDate must be YYYY-MM-DD", ErrInvalidInput)
	}
	store, ok := s.store.(PortfolioReturnStore)
	if !ok {
		return PortfolioReturnProjection{}, ErrPortfolioReturnProjectionUnavailable
	}
	return store.GetPortfolioReturns(ctx, subjectID, portfolioID, asOfDate)
}

func BuildPortfolioReturnProjection(
	portfolioID string,
	asOfDate string,
	externalCashFlows []PortfolioReturnCashFlow,
	terminalPortfolioValue *Money,
	valuationComplete bool,
) (PortfolioReturnProjection, error) {
	asOf, err := parseBusinessDate(asOfDate)
	if err != nil {
		return PortfolioReturnProjection{}, fmt.Errorf("%w: asOfDate must be YYYY-MM-DD", ErrInvalidInput)
	}
	projection := PortfolioReturnProjection{
		PortfolioID:        portfolioID,
		AsOfDate:           asOfDate,
		Status:             PortfolioReturnUnavailableStatus,
		ExternalCashFlows:  []PortfolioReturnCashFlow{},
		MethodologyVersion: PortfolioReturnMethodologyVersion,
		DayCountConvention: PortfolioReturnDayCountConvention,
	}
	if !valuationComplete || terminalPortfolioValue == nil {
		projection.Reason = PortfolioReturnReasonIncompleteValuation
		return projection, nil
	}
	if terminalPortfolioValue.Currency != RUB {
		return PortfolioReturnProjection{}, fmt.Errorf("%w: terminal portfolio value must be RUB", ErrInvalidInput)
	}
	if !terminalPortfolioValue.Amount.FitsStorage() {
		return PortfolioReturnProjection{}, fmt.Errorf("%w: terminal portfolio value exceeds canonical Decimal constraints", ErrInvalidInput)
	}
	terminal := *terminalPortfolioValue
	projection.TerminalPortfolioValue = &terminal
	if !terminal.Amount.IsPositive() {
		projection.Reason = PortfolioReturnReasonNonPositiveTerminalValue
		return projection, nil
	}

	aggregated := map[string]decimal.Decimal{}
	hasContribution := false
	for _, flow := range externalCashFlows {
		date, err := parseBusinessDate(flow.Date)
		if err != nil {
			return PortfolioReturnProjection{}, fmt.Errorf("%w: external cash-flow date must be YYYY-MM-DD", ErrInvalidInput)
		}
		if date.After(asOf) {
			continue
		}
		if !flow.Amount.FitsStorage() {
			return PortfolioReturnProjection{}, fmt.Errorf("%w: external cash flow exceeds canonical Decimal constraints", ErrInvalidInput)
		}
		if flow.Amount.IsZero() {
			continue
		}
		if flow.Amount.IsNegative() {
			hasContribution = true
		}
		if current, ok := aggregated[flow.Date]; ok {
			combined := current.Add(flow.Amount)
			if !combined.FitsStorage() {
				return PortfolioReturnProjection{}, fmt.Errorf("%w: aggregated external cash flow exceeds canonical Decimal constraints", ErrInvalidInput)
			}
			aggregated[flow.Date] = combined
		} else {
			aggregated[flow.Date] = flow.Amount
		}
	}
	if !hasContribution {
		projection.Reason = PortfolioReturnReasonNoExternalContributions
		projection.ExternalCashFlows = sortedNonZeroReturnFlows(aggregated)
		return projection, nil
	}

	projection.ExternalCashFlows = sortedNonZeroReturnFlows(aggregated)
	if len(projection.ExternalCashFlows) == 0 {
		projection.Reason = PortfolioReturnReasonNoExternalContributions
		return projection, nil
	}

	combined := make([]PortfolioReturnCashFlow, 0, len(projection.ExternalCashFlows)+1)
	combined = append(combined, projection.ExternalCashFlows...)
	combined = append(combined, PortfolioReturnCashFlow{Date: asOfDate, Amount: terminal.Amount})
	combined = aggregateReturnFlows(combined)
	if len(combined) < 2 || combined[0].Date == combined[len(combined)-1].Date {
		projection.Reason = PortfolioReturnReasonInsufficientDateSpan
		return projection, nil
	}

	terms, err := xirrTerms(combined)
	if err != nil {
		return PortfolioReturnProjection{}, err
	}
	roots, err := isolateExponentialRoots(terms)
	if err != nil {
		projection.Reason = PortfolioReturnReasonNumericalSolutionFailed
		return projection, nil
	}
	if len(roots) == 0 {
		projection.Reason = PortfolioReturnReasonNoSignChange
		return projection, nil
	}
	if len(roots) > 1 {
		projection.Reason = PortfolioReturnReasonAmbiguousMultipleRoots
		return projection, nil
	}

	derivative := derivativeTerms(terms)
	if len(derivative) > 0 {
		value, scale, err := scaledExponentialValue(derivative, roots[0])
		if err != nil {
			projection.Reason = PortfolioReturnReasonNumericalSolutionFailed
			return projection, nil
		}
		if nearScaledZero(value, scale) {
			projection.Reason = PortfolioReturnReasonAmbiguousMultipleRoots
			return projection, nil
		}
	}

	rate := math.Expm1(roots[0])
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate <= -1 {
		projection.Reason = PortfolioReturnReasonNumericalSolutionFailed
		return projection, nil
	}
	quantized, err := quantizeXIRR(rate, terms, combined)
	if err != nil || quantized.String() == "-1.00000000" {
		projection.Reason = PortfolioReturnReasonNumericalSolutionFailed
		return projection, nil
	}
	projection.Status = PortfolioReturnAvailableStatus
	projection.Reason = ""
	projection.XIRR = &quantized
	return projection, nil
}

func sortedNonZeroReturnFlows(values map[string]decimal.Decimal) []PortfolioReturnCashFlow {
	dates := make([]string, 0, len(values))
	for date, amount := range values {
		if !amount.IsZero() {
			dates = append(dates, date)
		}
	}
	sort.Strings(dates)
	result := make([]PortfolioReturnCashFlow, 0, len(dates))
	for _, date := range dates {
		result = append(result, PortfolioReturnCashFlow{Date: date, Amount: values[date]})
	}
	return result
}

func aggregateReturnFlows(flows []PortfolioReturnCashFlow) []PortfolioReturnCashFlow {
	values := make(map[string]decimal.Decimal, len(flows))
	for _, flow := range flows {
		if current, ok := values[flow.Date]; ok {
			values[flow.Date] = current.Add(flow.Amount)
		} else {
			values[flow.Date] = flow.Amount
		}
	}
	return sortedNonZeroReturnFlows(values)
}

type xirrTerm struct {
	years       float64
	coefficient float64
}

func xirrTerms(flows []PortfolioReturnCashFlow) ([]xirrTerm, error) {
	if len(flows) == 0 {
		return nil, nil
	}
	first, err := parseBusinessDate(flows[0].Date)
	if err != nil {
		return nil, err
	}
	firstDay := first.Unix() / 86400
	terms := make([]xirrTerm, 0, len(flows))
	for _, flow := range flows {
		date, err := parseBusinessDate(flow.Date)
		if err != nil {
			return nil, err
		}
		days := date.Unix()/86400 - firstDay
		amount, err := strconv.ParseFloat(flow.Amount.String(), 64)
		if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
			return nil, errors.New("portfolio XIRR cash flow is not representable as float64")
		}
		terms = append(terms, xirrTerm{years: float64(days) / 365.0, coefficient: amount})
	}
	return normalizeExponentialTerms(terms), nil
}

func parseBusinessDate(value string) (time.Time, error) {
	if len(value) != len("2006-01-02") {
		return time.Time{}, errors.New("invalid BusinessDate")
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, errors.New("invalid BusinessDate")
	}
	return parsed, nil
}

func isolateExponentialRoots(raw []xirrTerm) ([]float64, error) {
	terms := normalizeExponentialTerms(raw)
	if len(terms) <= 1 {
		return nil, nil
	}
	signChanges := exponentialCoefficientSignChanges(terms)
	if signChanges == 0 {
		return nil, nil
	}
	if len(terms) == 2 {
		left := terms[0]
		right := terms[1]
		if left.coefficient == 0 || right.coefficient == 0 || math.Signbit(left.coefficient) == math.Signbit(right.coefficient) {
			return nil, nil
		}
		logRatio := math.Log(math.Abs(left.coefficient)) - math.Log(math.Abs(right.coefficient))
		if math.IsNaN(logRatio) || math.IsInf(logRatio, 0) {
			return nil, errors.New("invalid two-term XIRR log ratio")
		}
		delta := right.years - left.years
		if !(delta > 0) {
			return nil, errors.New("non-increasing XIRR exponents")
		}
		root := -logRatio / delta
		if math.IsNaN(root) || math.IsInf(root, 0) {
			return nil, errors.New("non-finite two-term XIRR root")
		}
		return []float64{root}, nil
	}

	if signChanges == 1 {
		lower, upper, err := exponentialDominanceBounds(terms)
		if err != nil {
			return nil, err
		}
		lowerValue, _, err := scaledExponentialValue(terms, lower)
		if err != nil {
			return nil, err
		}
		upperValue, _, err := scaledExponentialValue(terms, upper)
		if err != nil {
			return nil, err
		}
		if lowerValue == 0 {
			return []float64{lower}, nil
		}
		if upperValue == 0 {
			return []float64{upper}, nil
		}
		if math.Signbit(lowerValue) == math.Signbit(upperValue) {
			return nil, errors.New("one-sign-change XIRR bounds do not bracket a root")
		}
		root, err := bisectExponentialRoot(terms, lower, upper, lowerValue)
		if err != nil {
			return nil, err
		}
		return []float64{root}, nil
	}

	// Multi-sign-change cash-flow series require derivative-recursive isolation and can
	// become super-linear/adversarial with hundreds of alternating dates. Ordinary
	// contribution histories use the one-sign-change fast path above and are not capped.
	// Fail closed rather than allowing one request to monopolize the API process.
	if len(terms) > maxAmbiguousXIRRTerms {
		return nil, errors.New("ambiguous XIRR series exceeds safe root-isolation complexity")
	}

	lower, upper, err := exponentialDominanceBounds(terms)
	if err != nil {
		return nil, err
	}
	critical, err := isolateExponentialRoots(derivativeTerms(terms))
	if err != nil {
		return nil, err
	}
	points := make([]float64, 0, len(critical)+2)
	points = append(points, lower)
	for _, point := range critical {
		if point > lower && point < upper {
			points = append(points, point)
		}
	}
	points = append(points, upper)
	sort.Float64s(points)
	points = dedupeRootCandidates(points)

	roots := make([]float64, 0, len(points))
	for _, point := range points[1 : len(points)-1] {
		value, scale, err := scaledExponentialValue(terms, point)
		if err != nil {
			return nil, err
		}
		if nearScaledZero(value, scale) {
			roots = append(roots, point)
		}
	}
	for index := 0; index+1 < len(points); index++ {
		left := points[index]
		right := points[index+1]
		leftValue, _, err := scaledExponentialValue(terms, left)
		if err != nil {
			return nil, err
		}
		rightValue, _, err := scaledExponentialValue(terms, right)
		if err != nil {
			return nil, err
		}
		if leftValue == 0 || rightValue == 0 || math.Signbit(leftValue) == math.Signbit(rightValue) {
			continue
		}
		root, err := bisectExponentialRoot(terms, left, right, leftValue)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}
	return dedupeRootCandidates(roots), nil
}

func exponentialCoefficientSignChanges(terms []xirrTerm) int {
	changes := 0
	previousSign := 0
	for _, term := range terms {
		sign := 0
		if term.coefficient > 0 {
			sign = 1
		} else if term.coefficient < 0 {
			sign = -1
		}
		if sign == 0 {
			continue
		}
		if previousSign != 0 && sign != previousSign {
			changes++
		}
		previousSign = sign
	}
	return changes
}

func derivativeTerms(terms []xirrTerm) []xirrTerm {
	result := make([]xirrTerm, 0, len(terms)-1)
	for _, term := range terms {
		coefficient := -term.years * term.coefficient
		if coefficient != 0 {
			result = append(result, xirrTerm{years: term.years, coefficient: coefficient})
		}
	}
	return normalizeExponentialTerms(result)
}

func normalizeExponentialTerms(terms []xirrTerm) []xirrTerm {
	filtered := make([]xirrTerm, 0, len(terms))
	for _, term := range terms {
		if term.coefficient == 0 {
			continue
		}
		filtered = append(filtered, term)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].years < filtered[j].years })
	merged := make([]xirrTerm, 0, len(filtered))
	for _, term := range filtered {
		if len(merged) > 0 && merged[len(merged)-1].years == term.years {
			merged[len(merged)-1].coefficient += term.coefficient
			if merged[len(merged)-1].coefficient == 0 {
				merged = merged[:len(merged)-1]
			}
			continue
		}
		merged = append(merged, term)
	}
	if len(merged) > 0 {
		minimumYears := merged[0].years
		for index := range merged {
			merged[index].years -= minimumYears
		}
	}
	maxCoefficient := 0.0
	for _, term := range merged {
		maxCoefficient = math.Max(maxCoefficient, math.Abs(term.coefficient))
	}
	if maxCoefficient == 0 || math.IsNaN(maxCoefficient) || math.IsInf(maxCoefficient, 0) {
		return merged
	}
	for index := range merged {
		merged[index].coefficient /= maxCoefficient
	}
	return merged
}

func exponentialDominanceBounds(terms []xirrTerm) (float64, float64, error) {
	if len(terms) < 2 {
		return 0, 0, errors.New("at least two XIRR terms are required for root bounds")
	}
	first := terms[0]
	last := terms[len(terms)-1]
	upperDelta := terms[1].years - first.years
	lowerDelta := last.years - terms[len(terms)-2].years
	if !(upperDelta > 0) || !(lowerDelta > 0) || first.coefficient == 0 || last.coefficient == 0 {
		return 0, 0, errors.New("invalid XIRR root-bound terms")
	}
	upperRest := 0.0
	for _, term := range terms[1:] {
		upperRest += math.Abs(term.coefficient)
	}
	lowerRest := 0.0
	for _, term := range terms[:len(terms)-1] {
		lowerRest += math.Abs(term.coefficient)
	}
	upper := 0.0
	if upperRest > 0 {
		candidate := (math.Log(upperRest) - math.Log(math.Abs(first.coefficient)) + math.Log(2)) / upperDelta
		if candidate > upper {
			upper = candidate
		}
	}
	lower := 0.0
	if lowerRest > 0 {
		candidate := -(math.Log(lowerRest) - math.Log(math.Abs(last.coefficient)) + math.Log(2)) / lowerDelta
		if candidate < lower {
			lower = candidate
		}
	}
	if math.IsNaN(lower) || math.IsNaN(upper) || math.IsInf(lower, 0) || math.IsInf(upper, 0) || !(lower < upper) {
		return 0, 0, errors.New("non-finite XIRR root bounds")
	}
	return lower, upper, nil
}

func scaledExponentialValue(terms []xirrTerm, y float64) (float64, float64, error) {
	if len(terms) == 0 || math.IsNaN(y) || math.IsInf(y, 0) {
		return 0, 0, errors.New("invalid XIRR evaluation input")
	}
	maxLogMagnitude := math.Inf(-1)
	for _, term := range terms {
		if term.coefficient == 0 {
			continue
		}
		logMagnitude := math.Log(math.Abs(term.coefficient)) - term.years*y
		if math.IsNaN(logMagnitude) || math.IsInf(logMagnitude, 1) {
			return 0, 0, errors.New("non-finite XIRR log magnitude")
		}
		if logMagnitude > maxLogMagnitude {
			maxLogMagnitude = logMagnitude
		}
	}
	if math.IsInf(maxLogMagnitude, -1) {
		return 0, 0, errors.New("XIRR evaluation has no non-zero terms")
	}

	sum := 0.0
	compensation := 0.0
	absoluteSum := 0.0
	for _, term := range terms {
		if term.coefficient == 0 {
			continue
		}
		logMagnitude := math.Log(math.Abs(term.coefficient)) - term.years*y
		value := math.Exp(logMagnitude - maxLogMagnitude)
		if math.Signbit(term.coefficient) {
			value = -value
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, 0, errors.New("non-finite XIRR term")
		}
		adjusted := value - compensation
		next := sum + adjusted
		compensation = (next - sum) - adjusted
		sum = next
		absoluteSum += math.Abs(value)
	}
	if math.IsNaN(sum) || math.IsInf(sum, 0) || math.IsNaN(absoluteSum) || math.IsInf(absoluteSum, 0) {
		return 0, 0, errors.New("non-finite XIRR evaluation")
	}
	return sum, absoluteSum, nil
}

func nearScaledZero(value float64, absoluteSum float64) bool {
	if value == 0 {
		return true
	}
	if absoluteSum == 0 {
		return false
	}
	return math.Abs(value) <= 1e-12*absoluteSum
}

func bisectExponentialRoot(terms []xirrTerm, left, right, leftValue float64) (float64, error) {
	if !(left < right) || leftValue == 0 {
		return 0, errors.New("invalid XIRR bisection bracket")
	}
	for iteration := 0; iteration < 256; iteration++ {
		midpoint := left + (right-left)/2
		if midpoint == left || midpoint == right {
			return midpoint, nil
		}
		value, _, err := scaledExponentialValue(terms, midpoint)
		if err != nil {
			return 0, err
		}
		if value == 0 {
			return midpoint, nil
		}
		if math.Signbit(value) == math.Signbit(leftValue) {
			left = midpoint
			leftValue = value
		} else {
			right = midpoint
		}
	}
	return left + (right-left)/2, nil
}

func dedupeRootCandidates(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	sort.Float64s(values)
	result := []float64{values[0]}
	for _, value := range values[1:] {
		previous := result[len(result)-1]
		threshold := 1e-13 * (1 + math.Max(math.Abs(previous), math.Abs(value)))
		if math.Abs(value-previous) <= threshold {
			result[len(result)-1] = previous + (value-previous)/2
			continue
		}
		result = append(result, value)
	}
	return result
}

func quantizeXIRR(rate float64, terms []xirrTerm, flows []PortfolioReturnCashFlow) (decimal.Decimal, error) {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate <= -1 {
		return decimal.Zero(), errors.New("invalid XIRR rate")
	}
	// One local ULP must remain comfortably smaller than the scale-8 output quantum.
	// The additional uncertainty band below accounts for bisection termination and Expm1.
	nextGap := math.Abs(math.Nextafter(rate, math.Inf(1)) - rate)
	previousGap := math.Abs(rate - math.Nextafter(rate, math.Inf(-1)))
	maxGap := math.Max(nextGap, previousGap)
	if maxGap >= 1e-8/128 {
		return decimal.Zero(), errors.New("XIRR rate exceeds reliable scale-8 float resolution")
	}
	scaled := rate * 1e8
	if math.IsNaN(scaled) || math.IsInf(scaled, 0) {
		return decimal.Zero(), errors.New("XIRR rate exceeds Decimal range")
	}
	// Above 2^53, float64 cannot represent every scale-8 integer. Publishing a rounded
	// Decimal there could claim precision the numerical representation does not have.
	if math.Abs(scaled) >= 1<<53 {
		return decimal.Zero(), errors.New("XIRR rate exceeds reliable scale-8 float quantization")
	}

	lowerInteger := math.Floor(scaled)
	halfBoundary := lowerInteger + 0.5
	uncertaintyUnits := 16*maxGap*1e8 + 1e-12
	if uncertaintyUnits >= 0.5 {
		return decimal.Zero(), errors.New("XIRR scale-8 rounding interval is numerically unresolved")
	}

	rounded := math.RoundToEven(scaled)
	if math.Abs(scaled-halfBoundary) <= uncertaintyUnits {
		lowerUnits := int64(lowerInteger)
		exactSign, exact, err := exactPolynomialMidpointSign(flows, lowerUnits)
		if err != nil {
			return decimal.Zero(), err
		}
		if exact {
			if exactSign == 0 {
				if lowerUnits%2 == 0 {
					rounded = float64(lowerUnits)
				} else {
					rounded = float64(lowerUnits + 1)
				}
			} else {
				midpointRate := (float64(lowerUnits) + 0.5) / 1e8
				midpointY := math.Log1p(midpointRate)
				derivative := derivativeTerms(terms)
				derivativeValue, derivativeScale, err := scaledExponentialValue(derivative, midpointY)
				if err != nil || derivativeScale == 0 || nearScaledZero(derivativeValue, derivativeScale) {
					return decimal.Zero(), errors.New("XIRR midpoint rounding side is numerically unresolved")
				}
				if -float64(exactSign)/derivativeValue > 0 {
					rounded = float64(lowerUnits + 1)
				} else {
					rounded = float64(lowerUnits)
				}
			}
		} else {
			// Irregular ACT/365 exponents are generalized exponential terms rather than an
			// ordinary rational polynomial. Resolve a clearly separated midpoint numerically,
			// but never guess when the root is within the evaluator's error envelope.
			midpointRate := (float64(lowerUnits) + 0.5) / 1e8
			if midpointRate <= -1 {
				return decimal.Zero(), errors.New("XIRR midpoint is outside admissible domain")
			}
			midpointY := math.Log1p(midpointRate)
			value, absoluteSum, err := scaledExponentialValue(terms, midpointY)
			if err != nil {
				return decimal.Zero(), err
			}
			epsilon := math.Nextafter(1, 2) - 1
			if math.Abs(value) <= 64*epsilon*math.Max(absoluteSum, 1) {
				return decimal.Zero(), errors.New("irregular XIRR midpoint is numerically unresolved")
			}
			derivative := derivativeTerms(terms)
			derivativeValue, derivativeScale, err := scaledExponentialValue(derivative, midpointY)
			if err != nil || derivativeScale == 0 || nearScaledZero(derivativeValue, derivativeScale) {
				return decimal.Zero(), errors.New("XIRR midpoint rounding side is numerically unresolved")
			}
			if -value/derivativeValue > 0 {
				rounded = float64(lowerUnits + 1)
			} else {
				rounded = float64(lowerUnits)
			}
		}
	}
	if math.IsNaN(rounded) || math.IsInf(rounded, 0) {
		return decimal.Zero(), errors.New("XIRR rounding failed")
	}
	units := int64(rounded)
	absUnits := units
	sign := ""
	if absUnits < 0 {
		sign = "-"
		absUnits = -absUnits
	}
	text := fmt.Sprintf("%s%d.%08d", sign, absUnits/100000000, absUnits%100000000)
	value, err := decimal.FromString(text)
	if err != nil || !value.FitsStorage() {
		return decimal.Zero(), errors.New("XIRR rate exceeds canonical Decimal constraints")
	}
	return value, nil
}

// exactPolynomialMidpointSign evaluates the XIRR equation exactly at the scale-8
// rounding midpoint when every ACT/365 exponent is an integer year. Multiplying the
// equation by q^maxYear, where q=1+r>0, turns it into a rational polynomial without
// changing its sign. This distinguishes exact Half-Even ties from roots that are only
// infinitesimally above or below the midpoint but collapse to the same float64.
func exactPolynomialMidpointSign(flows []PortfolioReturnCashFlow, lowerUnits int64) (int, bool, error) {
	if len(flows) == 0 {
		return 0, false, nil
	}
	first, err := parseBusinessDate(flows[0].Date)
	if err != nil {
		return 0, false, err
	}
	firstDay := first.Unix() / 86400
	years := make([]int64, len(flows))
	maxYear := int64(0)
	for index, flow := range flows {
		date, err := parseBusinessDate(flow.Date)
		if err != nil {
			return 0, false, err
		}
		days := date.Unix()/86400 - firstDay
		if days < 0 || days%365 != 0 {
			return 0, false, nil
		}
		year := days / 365
		years[index] = year
		if year > maxYear {
			maxYear = year
		}
	}

	denominator := big.NewInt(200000000)
	numerator := new(big.Int).Add(
		new(big.Int).Set(denominator),
		big.NewInt(2*lowerUnits+1),
	)
	if numerator.Sign() <= 0 {
		return 0, false, errors.New("XIRR midpoint is outside admissible domain")
	}
	q := new(big.Rat).SetFrac(numerator, denominator)
	sum := new(big.Rat)
	for index, flow := range flows {
		coefficient, ok := new(big.Rat).SetString(flow.Amount.String())
		if !ok {
			return 0, false, errors.New("invalid canonical Decimal in XIRR midpoint evaluation")
		}
		power := maxYear - years[index]
		factor := ratPow(q, power)
		term := new(big.Rat).Mul(coefficient, factor)
		sum.Add(sum, term)
	}
	return sum.Sign(), true, nil
}

func ratPow(base *big.Rat, exponent int64) *big.Rat {
	result := new(big.Rat).SetInt64(1)
	factor := new(big.Rat).Set(base)
	for exponent > 0 {
		if exponent&1 == 1 {
			result.Mul(result, factor)
		}
		exponent >>= 1
		if exponent > 0 {
			factor.Mul(factor, factor)
		}
	}
	return result
}
