export type IdempotencyIntentState = {
  intent: string | null;
  key: string | null;
};

export type IdempotencyIntentStorage = Pick<Storage, "getItem" | "setItem" | "removeItem">;

export const emptyIdempotencyIntent: IdempotencyIntentState = {
  intent: null,
  key: null,
};

const browserRetryIntentVersion = 1;
const browserRetryIntentTTLMilliseconds = 24 * 60 * 60 * 1000;
const browserRetryIntentPrefix = "oi:idempotency-retry:v1:";
const idempotencyKeyPattern = /^[A-Za-z0-9._:-]{16,128}$/;

type StoredRetryIntent = {
  version: number;
  key: string;
  expiresAt: number;
};

export function idempotencyIntentFor(
  current: IdempotencyIntentState,
  intent: string,
  createKey: () => string,
): IdempotencyIntentState {
  if (current.intent === intent && current.key !== null) {
    return current;
  }
  return { intent, key: createKey() };
}

export async function idempotencyIntentForBrowser(
  current: IdempotencyIntentState,
  intent: string,
  technicalScope: string,
  createKey: () => string = () => crypto.randomUUID(),
  storage: IdempotencyIntentStorage | null = browserSessionStorage(),
  now: number = Date.now(),
): Promise<IdempotencyIntentState> {
  if (current.intent === intent && current.key !== null) {
    return current;
  }

  // A changed intent during the same mounted interaction is known to be a new command, so rotate
  // immediately. After a reload current.intent is intentionally empty; in that case the browser
  // must first retry the outstanding technical key and let the server prove whether the payload is
  // identical or conflicting.
  if (current.intent !== null && current.intent !== intent) {
    const next = { intent, key: createValidKey(createKey) };
    await persistRetryIntent(technicalScope, next.key, storage, now);
    return next;
  }

  const restoredKey = await readRetryIntent(technicalScope, storage, now);
  if (restoredKey !== null) {
    return { intent, key: restoredKey };
  }

  const next = { intent, key: createValidKey(createKey) };
  await persistRetryIntent(technicalScope, next.key, storage, now);
  return next;
}

export async function clearBrowserIdempotencyIntent(
  technicalScope: string,
  storage: IdempotencyIntentStorage | null = browserSessionStorage(),
): Promise<void> {
  if (storage === null) {
    return;
  }
  const storageKey = await retryIntentStorageKey(technicalScope);
  if (storageKey === null) {
    return;
  }
  try {
    storage.removeItem(storageKey);
  } catch {
    // Storage can be unavailable in locked-down browser contexts. The caller still clears its
    // in-memory state; failure to remove a technical retry key must never fail a confirmed write.
  }
}

async function readRetryIntent(
  technicalScope: string,
  storage: IdempotencyIntentStorage | null,
  now: number,
): Promise<string | null> {
  if (storage === null) {
    return null;
  }
  const storageKey = await retryIntentStorageKey(technicalScope);
  if (storageKey === null) {
    return null;
  }

  let raw: string | null;
  try {
    raw = storage.getItem(storageKey);
  } catch {
    return null;
  }
  if (raw === null) {
    return null;
  }

  let stored: StoredRetryIntent;
  try {
    stored = JSON.parse(raw) as StoredRetryIntent;
  } catch {
    removeStoredRetryIntent(storage, storageKey);
    return null;
  }

  const valid =
    stored.version === browserRetryIntentVersion &&
    typeof stored.key === "string" &&
    idempotencyKeyPattern.test(stored.key) &&
    Number.isSafeInteger(stored.expiresAt) &&
    stored.expiresAt > now &&
    stored.expiresAt <= now + browserRetryIntentTTLMilliseconds;
  if (!valid) {
    removeStoredRetryIntent(storage, storageKey);
    return null;
  }
  return stored.key;
}

async function persistRetryIntent(
  technicalScope: string,
  key: string,
  storage: IdempotencyIntentStorage | null,
  now: number,
): Promise<void> {
  if (storage === null) {
    return;
  }
  const storageKey = await retryIntentStorageKey(technicalScope);
  if (storageKey === null) {
    return;
  }
  const stored: StoredRetryIntent = {
    version: browserRetryIntentVersion,
    key,
    expiresAt: now + browserRetryIntentTTLMilliseconds,
  };
  try {
    storage.setItem(storageKey, JSON.stringify(stored));
  } catch {
    // Persistence is a retry-hardening mechanism, not permission to fail the user's write when
    // sessionStorage is unavailable. The current mounted interaction still keeps the same key.
  }
}

async function retryIntentStorageKey(technicalScope: string): Promise<string | null> {
  const normalizedScope = technicalScope.trim();
  if (normalizedScope === "" || typeof crypto === "undefined" || crypto.subtle === undefined) {
    return null;
  }
  try {
    const encoded = new TextEncoder().encode(`openinvest/idempotency-retry/v1/${normalizedScope}`);
    const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", encoded));
    const hash = Array.from(digest, (byte) => byte.toString(16).padStart(2, "0")).join("");
    return `${browserRetryIntentPrefix}${hash}`;
  } catch {
    return null;
  }
}

function createValidKey(createKey: () => string): string {
  const key = createKey();
  if (!idempotencyKeyPattern.test(key)) {
    throw new Error("generated Idempotency-Key does not satisfy the API contract");
  }
  return key;
}

function removeStoredRetryIntent(storage: IdempotencyIntentStorage, storageKey: string) {
  try {
    storage.removeItem(storageKey);
  } catch {
    // Ignore browser storage failures; malformed technical state is treated as absent.
  }
}

function browserSessionStorage(): IdempotencyIntentStorage | null {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    return window.sessionStorage;
  } catch {
    return null;
  }
}
