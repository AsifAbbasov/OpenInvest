import type { ApiResult, AuthData, AuthSession, AuthUser } from "../../common/api/openinvest";

export type AuthState =
  | { status: "anonymous"; message: string | null }
  | { status: "authenticated"; user: AuthUser; session: AuthSession; message: string | null };

export type SessionOperation = "refresh" | "logout";

export type SessionOperationState = {
  generation: number;
  pending: SessionOperation | null;
};

export type SessionControllerRuntime = {
  getState: () => AuthState;
  getOperation: () => SessionOperationState;
  setOperation: (operation: SessionOperationState) => void;
  setPendingOperation: (operation: SessionOperation | null) => void;
  setState: (state: AuthState | ((state: AuthState) => AuthState)) => void;
};

export type SessionControllerClient = {
  refreshSession: (csrfToken: string) => Promise<ApiResult<AuthSession>>;
  logout: (payload: { allSessions: boolean }, csrfToken: string) => Promise<ApiResult<{ revoked: boolean }>>;
};

export function anonymousState(message: string | null = null): AuthState {
  return { status: "anonymous", message };
}

export function authenticatedState(data: AuthData): AuthState {
  return { status: "authenticated", user: data.user, session: data.session, message: null };
}

export function startSessionOperation(
  current: SessionOperationState,
  operation: SessionOperation,
): SessionOperationState {
  return { generation: current.generation + 1, pending: operation };
}

export function finishSessionOperation(
  current: SessionOperationState,
  generation: number,
): SessionOperationState {
  if (current.generation !== generation) {
    return current;
  }
  return { ...current, pending: null };
}

export function isCurrentOperation(current: SessionOperationState, generation: number) {
  return current.generation === generation;
}

export function applyAuthResult(result: ApiResult<AuthData>): AuthState {
  if (!result.ok) {
    return anonymousState(result.message);
  }
  return authenticatedState(result.data);
}

export function applyRefreshResult(
  state: AuthState,
  result: ApiResult<AuthSession>,
): AuthState {
  if (!result.ok) {
    return anonymousState("Session expired. Sign in again.");
  }
  if (state.status !== "authenticated") {
    return state;
  }
  return { ...state, session: result.data, message: "Session refreshed." };
}

export function beginSessionOperation(runtime: SessionControllerRuntime, operation: SessionOperation) {
  const next = startSessionOperation(runtime.getOperation(), operation);
  runtime.setOperation(next);
  runtime.setPendingOperation(next.pending);
  return next.generation;
}

export function completeSessionOperation(runtime: SessionControllerRuntime, generation: number) {
  const next = finishSessionOperation(runtime.getOperation(), generation);
  runtime.setOperation(next);
  runtime.setPendingOperation(next.pending);
}

export async function refreshActiveSession(
  runtime: SessionControllerRuntime,
  client: Pick<SessionControllerClient, "refreshSession">,
): Promise<ApiResult<AuthSession>> {
  if (runtime.getOperation().pending !== null) {
    return { ok: false, message: "Session operation already in progress." };
  }
  const currentState = runtime.getState();
  if (currentState.status !== "authenticated") {
    return { ok: false, message: "Sign in before refreshing the session." };
  }

  const generation = beginSessionOperation(runtime, "refresh");
  const result = await client.refreshSession(currentState.session.csrfToken);
  if (!isCurrentOperation(runtime.getOperation(), generation)) {
    return result;
  }
  completeSessionOperation(runtime, generation);
  runtime.setState((latestState) => applyRefreshResult(latestState, result));
  return result;
}

export async function logoutActiveSession(
  runtime: SessionControllerRuntime,
  client: Pick<SessionControllerClient, "logout">,
  allSessions = false,
): Promise<ApiResult<{ revoked: boolean }>> {
  const currentState = runtime.getState();
  if (currentState.status !== "authenticated") {
    return { ok: false, message: "No active session." };
  }

  const generation = beginSessionOperation(runtime, "logout");
  const result = await client.logout({ allSessions }, currentState.session.csrfToken);
  if (!isCurrentOperation(runtime.getOperation(), generation)) {
    return result;
  }
  completeSessionOperation(runtime, generation);
  runtime.setState(anonymousState(result.ok ? "Signed out." : result.message));
  return result;
}
