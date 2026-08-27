export type UnicodeTextValidationError = "ILL_FORMED" | "TOO_LONG";

export function isWellFormedUnicode(value: string): boolean {
  for (let index = 0; index < value.length; index += 1) {
    const unit = value.charCodeAt(index);
    if (unit >= 0xd800 && unit <= 0xdbff) {
      if (index + 1 >= value.length) {
        return false;
      }
      const next = value.charCodeAt(index + 1);
      if (next < 0xdc00 || next > 0xdfff) {
        return false;
      }
      index += 1;
      continue;
    }
    if (unit >= 0xdc00 && unit <= 0xdfff) {
      return false;
    }
  }
  return true;
}

export function unicodeCodePointCount(value: string): number {
  let count = 0;
  for (const _codePoint of value) {
    count += 1;
  }
  return count;
}

export function unicodeTextValidationError(
  value: string,
  maxCodePoints: number,
): UnicodeTextValidationError | null {
  if (!isWellFormedUnicode(value)) {
    return "ILL_FORMED";
  }
  return unicodeCodePointCount(value) > maxCodePoints ? "TOO_LONG" : null;
}
