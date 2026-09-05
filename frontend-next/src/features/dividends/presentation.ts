export function formatDividendMoney(amount: string) {
  return `${roundMoneyForDisplay(amount)} ₽`;
}

export function formatGrossYield(ratio: string | null) {
  if (ratio === null) {
    return "—";
  }
  return `${ratioToPercentText(ratio)}%`;
}

function roundMoneyForDisplay(value: string) {
  const negative = value.startsWith("-");
  const unsigned = negative ? value.slice(1) : value;
  const [rawWhole = "0", fraction = ""] = unsigned.split(".");
  const whole = rawWhole.replace(/^0+(?=\d)/, "") || "0";
  const keptFraction = fraction.slice(0, 2).padEnd(2, "0");
  const firstDiscarded = fraction[2] ?? "0";
  const remainingDiscarded = fraction.slice(3);
  const lastKeptDigit = keptFraction[1] ?? "0";
  const roundUp =
    firstDiscarded > "5" ||
    (firstDiscarded === "5" &&
      (/[1-9]/.test(remainingDiscarded) || "13579".includes(lastKeptDigit)));

  let unscaled = `${whole}${keptFraction}`;
  if (roundUp) {
    unscaled = incrementDecimalDigits(unscaled);
  }

  const splitAt = unscaled.length - 2;
  const roundedWhole = splitAt > 0 ? unscaled.slice(0, splitAt) : "0";
  const roundedFraction = splitAt > 0 ? unscaled.slice(splitAt) : unscaled.padStart(2, "0");
  const rounded = trimFixedDecimal(`${roundedWhole}.${roundedFraction}`);
  return negative && rounded !== "0" ? `-${rounded}` : rounded;
}

function incrementDecimalDigits(value: string) {
  const digits = value.split("");
  for (let index = digits.length - 1; index >= 0; index -= 1) {
    if (digits[index] !== "9") {
      digits[index] = String.fromCharCode(digits[index].charCodeAt(0) + 1);
      return digits.join("");
    }
    digits[index] = "0";
  }
  return `1${digits.join("")}`;
}

function trimFixedDecimal(value: string) {
  if (!value.includes(".")) {
    return value;
  }
  const trimmed = value.replace(/0+$/, "").replace(/\.$/, "");
  return trimmed === "-0" ? "0" : trimmed;
}

function ratioToPercentText(value: string) {
  const negative = value.startsWith("-");
  const unsigned = negative ? value.slice(1) : value;
  const [whole = "0", fraction = ""] = unsigned.split(".");
  const digits = `${whole}${fraction.padEnd(2, "0")}`.replace(/^0+(?=\d)/, "");
  const decimalPlaces = fraction.length > 2 ? fraction.length - 2 : 0;
  const padded = digits.padStart(decimalPlaces + 1, "0");
  const integerPart = decimalPlaces === 0 ? padded : padded.slice(0, -decimalPlaces);
  const fractionalPart = decimalPlaces === 0 ? "" : padded.slice(-decimalPlaces).replace(/0+$/, "");
  const text = fractionalPart === "" ? integerPart : `${integerPart}.${fractionalPart}`;
  return `${negative ? "-" : ""}${text}`;
}
