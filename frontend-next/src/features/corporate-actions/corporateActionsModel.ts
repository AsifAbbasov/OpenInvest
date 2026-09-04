export const MAX_CORPORATE_ACTION_INSTRUMENTS = 50;

import type { CorporateActionEvent, CorporateActionHeatmapBucket } from "@/common/api/openinvest";

export function parseCorporateActionInstrumentInput(value: string): string[] {
  return [...new Set(
    value
      .split(/[\s,]+/)
      .map((item) => item.trim())
      .filter(Boolean),
  )];
}

export function corporateActionEffectiveDateLabel(event: CorporateActionEvent, effectiveDate: string): string {
  if (event.kind === "DIVIDEND" && event.recordDate === effectiveDate) {
    return "Record date";
  }
  return "Payment date";
}

export function corporateActionHeatmapLevel(totalCount: number, maximumCount: number): 0 | 1 | 2 | 3 | 4 {
  if (totalCount <= 0 || maximumCount <= 0) {
    return 0;
  }
  const ratio = totalCount / maximumCount;
  if (ratio <= 0.25) return 1;
  if (ratio <= 0.5) return 2;
  if (ratio <= 0.75) return 3;
  return 4;
}

export function corporateActionHeatmapMaximum(buckets: CorporateActionHeatmapBucket[]): number {
  return buckets.reduce((maximum, bucket) => Math.max(maximum, bucket.totalCount), 0);
}
