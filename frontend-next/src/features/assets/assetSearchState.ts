import type { AssetSummary, AssetType, Pagination } from "@/common/api/openinvest";

export type AssetSearchKey = {
  query: string;
  assetType: AssetType | "";
};

export type AssetSearchAttempt = AssetSearchKey & {
  generation: number;
  cursor: string | null;
};

export type AssetSearchState = AssetSearchKey & {
  generation: number;
  items: AssetSummary[];
  pagination: Pagination;
  selectedTicker: string | null;
};

export const emptyAssetPagination: Pagination = {
  nextCursor: null,
  hasMore: false,
  limit: 20,
};

export function initialAssetSearchState(): AssetSearchState {
  return {
    query: "",
    assetType: "",
    generation: 0,
    items: [],
    pagination: emptyAssetPagination,
    selectedTicker: null,
  };
}

export function beginAssetSearch(
  state: AssetSearchState,
  key: AssetSearchKey,
  cursor: string | null,
): { state: AssetSearchState; attempt: AssetSearchAttempt } {
  const sameKey = state.query === key.query && state.assetType === key.assetType;
  const generation = state.generation + 1;
  const nextState: AssetSearchState = {
    query: key.query,
    assetType: key.assetType,
    generation,
    items: sameKey && cursor ? state.items : [],
    pagination: sameKey && cursor ? state.pagination : emptyAssetPagination,
    selectedTicker: sameKey && cursor ? state.selectedTicker : null,
  };
  return {
    state: nextState,
    attempt: { ...key, generation, cursor },
  };
}

export function resetAssetSearchForKeyChange(state: AssetSearchState, key: AssetSearchKey): AssetSearchState {
  if (state.query === key.query && state.assetType === key.assetType) {
    return state;
  }
  return {
    query: key.query,
    assetType: key.assetType,
    generation: state.generation + 1,
    items: [],
    pagination: emptyAssetPagination,
    selectedTicker: null,
  };
}

export function acceptAssetSearchResult(
  state: AssetSearchState,
  attempt: AssetSearchAttempt,
  items: AssetSummary[],
  pagination: Pagination,
): AssetSearchState {
  if (
    state.generation !== attempt.generation ||
    state.query !== attempt.query ||
    state.assetType !== attempt.assetType
  ) {
    return state;
  }
  if (attempt.cursor) {
    if (state.pagination.nextCursor !== attempt.cursor) {
      return state;
    }
    return {
      ...state,
      items: [...state.items, ...items],
      pagination,
    };
  }
  return {
    ...state,
    items,
    pagination,
    selectedTicker: null,
  };
}

export function selectAsset(state: AssetSearchState, ticker: string): AssetSearchState {
  return {
    ...state,
    selectedTicker: ticker,
  };
}
