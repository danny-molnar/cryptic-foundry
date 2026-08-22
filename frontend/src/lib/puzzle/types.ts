export type Direction = 'across' | 'down';

export type Cell = {
	row: number;
	column: number;
	block: boolean;
	solution: string;
	number?: number;
};

export type CellReference = Pick<Cell, 'row' | 'column'>;

export type Entry = {
	id: string;
	number: number;
	direction: Direction;
	cells: CellReference[];
};

export type PuzzleEntry = Entry & {
	answer?: string;
	enumeration?: string;
	clue?: string;
	explanation?: string;
	tags?: string[];
};

export type PuzzleDocument = {
	schemaVersion: 1;
	id?: string;
	title: string;
	author: string;
	type: 'quick' | 'cryptic' | 'mixed';
	status: 'draft' | 'published';
	grid: {
		rows: number;
		cols: number;
		cells: Array<Array<{ block?: boolean; solution?: string; given?: boolean }>>;
	};
	entries: PuzzleEntry[];
};
