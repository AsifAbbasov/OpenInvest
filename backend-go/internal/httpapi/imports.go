package httpapi

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/importflow"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"net/http"
	"strings"
	"unicode/utf8"
)

func (api *API) reviewImport(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request importReviewRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	portfolioID := c.Params("portfolioId")
	fileHash := importPayloadHash(request.CSVPayload)

	parserReview, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := validateImportRowCount(parserReview.Summary.TotalRows); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	existing, err := api.service.ListImportReviewTransactions(
		c.Context(),
		subjectID,
		portfolioID,
		importer.ReviewHistoryFilter(parserReview),
	)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	review, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Existing:           existing,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	reviewToken, err := api.signImportReviewToken(subjectID, parserReview, review)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusOK, mapImportReview(review, reviewToken))
}

func (api *API) appendImport(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := verticalslice.ValidateIdempotencyKey(c.Get("Idempotency-Key")); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request importAppendRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	portfolioID := c.Params("portfolioId")
	fileHash := importPayloadHash(request.CSVPayload)
	if request.SourceFileHash != fileHash {
		return writeMappedErrorWithMeta(c, meta, fmt.Errorf("%w: sourceFileHash does not match import payload", importer.ErrUnsafeAppend))
	}
	preflightReview, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := validateImportRowCount(preflightReview.Summary.TotalRows); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := api.verifyImportReviewToken(
		request.ReviewToken,
		subjectID,
		portfolioID,
		importer.SourceKindUserUploadedFile,
		request.SourceAccountLabel,
		fileHash,
		preflightReview,
		request.toAppDecisions(),
	); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := importer.VerifyDecisionIdentities(preflightReview, request.toAppDecisionIdentities()); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	result, err := importflow.ReviewAndAppend(c.Context(), api.service, importflow.Request{
		RequestContext:     meta.toApp(),
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		IdempotencyKey:     c.Get("Idempotency-Key"),
		RequestPath:        c.Path(),
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		SourceFileHash:     fileHash,
		// The store performs the atomic duplicate check after reserving the command. Passing the
		// mutable current ledger here would reject an idempotent retry before that reservation.
		Existing:  nil,
		Reader:    strings.NewReader(request.CSVPayload),
		Decisions: request.toAppDecisions(),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapImportAppendResult(portfolioID, importer.SourceKindUserUploadedFile, fileHash, result))
}

type importReviewRequestDTO struct {
	SourceAccountLabel string `json:"sourceAccountLabel"`
	CSVPayload         string `json:"csvPayload"`
}

func (request importReviewRequestDTO) validate() error {
	if strings.TrimSpace(request.CSVPayload) == "" {
		return fmt.Errorf("%w: csvPayload is required", verticalslice.ErrInvalidInput)
	}
	if len([]byte(request.CSVPayload)) > maxHTTPImportPayloadBytes {
		return fmt.Errorf("%w: csvPayload exceeds 2097152 bytes", verticalslice.ErrInvalidInput)
	}
	if !utf8.ValidString(request.SourceAccountLabel) {
		return fmt.Errorf("%w: sourceAccountLabel must be valid UTF-8", verticalslice.ErrInvalidInput)
	}
	if utf8.RuneCountInString(request.SourceAccountLabel) > 120 {
		return fmt.Errorf("%w: sourceAccountLabel must be at most 120 characters", verticalslice.ErrInvalidInput)
	}
	return nil
}

type importAppendRequestDTO struct {
	SourceAccountLabel string              `json:"sourceAccountLabel"`
	SourceFileHash     string              `json:"sourceFileHash"`
	ReviewToken        string              `json:"reviewToken"`
	CSVPayload         string              `json:"csvPayload"`
	Decisions          []importDecisionDTO `json:"decisions"`
}

func (request importAppendRequestDTO) validate() error {
	if err := (importReviewRequestDTO{
		SourceAccountLabel: request.SourceAccountLabel,
		CSVPayload:         request.CSVPayload,
	}).validate(); err != nil {
		return err
	}
	if len(request.Decisions) == 0 {
		return fmt.Errorf("%w: decisions are required", verticalslice.ErrInvalidInput)
	}
	if len(request.Decisions) > 100 {
		return fmt.Errorf("%w: decisions must contain at most 100 rows", verticalslice.ErrInvalidInput)
	}
	if !validSHA256Hex(request.SourceFileHash) {
		return fmt.Errorf("%w: sourceFileHash must be a SHA-256 hex digest from the review response", verticalslice.ErrInvalidInput)
	}
	if len(request.ReviewToken) < 32 || len(request.ReviewToken) > maxImportReviewTokenBytes {
		return fmt.Errorf("%w: reviewToken must be between 32 and %d bytes", verticalslice.ErrInvalidInput, maxImportReviewTokenBytes)
	}
	for _, decision := range request.Decisions {
		if !validSHA256Hex(decision.RowHash) {
			return fmt.Errorf("%w: rowHash must be a SHA-256 hex digest from the review response", verticalslice.ErrInvalidInput)
		}
	}
	return nil
}

func (request importAppendRequestDTO) toAppDecisions() []importer.Decision {
	decisions := make([]importer.Decision, 0, len(request.Decisions))
	for _, decision := range request.Decisions {
		decisions = append(decisions, importer.Decision{RowNumber: decision.RowNumber, RowHash: decision.RowHash, Action: decision.Action})
	}
	return decisions
}

func (request importAppendRequestDTO) toAppDecisionIdentities() []importer.DecisionIdentity {
	identities := make([]importer.DecisionIdentity, 0, len(request.Decisions))
	for _, decision := range request.Decisions {
		identities = append(identities, importer.DecisionIdentity{RowNumber: decision.RowNumber, RowHash: decision.RowHash})
	}
	return identities
}

type importDecisionDTO struct {
	RowNumber int    `json:"rowNumber"`
	RowHash   string `json:"rowHash"`
	Action    string `json:"action"`
}

type importReviewDTO struct {
	PortfolioID        string               `json:"portfolioId"`
	SourceKind         string               `json:"sourceKind"`
	SourceAccountLabel string               `json:"sourceAccountLabel"`
	SourceFileHash     string               `json:"sourceFileHash"`
	ReviewToken        string               `json:"reviewToken"`
	RetentionPolicy    string               `json:"retentionPolicy"`
	ReviewGuarantee    string               `json:"reviewGuarantee"`
	Summary            importSummaryDTO     `json:"summary"`
	Rows               []importRowReviewDTO `json:"rows"`
}

type importSummaryDTO struct {
	TotalRows      int `json:"totalRows"`
	AppendableRows int `json:"appendableRows"`
	DuplicateRows  int `json:"duplicateRows"`
	ConflictRows   int `json:"conflictRows"`
	InvalidRows    int `json:"invalidRows"`
}

type importRowReviewDTO struct {
	RowNumber   int                 `json:"rowNumber"`
	RowHash     string              `json:"rowHash"`
	Status      string              `json:"status"`
	ReasonCodes []string            `json:"reasonCodes"`
	Fingerprint string              `json:"fingerprint,omitempty"`
	Candidate   *importCandidateDTO `json:"candidate,omitempty"`
}

type importCandidateDTO struct {
	TransactionType string    `json:"transactionType"`
	Ticker          *string   `json:"ticker,omitempty"`
	Quantity        *string   `json:"quantity,omitempty"`
	UnitPrice       *moneyDTO `json:"unitPrice,omitempty"`
	GrossAmount     moneyDTO  `json:"grossAmount"`
	Commission      moneyDTO  `json:"commission"`
	Tax             moneyDTO  `json:"tax"`
	TradeDate       string    `json:"tradeDate"`
	SettlementDate  *string   `json:"settlementDate,omitempty"`
	SafeNote        *string   `json:"safeNote,omitempty"`
}

type importAppendResultDTO struct {
	PortfolioID             string   `json:"portfolioId"`
	SourceKind              string   `json:"sourceKind"`
	SourceFileHash          string   `json:"sourceFileHash"`
	ParsedRowCount          int      `json:"parsedRowCount"`
	AcceptedRowCount        int      `json:"acceptedRowCount"`
	NonAppendedRowCount     int      `json:"nonAppendedRowCount"`
	AppendedTransactionIDs  []string `json:"appendedTransactionIds"`
	SnapshotDatesRebuilt    []string `json:"snapshotDatesRebuilt"`
	AuditActionCode         string   `json:"auditActionCode"`
	NonSensitiveWarnings    []string `json:"nonSensitiveWarnings"`
	AppendValidationPolicy  string   `json:"appendValidationPolicy"`
	RawPayloadRetentionRule string   `json:"rawPayloadRetentionRule"`
}

func mapImportReview(review importer.Review, reviewToken string) importReviewDTO {
	rows := make([]importRowReviewDTO, 0, len(review.Rows))
	for _, row := range review.Rows {
		rows = append(rows, mapImportRowReview(row))
	}
	return importReviewDTO{
		PortfolioID:        review.PortfolioID,
		SourceKind:         review.SourceKind,
		SourceAccountLabel: review.SourceAccountLabel,
		SourceFileHash:     review.FileHash,
		ReviewToken:        reviewToken,
		RetentionPolicy:    "TRANSIENT_NOT_STORED",
		ReviewGuarantee:    "SIGNED_REVIEW_TOKEN_APPEND_RERUNS_REVIEW_AND_STORE_CHECKS",
		Summary: importSummaryDTO{
			TotalRows:      review.Summary.TotalRows,
			AppendableRows: review.Summary.AppendableRows,
			DuplicateRows:  review.Summary.DuplicateRows,
			ConflictRows:   review.Summary.ConflictRows,
			InvalidRows:    review.Summary.InvalidRows,
		},
		Rows: rows,
	}
}

func mapImportRowReview(row importer.RowReview) importRowReviewDTO {
	return importRowReviewDTO{
		RowNumber:   row.RowNumber,
		RowHash:     row.RowHash,
		Status:      row.Status,
		ReasonCodes: row.ReasonCodes,
		Fingerprint: row.Fingerprint,
		Candidate:   mapImportCandidate(row.Candidate),
	}
}

func mapImportCandidate(candidate *importer.Candidate) *importCandidateDTO {
	if candidate == nil {
		return nil
	}
	var quantity *string
	if candidate.Quantity != nil {
		value := candidate.Quantity.String()
		quantity = &value
	}
	return &importCandidateDTO{
		TransactionType: candidate.TransactionType,
		Ticker:          candidate.Ticker,
		Quantity:        quantity,
		UnitPrice:       mapOptionalMoney(candidate.UnitPrice),
		GrossAmount:     mapMoney(candidate.GrossAmount),
		Commission:      mapMoney(candidate.Commission),
		Tax:             mapMoney(candidate.Tax),
		TradeDate:       candidate.TradeDate,
		SettlementDate:  candidate.SettlementDate,
		SafeNote:        candidate.SafeNote,
	}
}

func mapImportAppendResult(portfolioID string, sourceKind string, fileHash string, result importflow.Result) importAppendResultDTO {
	return importAppendResultDTO{
		PortfolioID:             portfolioID,
		SourceKind:              sourceKind,
		SourceFileHash:          fileHash,
		ParsedRowCount:          result.ParsedRowCount,
		AcceptedRowCount:        result.AcceptedRowCount,
		NonAppendedRowCount:     result.NonAppendedRowCount,
		AppendedTransactionIDs:  result.AppendedTransactionIDs,
		SnapshotDatesRebuilt:    result.SnapshotDatesRebuilt,
		AuditActionCode:         result.AuditActionCode,
		NonSensitiveWarnings:    result.NonSensitiveWarnings,
		AppendValidationPolicy:  "REVIEW_RERUN_AND_ATOMIC_STORE_REVALIDATION",
		RawPayloadRetentionRule: "RAW_CSV_NOT_STORED",
	}
}

type importReviewTokenRowIdentity struct {
	RowNumber int    `json:"n"`
	RowHash   string `json:"h"`
}

type importReviewTokenPayload struct {
	Version            int                            `json:"v"`
	ParserVersion      int                            `json:"pv"`
	IssuedAt           int64                          `json:"iat"`
	ExpiresAt          int64                          `json:"exp"`
	SubjectID          string                         `json:"s"`
	PortfolioID        string                         `json:"p"`
	SourceKind         string                         `json:"k"`
	SourceAccountLabel string                         `json:"l"`
	SourceFileHash     string                         `json:"f"`
	ParserReviewDigest string                         `json:"pd"`
	FinalReviewDigest  string                         `json:"fd"`
	AppendableRows     []int                          `json:"a"`
	Rows               []importReviewTokenRowIdentity `json:"r"`
}
