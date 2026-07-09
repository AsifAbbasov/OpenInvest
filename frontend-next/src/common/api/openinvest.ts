export type Money = {
  amount: string;
  currency: "RUB";
};

export type Portfolio = {
  id: string;
  name: string;
  baseCurrency: "RUB";
  version: number;
  createdAt: string;
  updatedAt: string;
};

export type PortfolioSummary = {
  portfolioId: string;
  asOfDate: string;
  totalValue: Money;
  cashValue: Money;
  stockValue: Money;
  bondValue: Money;
  investedCapital: Money;
  dividendsReceived: Money;
  couponsReceived: Money;
  nominalReturnRate: string;
  xirr: string | null;
  realReturn: {
    nominalReturnRate: string;
    inflationRate: string;
    realReturnRate: string;
    nominalGain: Money;
    realGain: Money;
    fromDate: string;
    toDate: string;
    methodologyVersion: string;
  };
  purchasingPower: {
    portfolioValue: Money;
    asOfDate: string;
    equivalents: unknown[];
  };
  positions: unknown[];
  calculation: {
    methodologyVersion: string;
    calculatedAt: string;
    inputsAsOf: string;
  };
};

export type TransactionType =
  | "BUY"
  | "SELL"
  | "DIVIDEND"
  | "COUPON"
  | "FEE"
  | "TAX"
  | "DEPOSIT"
  | "WITHDRAWAL";

export type Transaction = {
  id: string;
  portfolioId: string;
  transactionType: TransactionType;
  status: "ACTIVE" | "CORRECTED" | "REVERSED";
  ticker: string | null;
  quantity: string | null;
  unitPrice: Money | null;
  grossAmount: Money;
  commission: Money;
  tax: Money;
  tradeDate: string;
  settlementDate: string | null;
  note?: string | null;
  revision: number;
  createdAt: string;
  updatedAt: string;
};

export type CreatePortfolioPayload = {
  name: string;
  baseCurrency: "RUB";
};

export type CreateTransactionPayload = {
  transactionType: TransactionType;
  ticker: string | null;
  quantity: string | null;
  unitPrice: Money | null;
  grossAmount?: Money | null;
  commission: Money;
  tax: Money;
  tradeDate: string;
  settlementDate: string | null;
  note?: string | null;
};

export type ImportReviewStatus = "APPENDABLE" | "DUPLICATE" | "CONFLICT" | "INVALID";

export type ImportDecisionAction = "APPROVE" | "IGNORE" | "REJECT";

export type ImportSummary = {
  totalRows: number;
  appendableRows: number;
  duplicateRows: number;
  conflictRows: number;
  invalidRows: number;
};

export type ImportCandidate = {
  transactionType: TransactionType;
  ticker?: string;
  quantity?: string;
  unitPrice?: Money;
  grossAmount: Money;
  commission: Money;
  tax: Money;
  tradeDate: string;
  settlementDate?: string;
  safeNote?: string;
};

export type ImportRowReview = {
  rowNumber: number;
  rowHash: string;
  status: ImportReviewStatus;
  reasonCodes: string[];
  fingerprint?: string;
  candidate?: ImportCandidate;
};

export type ImportReviewResult = {
  portfolioId: string;
  sourceKind: "USER_UPLOADED_FILE";
  sourceAccountLabel: string;
  sourceFileHash: string;
  retentionPolicy: "TRANSIENT_NOT_STORED";
  reviewGuarantee: "PREFLIGHT_ONLY_APPEND_RERUNS_REVIEW_AND_STORE_CHECKS";
  summary: ImportSummary;
  rows: ImportRowReview[];
};

export type ImportDecision = {
  rowNumber: number;
  action: ImportDecisionAction;
};

export type ImportAppendResult = {
  portfolioId: string;
  sourceKind: "USER_UPLOADED_FILE";
  sourceFileHash: string;
  parsedRowCount: number;
  acceptedRowCount: number;
  nonAppendedRowCount: number;
  appendedTransactionIds: string[];
  snapshotDatesRebuilt: string[];
  auditActionCode: "IMPORT_APPEND_BATCH";
  nonSensitiveWarnings: string[];
  appendValidationPolicy: "REVIEW_RERUN_AND_ATOMIC_STORE_REVALIDATION";
  rawPayloadRetentionRule: "RAW_CSV_NOT_STORED";
};

export type ImportReviewPayload = {
  sourceAccountLabel?: string;
  csvPayload: string;
};

export type ImportAppendPayload = ImportReviewPayload & {
  decisions: ImportDecision[];
};

type Pagination = {
  nextCursor: string | null;
  hasMore: boolean;
  limit: number;
};

type BaseResponse<T> = {
  data: T;
  meta: {
    requestId: string;
    traceId: string;
    generatedAt: string;
  };
};

type ListData<T> = {
  items: T[];
  pagination: Pagination;
};

type ApiErrorResponse = {
  error?: {
    code?: string;
    message?: string;
  };
};

export type ApiResult<T> =
  | { ok: true; data: T; requestId: string }
  | { ok: false; message: string; status?: number };

const DEFAULT_API_BASE_URL = "http://localhost:8080";

export function openInvestApiBaseUrl() {
  return (process.env.NEXT_PUBLIC_OPENINVEST_API_BASE_URL ?? DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

export async function listPortfolios(): Promise<ApiResult<Portfolio[]>> {
  const response = await request<ListData<Portfolio>>("/api/v1/portfolios");
  if (!response.ok) {
    return response;
  }
  return { ok: true, data: response.data.items, requestId: response.requestId };
}

export async function createPortfolio(payload: CreatePortfolioPayload): Promise<ApiResult<Portfolio>> {
  return request<Portfolio>("/api/v1/portfolios", {
    method: "POST",
    headers: idempotentHeaders(),
    body: JSON.stringify(payload),
  });
}

export async function getPortfolio(portfolioId: string): Promise<ApiResult<Portfolio>> {
  return request<Portfolio>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}`);
}

export async function getPortfolioSummary(portfolioId: string): Promise<ApiResult<PortfolioSummary>> {
  return request<PortfolioSummary>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/summary`);
}

export async function listTransactions(portfolioId: string): Promise<ApiResult<Transaction[]>> {
  const response = await request<ListData<Transaction>>(
    `/api/v1/portfolios/${encodeURIComponent(portfolioId)}/transactions`,
  );
  if (!response.ok) {
    return response;
  }
  return { ok: true, data: response.data.items, requestId: response.requestId };
}

export async function appendTransaction(
  portfolioId: string,
  payload: CreateTransactionPayload,
): Promise<ApiResult<Transaction>> {
  return request<Transaction>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/transactions`, {
    method: "POST",
    headers: idempotentHeaders(),
    body: JSON.stringify(payload),
  });
}

export async function reviewPortfolioImport(
  portfolioId: string,
  payload: ImportReviewPayload,
): Promise<ApiResult<ImportReviewResult>> {
  return request<ImportReviewResult>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/imports/review`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function appendReviewedPortfolioImport(
  portfolioId: string,
  payload: ImportAppendPayload,
): Promise<ApiResult<ImportAppendResult>> {
  return request<ImportAppendResult>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/imports/append`, {
    method: "POST",
    headers: idempotentHeaders(),
    body: JSON.stringify(payload),
  });
}

async function request<T>(path: string, init: RequestInit = {}): Promise<ApiResult<T>> {
  try {
    const response = await fetch(`${openInvestApiBaseUrl()}${path}`, {
      ...init,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...init.headers,
      },
      cache: "no-store",
    });
    const payload = (await response.json()) as BaseResponse<T> | ApiErrorResponse;
    if (!response.ok) {
      return {
        ok: false,
        status: response.status,
        message: "error" in payload ? payload.error?.message ?? "Go API request failed" : "Go API request failed",
      };
    }
    if (!("data" in payload)) {
      return { ok: false, status: response.status, message: "Go API returned an unexpected response" };
    }
    return {
      ok: true,
      data: payload.data,
      requestId: payload.meta.requestId,
    };
  } catch {
    return { ok: false, message: "Go API is unavailable. Start backend-go and local PostgreSQL first." };
  }
}

function idempotentHeaders() {
  return {
    "Idempotency-Key": crypto.randomUUID(),
  };
}
