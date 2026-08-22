import type { Cell, Direction, Entry } from './types';

export function createGrid(rows: number, columns: number): Cell[][] {
	return Array.from({ length: rows }, (_, row) =>
		Array.from({ length: columns }, (_, column) => ({
			row,
			column,
			block: false,
			solution: ''
		}))
	);
}

export function deriveEntries(grid: Cell[][]): Entry[] {
	if (grid.length === 0 || grid[0]?.length === 0) return [];

	const rows = grid.length;
	const columns = grid[0].length;
	const entries: Entry[] = [];
	let nextNumber = 1;

	for (let row = 0; row < rows; row += 1) {
		for (let column = 0; column < columns; column += 1) {
			if (grid[row][column].block) continue;

			const across = column === 0 || grid[row][column - 1].block;
			const down = row === 0 || grid[row - 1][column].block;
			const acrossCells = across ? collect(grid, row, column, 'across') : [];
			const downCells = down ? collect(grid, row, column, 'down') : [];
			const startsAcross = acrossCells.length >= 2;
			const startsDown = downCells.length >= 2;
			if (!startsAcross && !startsDown) continue;

			const number = nextNumber;
			nextNumber += 1;
			if (startsAcross) entries.push(entry(number, 'across', row, column, acrossCells));
			if (startsDown) entries.push(entry(number, 'down', row, column, downCells));
		}
	}

	return entries.sort((left, right) =>
		left.direction === right.direction
			? left.number - right.number
			: left.direction === 'across'
				? -1
				: 1
	);
}

function collect(
	grid: Cell[][],
	row: number,
	column: number,
	direction: Direction
): Array<{ row: number; column: number }> {
	const cells = [];
	while (grid[row]?.[column] && !grid[row][column].block) {
		cells.push({ row, column });
		if (direction === 'across') column += 1;
		else row += 1;
	}
	return cells;
}

function entry(
	number: number,
	direction: Direction,
	row: number,
	column: number,
	cells: Array<{ row: number; column: number }>
): Entry {
	return { id: `${row}-${column}-${direction}`, number, direction, cells };
}
