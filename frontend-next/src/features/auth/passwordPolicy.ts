export type PasswordMode = "login" | "register";

export const PASSWORD_MAX_CODE_POINTS = 256;
export const REGISTRATION_PASSWORD_MIN_CODE_POINTS = 12;

export function isWellFormedPasswordUnicode(password: string) {
  for (let index = 0; index < password.length; index += 1) {
    const codeUnit = password.charCodeAt(index);
    if (codeUnit >= 0xd800 && codeUnit <= 0xdbff) {
      if (index + 1 >= password.length) {
        return false;
      }
      const next = password.charCodeAt(index + 1);
      if (next < 0xdc00 || next > 0xdfff) {
        return false;
      }
      index += 1;
      continue;
    }
    if (codeUnit >= 0xdc00 && codeUnit <= 0xdfff) {
      return false;
    }
  }
  return true;
}

export function passwordCodePointCount(password: string) {
  return Array.from(password).length;
}

export function passwordValidationMessage(password: string, mode: PasswordMode) {
  if (!isWellFormedPasswordUnicode(password)) {
    return "Password contains invalid Unicode.";
  }
  const length = passwordCodePointCount(password);
  const minimum = mode === "register" ? REGISTRATION_PASSWORD_MIN_CODE_POINTS : 1;
  if (length < minimum || length > PASSWORD_MAX_CODE_POINTS) {
    return mode === "register"
      ? "Password must contain 12 to 256 Unicode code points."
      : "Password must contain 1 to 256 Unicode code points.";
  }
  return null;
}
