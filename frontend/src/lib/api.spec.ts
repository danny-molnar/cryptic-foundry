import { afterEach, describe, expect, it, vi } from 'vitest';
import { analyseClue, savePuzzle } from './api';
import type { PuzzleDocument } from './puzzle/types';

const document: PuzzleDocument = {
	schemaVersion: 1,
	title: 'Draft',
	author: 'Setter',
	type: 'cryptic',
	status: 'draft',
	grid: { rows: 1, cols: 1, cells: [[{}]] },
	entries: []
};

afterEach(() => vi.unstubAllGlobals());

describe('editor API', () => {
	it('creates a document without an ID', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ ...document, id: 'puzzle-1' }), {
				status: 201,
				headers: { 'Content-Type': 'application/json' }
			})
		);
		vi.stubGlobal('fetch', fetch);

		await expect(savePuzzle(document)).resolves.toMatchObject({ id: 'puzzle-1' });
		expect(fetch).toHaveBeenCalledWith('/v1/puzzles', expect.objectContaining({ method: 'POST' }));
	});

	it('sends clue analysis to the Rust-backed endpoint', async () => {
		const result = {
			clue: 'Mixed caret',
			enumeration: { raw: '5', parts: [5], total: 5 },
			candidates: []
		};
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify(result), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			})
		);
		vi.stubGlobal('fetch', fetch);

		await expect(analyseClue('Mixed caret (5)', 'R??C?')).resolves.toEqual(result);
		expect(fetch).toHaveBeenCalledWith(
			'/v1/tools/analyse',
			expect.objectContaining({ method: 'POST' })
		);
	});
});
