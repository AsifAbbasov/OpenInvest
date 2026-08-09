export type IdempotencyIntentState = {
  intent: string | null;
  key: string | null;
};

export const emptyIdempotencyIntent: IdempotencyIntentState = {
  intent: null,
  key: null,
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
