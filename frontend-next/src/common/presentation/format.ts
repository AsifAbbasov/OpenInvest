import type { Money } from "@/common/api/openinvest";

export function formatMoney(value: Money) {
  return `${formatDecimalForDisplay(value.amount)} ₽`;
}

export function formatNullableDecimal(value: string | null) {
  return value ?? "Not available yet";
}

export function formatDecimalForDisplay(value: string) {
  const [integerPart, fractionalPart = ""] = value.split(".");
  const groupedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  const visibleFraction = fractionalPart.slice(0, 2).padEnd(2, "0");
  return `${groupedInteger}.${visibleFraction}`;
}
