import { describe, expect, it } from 'vitest';
import { parsePuzzleDocument, validatePuzzleDocument } from './document';
import type { PuzzleDocument } from './types';

const document: PuzzleDocument = {
	schemaVersion: 1,
	title: 'Tiny cryptic',
	author: 'Setter',
	type: 'cryptic',
	status: 'draft',
	grid: {
		rows: 2,
		cols: 2,
		cells: [
			[{ solution: 'A' }, { solution: 'B' }],
			[{ solution: 'C' }, { solution: 'D' }]
		]
	},
	entries: [
		{
			id: '0-0-across',
			number: 1,
			direction: 'across',
			cells: [
				{ row: 0, column: 0 },
				{ row: 0, column: 1 }
			],
			answer: 'AB',
			enumeration: '2',
			clue: 'A tiny test (2)'
		}
	]
};

describe('puzzle documents', () => {
	it('parses a valid versioned document', () => {
		expect(parsePuzzleDocument(JSON.stringify(document))).toEqual(document);
	});

	it('reports unfinished draft content as warnings', () => {
		const draft = structuredClone(document);
		draft.entries[0].answer = undefined;
		draft.entries[0].clue = '';
		expect(validatePuzzleDocument(draft)).toEqual([
			{ severity: 'warning', message: '1 entries still need complete answers.' },
			{ severity: 'warning', message: '1 entries still need clues.' }
		]);
	});

	it('rejects unknown schema versions', () => {
		expect(() => parsePuzzleDocument('{"schemaVersion":2}')).toThrow('Unsupported schema');
	});

	it('rejects rectangular grids the current editor cannot render', () => {
		const rectangular = structuredClone(document);
		rectangular.grid.cols = 3;
		rectangular.grid.cells = [
			[{ solution: 'A' }, { solution: 'B' }, { solution: 'C' }],
			[{ solution: 'D' }, { solution: 'E' }, { solution: 'F' }]
		];
		expect(() => parsePuzzleDocument(JSON.stringify(rectangular))).toThrow('square grid');
	});
});
