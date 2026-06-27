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
