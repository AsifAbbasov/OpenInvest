export type ImportOperationGuardState = {
	scope: string;
	reviewGeneration: number;
	appendGeneration: number;
};

export type ImportReviewAttempt = {
	scope: string;
	reviewGeneration: number;
};

export type ImportAppendAttempt = {
	scope: string;
	appendGeneration: number;
};

export function synchronizeImportScope(current: ImportOperationGuardState, scope: string): ImportOperationGuardState {
	if (current.scope === scope) {
		return current;
	}
	return {
		scope,
		reviewGeneration: current.reviewGeneration + 1,
		appendGeneration: current.appendGeneration + 1,
	};
}

export function startImportReview(current: ImportOperationGuardState, scope: string): {
	state: ImportOperationGuardState;
	attempt: ImportReviewAttempt;
} {
	const reviewGeneration = current.reviewGeneration + 1;
	return {
		state: { scope, reviewGeneration, appendGeneration: current.appendGeneration + 1 },
		attempt: { scope, reviewGeneration },
	};
}

export function startImportAppend(current: ImportOperationGuardState, scope: string): {
	state: ImportOperationGuardState;
	attempt: ImportAppendAttempt;
} {
	const appendGeneration = current.appendGeneration + 1;
	return {
		state: { ...current, scope, appendGeneration },
		attempt: { scope, appendGeneration },
	};
}

export function shouldCommitImportReview(current: ImportOperationGuardState, attempt: ImportReviewAttempt) {
	return current.scope === attempt.scope && current.reviewGeneration === attempt.reviewGeneration;
}

export function shouldCommitImportAppend(current: ImportOperationGuardState, attempt: ImportAppendAttempt) {
	return current.scope === attempt.scope && current.appendGeneration === attempt.appendGeneration;
}
