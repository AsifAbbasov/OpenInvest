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

export type AuthUser = {
  id: string;
  email: string;
  language: string;
  theme: string;
  timezone: string;
  privacyMode: boolean;
  createdAt: string;
};

export type AuthSession = {
  accessToken: string;
  accessTokenExpiresAt: string;
  csrfToken: string;
};

export type AuthData = {
  user: AuthUser;
  session: AuthSession;
};

export type RegisterPayload = {
  email: string;
  password: string;
  language: "en";
  theme: "system";
  timezone: string;
};

export type LoginPayload = {
  email: string;
  password: string;
};

export type LogoutPayload = {
  allSessions: boolean;
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

export type AssetType = "STOCK" | "BOND";

export type AssetSummary = {
  ticker: string;
  name: string;
  assetType: AssetType;
  currency: "RUB";
  lotSize: string;
  lastPrice: Money | null;
};

type AssetStatus = "ACTIVE" | "INACTIVE" | "MATURED";

type SourceReference = {
  code: string;
  observedAt: string;
};

type AssetBase = AssetSummary & {
  market: "MOEX";
  status: AssetStatus;
  priceAsOf: string | null;
  source: SourceReference;
};

export type StockAsset = AssetBase & {
  assetType: "STOCK";
  stock: {
    sector: string;
    isin: string;
  };
};

export type BondAsset = AssetBase & {
  assetType: "BOND";
  bond: {
    isin: string;
    faceValue: Money;
    maturityDate: string;
    couponType: "FIXED" | "FLOATING" | "ZERO";
    couponRate?: string | null;
  };
};

export type Asset = StockAsset | BondAsset;

export type AssetSearchParams = {
  query: string;
  assetType?: AssetType;
  cursor?: string;
  limit?: number;
  signal?: AbortSignal;
};

export type Pagination = {
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

export type ListData<T> = {
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

export type AuthenticatedRequest = {
  accessToken: string;
};

export type CSRFAuthenticatedRequest = AuthenticatedRequest & {
  csrfToken: string;
};

const DEFAULT_API_BASE_URL = "http://localhost:8080";

export function openInvestApiBaseUrl() {
  return (process.env.NEXT_PUBLIC_OPENINVEST_API_BASE_URL ?? DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

export async function register(payload: RegisterPayload): Promise<ApiResult<AuthData>> {
  return request<AuthData>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
    credentials: "include",
  });
}

export async function login(payload: LoginPayload): Promise<ApiResult<AuthData>> {
  return request<AuthData>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
    credentials: "include",
  });
}

export async function refreshSession(csrfToken: string): Promise<ApiResult<AuthSession>> {
  return request<AuthSession>("/api/v1/auth/refresh", {
    method: "POST",
    headers: csrfHeaders(csrfToken),
    credentials: "include",
  });
}

export async function logout(payload: LogoutPayload, csrfToken: string): Promise<ApiResult<{ revoked: boolean }>> {
  return request<{ revoked: boolean }>("/api/v1/auth/logout", {
    method: "POST",
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(payload),
    credentials: "include",
  });
}

export async function listPortfolios(auth: AuthenticatedRequest): Promise<ApiResult<Portfolio[]>> {
  const response = await request<ListData<Portfolio>>("/api/v1/portfolios", {
    headers: bearerHeaders(auth.accessToken),
  });
  if (!response.ok) {
    return response;
  }
  return { ok: true, data: response.data.items, requestId: response.requestId };
}

export async function createPortfolio(
  payload: CreatePortfolioPayload,
  auth: AuthenticatedRequest,
): Promise<ApiResult<Portfolio>> {
  return request<Portfolio>("/api/v1/portfolios", {
    method: "POST",
    headers: { ...idempotentHeaders(), ...bearerHeaders(auth.accessToken) },
    body: JSON.stringify(payload),
  });
}

export async function getPortfolio(portfolioId: string, auth: AuthenticatedRequest): Promise<ApiResult<Portfolio>> {
  return request<Portfolio>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}`, {
    headers: bearerHeaders(auth.accessToken),
  });
}

export async function getPortfolioSummary(
  portfolioId: string,
  auth: AuthenticatedRequest,
): Promise<ApiResult<PortfolioSummary>> {
  return request<PortfolioSummary>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/summary`, {
    headers: bearerHeaders(auth.accessToken),
  });
}

export async function listTransactions(
  portfolioId: string,
  auth: AuthenticatedRequest,
): Promise<ApiResult<Transaction[]>> {
  const response = await request<ListData<Transaction>>(
    `/api/v1/portfolios/${encodeURIComponent(portfolioId)}/transactions`,
    { headers: bearerHeaders(auth.accessToken) },
  );
  if (!response.ok) {
    return response;
  }
  return { ok: true, data: response.data.items, requestId: response.requestId };
}

export async function appendTransaction(
  portfolioId: string,
  payload: CreateTransactionPayload,
  auth: AuthenticatedRequest,
): Promise<ApiResult<Transaction>> {
  return request<Transaction>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/transactions`, {
    method: "POST",
    headers: { ...idempotentHeaders(), ...bearerHeaders(auth.accessToken) },
    body: JSON.stringify(payload),
  });
}

export async function reviewPortfolioImport(
  portfolioId: string,
  payload: ImportReviewPayload,
  auth: AuthenticatedRequest,
): Promise<ApiResult<ImportReviewResult>> {
  return request<ImportReviewResult>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/imports/review`, {
    method: "POST",
    headers: bearerHeaders(auth.accessToken),
    body: JSON.stringify(payload),
  });
}

export async function appendReviewedPortfolioImport(
  portfolioId: string,
  payload: ImportAppendPayload,
  auth: AuthenticatedRequest,
): Promise<ApiResult<ImportAppendResult>> {
  return request<ImportAppendResult>(`/api/v1/portfolios/${encodeURIComponent(portfolioId)}/imports/append`, {
    method: "POST",
    headers: { ...idempotentHeaders(), ...bearerHeaders(auth.accessToken) },
    body: JSON.stringify(payload),
  });
}

export async function searchAssets(params: AssetSearchParams): Promise<ApiResult<ListData<AssetSummary>>> {
  const searchParams = new URLSearchParams({ query: params.query });
  if (params.assetType) {
    searchParams.set("assetType", params.assetType);
  }
  if (params.cursor) {
    searchParams.set("cursor", params.cursor);
  }
  if (params.limit) {
    searchParams.set("limit", String(params.limit));
  }
  return request<ListData<AssetSummary>>(`/api/v1/assets/search?${searchParams.toString()}`, {
    credentials: "omit",
    signal: params.signal,
  });
}

export async function getAsset(ticker: string, signal?: AbortSignal): Promise<ApiResult<Asset>> {
  return request<Asset>(`/api/v1/assets/${encodeURIComponent(ticker)}`, {
    credentials: "omit",
    signal,
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

function bearerHeaders(accessToken: string) {
  return {
    Authorization: `Bearer ${accessToken}`,
  };
}

function csrfHeaders(csrfToken: string) {
  return {
    "X-CSRF-Token": csrfToken,
  };
}
