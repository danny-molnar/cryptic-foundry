import type { PuzzleDocument } from '$lib/puzzle/types';

export type AnalysisCandidate = {
	answer: string;
	mechanism: string;
	fodder?: string;
	pattern?: string;
	indicator?: string;
	matches_pattern?: boolean;
};

export type AnalysisResponse = {
	clue: string;
	enumeration: { raw: string; parts: number[]; total: number };
	candidates: AnalysisCandidate[];
};

export async function savePuzzle(document: PuzzleDocument): Promise<PuzzleDocument> {
	const response = await fetch(document.id ? `/v1/puzzles/${document.id}` : '/v1/puzzles', {
		method: document.id ? 'PUT' : 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(document)
	});
	return decode<PuzzleDocument>(response);
}

export async function analyseClue(clue: string, known: string): Promise<AnalysisResponse> {
	const response = await fetch('/v1/tools/analyse', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ clue, known })
	});
	return decode<AnalysisResponse>(response);
}

async function decode<T>(response: Response): Promise<T> {
	const body = await response.json();
	if (!response.ok) throw new Error(body.error ?? `Request failed with HTTP ${response.status}`);
	return body as T;
}
