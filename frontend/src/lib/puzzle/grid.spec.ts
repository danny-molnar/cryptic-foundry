import { describe, expect, it } from 'vitest';
import { createGrid, deriveEntries } from './grid';

describe('deriveEntries', () => {
	it('numbers across and down entries from the same starting cell together', () => {
		const grid = createGrid(3, 3);
		grid[1][0].block = true;
		grid[1][2].block = true;

		const entries = deriveEntries(grid);

		expect(entries.find((entry) => entry.id === '0-0-across')?.number).toBe(1);
		expect(entries.find((entry) => entry.id === '0-1-down')?.number).toBe(2);
		expect(entries.find((entry) => entry.id === '2-0-across')?.number).toBe(3);
	});

	it('does not create one-letter entries', () => {
		const grid = createGrid(1, 3);
		grid[0][1].block = true;
		expect(deriveEntries(grid)).toEqual([]);
	});
});
