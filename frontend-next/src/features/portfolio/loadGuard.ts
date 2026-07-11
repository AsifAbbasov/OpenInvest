export type PortfolioLoadGuardState = {
  generation: number;
  accessToken: string;
};

export type PortfolioLoadAttempt = {
  generation: number;
  accessToken: string;
};

export function startPortfolioLoad(
  current: PortfolioLoadGuardState,
  accessToken: string,
): { state: PortfolioLoadGuardState; attempt: PortfolioLoadAttempt } {
  const generation = current.generation + 1;
  return {
    state: { generation, accessToken },
    attempt: { generation, accessToken },
  };
}

export function shouldCommitPortfolioLoad(current: PortfolioLoadGuardState, attempt: PortfolioLoadAttempt) {
  return current.generation === attempt.generation && current.accessToken === attempt.accessToken;
}
