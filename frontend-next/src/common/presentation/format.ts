import type { Money } from "@/common/api/openinvest";

export function formatMoney(value: Money) {
  return `${formatDecimalForDisplay(value.amount)} ₽`;
}

export function formatNullableDecimal(value: string | null) {
  return value ?? "Not available yet";
}

export function formatDecimalForDisplay(value: string) {
  const rounded = roundHalfEven(value, 2);
  const [integerPart, visibleFraction = ""] = rounded.split(".");
  const groupedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  return `${groupedInteger}.${visibleFraction}`;
}

export function formatQuantityForDisplay(value: string) {
  const sign = value.startsWith("-") ? "-" : "";
  const unsigned = sign ? value.slice(1) : value;
  const [integerPart = "0", fractionalPart = ""] = unsigned.split(".");
  const groupedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  const significantFraction = fractionalPart.replace(/0+$/, "");
  return significantFraction === "" ? `${sign}${groupedInteger}` : `${sign}${groupedInteger}.${significantFraction}`;
}

// acquisitionBasisWeight is already calculated by the Go backend. This only changes
// the representation from a canonical ratio string (0.34210000) to a UI percent (34.2%).
export function formatRatioAsPercent(value: string | null) {
  if (value === null) {
    return "Undefined";
  }
  return `${roundHalfEven(shiftDecimalRight(value, 2), 1)}%`;
}

function shiftDecimalRight(value: string, places: number) {
  const sign = value.startsWith("-") ? "-" : "";
  const unsigned = sign ? value.slice(1) : value;
  const [integerPart = "0", fractionalPart = ""] = unsigned.split(".");
  const digits = `${integerPart}${fractionalPart}`;
  const pointIndex = integerPart.length + places;
  const padded = pointIndex >= digits.length ? digits.padEnd(pointIndex + 1, "0") : digits;
  const shiftedIntegerRaw = padded.slice(0, pointIndex) || "0";
  const shiftedInteger = shiftedIntegerRaw.replace(/^0+(?=\d)/, "") || "0";
  const shiftedFraction = padded.slice(pointIndex) || "0";
  return `${sign}${shiftedInteger}.${shiftedFraction}`;
}

function roundHalfEven(value: string, scale: number) {
  const sign = value.startsWith("-") ? "-" : "";
  const unsigned = sign ? value.slice(1) : value;
  const [integerPart = "0", fractionalPart = ""] = unsigned.split(".");
  const paddedFraction = fractionalPart.padEnd(scale + 1, "0");
  const keptDigits = paddedFraction.slice(0, scale).split("");
  const nextDigit = Number(paddedFraction[scale] ?? "0");
  const remainingDigits = paddedFraction.slice(scale + 1);
  const lastKeptDigit = Number(keptDigits[keptDigits.length - 1] ?? "0");
  const hasRemainingNonZero = /[1-9]/.test(remainingDigits);
  const shouldRoundUp = nextDigit > 5 || (nextDigit === 5 && (hasRemainingNonZero || lastKeptDigit % 2 === 1));

  let roundedInteger = integerPart === "" ? "0" : integerPart;
  let roundedFraction = keptDigits;

  if (shouldRoundUp) {
    const incremented = incrementDecimalDigits(roundedInteger, roundedFraction);
    roundedInteger = incremented.integerPart;
    roundedFraction = incremented.fractionalDigits;
  }

  return `${sign}${roundedInteger}.${roundedFraction.join("").padEnd(scale, "0")}`;
}

function incrementDecimalDigits(integerPart: string, fractionalDigits: string[]) {
  const digits = [...integerPart, ...fractionalDigits];
  for (let index = digits.length - 1; index >= 0; index -= 1) {
    if (digits[index] !== "9") {
      digits[index] = String(Number(digits[index]) + 1);
      return {
        integerPart: digits.slice(0, integerPart.length).join(""),
        fractionalDigits: digits.slice(integerPart.length),
      };
    }
    digits[index] = "0";
  }
  digits.unshift("1");
  return {
    integerPart: digits.slice(0, integerPart.length + 1).join(""),
    fractionalDigits: digits.slice(integerPart.length + 1),
  };
}
