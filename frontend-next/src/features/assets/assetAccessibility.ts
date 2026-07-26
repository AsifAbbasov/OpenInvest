export type AssetSearchStatus = "idle" | "loading" | "loadingMore" | "ready" | "empty" | "error";
export type AssetDetailStatus = "idle" | "loading" | "available" | "deferred" | "error";
export type AssetDetailFocusState =
  | { status: "idle" }
  | { status: Exclude<AssetDetailStatus, "idle">; ticker: string };

export function assetSearchStatusMessage(status: AssetSearchStatus, count: number) {
  if (status === "loading") {
    return "Loading supported assets.";
  }
  if (status === "loadingMore") {
    return "Loading more supported assets.";
  }
  if (status === "empty") {
    return "No supported assets matched this search.";
  }
  if (count > 0) {
    return `${count} supported asset${count === 1 ? "" : "s"} shown.`;
  }
  return "";
}

export function assetDetailStatusMessage(status: AssetDetailStatus, ticker?: string) {
  if (!ticker) {
    return "";
  }
  if (status === "loading") {
    return `Checking detail availability for ${ticker}.`;
  }
  if (status === "available") {
    return `Detail is available for ${ticker}.`;
  }
  if (status === "deferred") {
    return `Detail is unavailable for ${ticker}.`;
  }
  return "";
}

export function shouldCloseAssetDetailForKey(key: string, detailOpen: boolean) {
  return detailOpen && key === "Escape";
}

export function assetDetailFocusRestoreTarget(originTicker: string | null, visibleTickers: string[]) {
  if (originTicker && visibleTickers.includes(originTicker)) {
    return { kind: "result" as const, ticker: originTicker };
  }
  return { kind: "searchInput" as const };
}

export function shouldFocusAssetDetailRegion(
  previous: AssetDetailFocusState,
  next: AssetDetailFocusState,
) {
  if (next.status === "idle") {
    return false;
  }
  if (previous.status === "idle") {
    return true;
  }
  if (next.status === "loading") {
    return true;
  }
  return previous.ticker !== next.ticker;
}

export type AssetDetailAttempt = {
  generation: number;
  ticker: string;
};

export function nextAssetDetailAttempt(currentGeneration: number, ticker: string): AssetDetailAttempt {
  return {
    generation: currentGeneration + 1,
    ticker,
  };
}

export function cancelAssetDetailGeneration(currentGeneration: number) {
  return currentGeneration + 1;
}

export function shouldAcceptAssetDetailResult(currentGeneration: number, attempt: AssetDetailAttempt) {
  return currentGeneration === attempt.generation;
}
