import type { ApiResult, AuthData, AuthSession } from "../../common/api/openinvest";

const SESSION_LOCK = "openinvest-auth-session-v1";
const SESSION_CHANNEL = "openinvest-auth-state-v1";
const SESSION_CHANGED = "session-changed";

// Web Locks provide the only trusted cross-tab mutex for the rotating refresh
// cookie. The per-tab queue below is deliberately limited to bootstrap/login:
// it is NOT a substitute for a cross-tab lock for refresh or logout.
let localQueue: Promise<void> = Promise.resolve();

export function hasCrossTabAuthLock(): boolean {
  return typeof navigator !== "undefined" &&
    typeof navigator.locks?.request === "function";
}

export function blockedWithoutCrossTabLock<T>(): ApiResult<T> {
  return {
    ok: false,
    failureKind: "REQUEST_SETUP",
    message: "Cross-tab session locking is unavailable. Refresh and sign out are disabled in this browser; use a browser supporting Web Locks.",
  };
}

// Refresh and logout MUST NOT fall back to localQueue. An uncoordinated
// refresh in two tabs may trigger the backend's refresh-token replay defense.
export async function withProtectedAuthLock<T>(
  operation: () => Promise<ApiResult<T>>,
): Promise<ApiResult<T>> {
  if (!hasCrossTabAuthLock()) return blockedWithoutCrossTabLock<T>();
  try {
    return await navigator.locks.request(SESSION_LOCK, { mode: "exclusive" }, operation);
  } catch {
    // The mutation may have reached the server; do not replay it or clear
    // the authenticated principal optimistically. A later bootstrap can
    // establish the current server-side session truth.
    return {
      ok: false,
      failureKind: "NETWORK",
      completion: "UNKNOWN",
      message: "Session operation outcome is uncertain. Reload to verify the session.",
    };
  }
}

export async function withAuthLock<T>(operation: () => Promise<T>): Promise<T> {
  if (hasCrossTabAuthLock()) {
    return navigator.locks.request(SESSION_LOCK, { mode: "exclusive" }, operation);
  }

  const previous = localQueue;
  let release!: () => void;
  localQueue = new Promise<void>((resolve) => {
    release = resolve;
  });
  await previous;
  try {
    return await operation();
  } finally {
    release();
  }
}

export function isSessionChangeMessage(message: unknown): boolean {
  return typeof message === "object" && message !== null &&
    "type" in message && message.type === SESSION_CHANGED;
}

// Only an invalidation signal crosses tabs. Never send user IDs, CSRF, access
// tokens, refresh tokens, account data or errors through BroadcastChannel.
export function announceAuthChange(): void {
  if (typeof BroadcastChannel === "undefined") return;
  const channel = new BroadcastChannel(SESSION_CHANNEL);
  try {
    channel.postMessage({ type: SESSION_CHANGED });
  } finally {
    channel.close();
  }
}

export function subscribeAuthChanges(onChange: () => void): () => void {
  if (typeof BroadcastChannel === "undefined") return () => {};
  const channel = new BroadcastChannel(SESSION_CHANNEL);
  channel.onmessage = (event: MessageEvent<unknown>) => {
    if (isSessionChangeMessage(event.data)) onChange();
  };
  return () => {
    channel.onmessage = null;
    channel.close();
  };
}

// Obtain the freshest CSRF under the same lock that protects refresh-token
// rotation. A stale React session must never supply the CSRF proof.
export async function refreshWithCurrentProof(
  expectedUserId: string,
  bootstrap: () => Promise<ApiResult<AuthData>>,
  refresh: (csrfToken: string) => Promise<ApiResult<AuthSession>>,
): Promise<ApiResult<AuthSession>> {
  const current = await bootstrap();
  if (!current.ok) return current;
  if (current.data.user.id !== expectedUserId) {
    return { ok: false, status: 401, message: "The signed-in account changed in another tab." };
  }
  return refresh(current.data.session.csrfToken);
}

export async function logoutWithCurrentProof(
  expectedUserId: string,
  payload: { allSessions: boolean },
  bootstrap: () => Promise<ApiResult<AuthData>>,
  logout: (payload: { allSessions: boolean }, csrf: string) => Promise<ApiResult<{ revoked: boolean }>>,
): Promise<ApiResult<{ revoked: boolean }>> {
  const current = await bootstrap();
  if (!current.ok) return current;
  if (current.data.user.id !== expectedUserId) {
    return { ok: false, status: 401, message: "The signed-in account changed in another tab." };
  }
  return logout(payload, current.data.session.csrfToken);
}
