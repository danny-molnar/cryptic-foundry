import type { PuzzleDocument } from './types';

export type ValidationIssue = {
	severity: 'error' | 'warning';
	message: string;
};

export function parsePuzzleDocument(input: string): PuzzleDocument {
	let value: unknown;
	try {
		value = JSON.parse(input);
	} catch {
		throw new Error('The selected file is not valid JSON.');
	}
	if (!value || typeof value !== 'object') throw new Error('Puzzle JSON must contain an object.');

	const document = value as Partial<PuzzleDocument>;
	const errors = validatePuzzleDocument(document as PuzzleDocument).filter(
		(issue) => issue.severity === 'error'
	);
	if (errors.length > 0) throw new Error(errors.map((issue) => issue.message).join(' '));
	return document as PuzzleDocument;
}

export function validatePuzzleDocument(document: PuzzleDocument): ValidationIssue[] {
	const issues: ValidationIssue[] = [];
	if (document.schemaVersion !== 1)
		addError(`Unsupported schema version ${document.schemaVersion}.`);
	if (!document.title?.trim()) addError('Give the puzzle a title.');
	if (!document.author?.trim()) addWarning('Add a setter name before publishing.');
	if (document.type !== 'cryptic' && document.type !== 'quick' && document.type !== 'mixed') {
		addError('Puzzle type is invalid.');
	}

	const rows = document.grid?.rows;
	const columns = document.grid?.cols;
	if (!Number.isInteger(rows) || !Number.isInteger(columns) || rows < 1 || columns < 1) {
		addError('Grid dimensions must be positive whole numbers.');
		return issues;
	}
	if (rows !== columns) addError('The editor currently requires a square grid.');
	if (!Array.isArray(document.grid.cells) || document.grid.cells.length !== rows) {
		addError(`Grid must contain ${rows} rows.`);
		return issues;
	}
	for (const [rowIndex, row] of document.grid.cells.entries()) {
		if (!Array.isArray(row) || row.length !== columns) {
			addError(`Grid row ${rowIndex + 1} must contain ${columns} cells.`);
			continue;
		}
		for (const [columnIndex, cell] of row.entries()) {
			if (!cell || typeof cell !== 'object') {
				addError(`Grid cell ${rowIndex + 1},${columnIndex + 1} is invalid.`);
			} else if (cell.solution && [...cell.solution].length !== 1) {
				addError(`Grid cell ${rowIndex + 1},${columnIndex + 1} must contain one character.`);
			}
		}
	}

	if (!Array.isArray(document.entries)) {
		addError('Entries must be an array.');
		return issues;
	}
	if (document.entries.length === 0) addWarning('The grid does not contain any entries.');
	const missingAnswers = document.entries.filter((entry) => !entry.answer).length;
	const missingClues = document.entries.filter((entry) => !entry.clue?.trim()).length;
	if (missingAnswers > 0) addWarning(`${missingAnswers} entries still need complete answers.`);
	if (missingClues > 0) addWarning(`${missingClues} entries still need clues.`);

	return issues;

	function addError(message: string) {
		issues.push({ severity: 'error', message });
	}
	function addWarning(message: string) {
		issues.push({ severity: 'warning', message });
	}
}
