<script lang="ts">
	import { browser } from '$app/environment';
	import { resolve } from '$app/paths';
	import { analyseClue, getPuzzle, savePuzzle, type AnalysisResponse } from '$lib/api';
	import { createGrid, deriveEntries } from '$lib/puzzle/grid';
	import { parsePuzzleDocument, validatePuzzleDocument } from '$lib/puzzle/document';
	import type { Cell, Direction, PuzzleDocument } from '$lib/puzzle/types';
	import { onMount } from 'svelte';

	const draftKey = 'cryptic-foundry:draft:v1';
	const legacyDraftKey = 'cryptic-workshop:draft:v1';
	type Snapshot = {
		puzzleID?: string;
		title: string;
		author: string;
		size: number;
		grid: Cell[][];
		clueText: Record<string, string>;
	};

	let title = $state('Untitled cryptic');
	let puzzleID = $state<string>();
	let author = $state('Danny Molnar');
	let size = $state(7);
	let grid = $state(createGrid(7, 7));
	let selected = $state({ row: 0, column: 0 });
	let direction = $state<Direction>('across');
	let clueText = $state<Record<string, string>>({});
	let statusMessage = $state('Draft is kept in this browser');
	let symmetry = $state(true);
	let undoStack = $state<string[]>([]);
	let redoStack = $state<string[]>([]);
	let readyToSave = $state(false);
	let fileInput = $state<HTMLInputElement>();
	let saving = $state(false);
	let analysing = $state<Record<string, boolean>>({});
	let analyses = $state<Record<string, AnalysisResponse>>({});
	let entries = $derived(deriveEntries(grid));
	let issues = $derived(validatePuzzleDocument(buildDocument()));

	onMount(() => {
		void initialise();
	});

	async function initialise() {
		const saved = localStorage.getItem(draftKey) ?? localStorage.getItem(legacyDraftKey);
		if (saved) {
			try {
				restore(saved);
				localStorage.setItem(draftKey, saved);
				localStorage.removeItem(legacyDraftKey);
				statusMessage = 'Recovered local draft';
			} catch {
				localStorage.removeItem(draftKey);
			}
		}
		const requestedID = new URLSearchParams(location.search).get('id');
		if (requestedID) {
			try {
				loadDocument(await getPuzzle(requestedID));
				statusMessage = 'Loaded persistent draft';
			} catch (error) {
				statusMessage = error instanceof Error ? error.message : 'Could not load draft';
			}
		}
		readyToSave = true;
	}

	$effect(() => {
		if (browser && readyToSave) localStorage.setItem(draftKey, capture());
	});

	function capture() {
		return JSON.stringify({ puzzleID, title, author, size, grid, clueText } satisfies Snapshot);
	}

	function restore(snapshot: string) {
		const value = JSON.parse(snapshot) as Snapshot;
		puzzleID = value.puzzleID;
		title = value.title;
		author = value.author;
		size = value.size;
		grid = value.grid;
		clueText = value.clueText;
		selected = { row: 0, column: 0 };
	}

	function recordHistory() {
		undoStack = [...undoStack.slice(-49), capture()];
		redoStack = [];
	}

	function undo() {
		const previous = undoStack.at(-1);
		if (!previous) return;
		redoStack = [...redoStack, capture()];
		undoStack = undoStack.slice(0, -1);
		restore(previous);
		statusMessage = 'Undid last edit';
	}

	function redo() {
		const next = redoStack.at(-1);
		if (!next) return;
		undoStack = [...undoStack, capture()];
		redoStack = redoStack.slice(0, -1);
		restore(next);
		statusMessage = 'Redid edit';
	}

	function resize(nextSize: number) {
		recordHistory();
		size = nextSize;
		grid = createGrid(size, size);
		selected = { row: 0, column: 0 };
		clueText = {};
	}

	function toggleBlock(row = selected.row, column = selected.column) {
		recordHistory();
		const cell = grid[row][column];
		cell.block = !cell.block;
		cell.solution = '';
		if (symmetry) {
			const mirror = grid[size - row - 1][size - column - 1];
			mirror.block = cell.block;
			mirror.solution = '';
		}
	}

	function move(deltaRow: number, deltaColumn: number) {
		selected = {
			row: Math.max(0, Math.min(size - 1, selected.row + deltaRow)),
			column: Math.max(0, Math.min(size - 1, selected.column + deltaColumn))
		};
	}

	function advance(backward = false) {
		const amount = backward ? -1 : 1;
		if (direction === 'across') move(0, amount);
		else move(amount, 0);
	}

	function handleKeydown(event: KeyboardEvent) {
		if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'z') {
			if (event.shiftKey) redo();
			else undo();
			event.preventDefault();
			return;
		}
		if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'y') {
			redo();
			event.preventDefault();
			return;
		}
		if ((event.target as HTMLElement).matches('input, textarea')) return;
		const cell = grid[selected.row][selected.column];
		if (/^[a-zA-Z]$/.test(event.key) && !cell.block) {
			recordHistory();
			cell.solution = event.key.toUpperCase();
			advance();
			event.preventDefault();
			return;
		}
		if (event.key === 'Backspace' && !cell.block) {
			if (cell.solution) {
				recordHistory();
				cell.solution = '';
			} else advance(true);
			event.preventDefault();
			return;
		}
		const moves: Record<string, [number, number]> = {
			ArrowUp: [-1, 0],
			ArrowDown: [1, 0],
			ArrowLeft: [0, -1],
			ArrowRight: [0, 1]
		};
		if (moves[event.key]) {
			move(...moves[event.key]);
			event.preventDefault();
		} else if (event.key === ' ') {
			toggleBlock();
			event.preventDefault();
		} else if (event.key === 'Enter') {
			direction = direction === 'across' ? 'down' : 'across';
			event.preventDefault();
		}
	}

	function answerFor(entry: (typeof entries)[number]) {
		return entry.cells.map(({ row, column }) => grid[row][column].solution || '?').join('');
	}

	function numberAt(row: number, column: number) {
		return entries.find((entry) => entry.cells[0]?.row === row && entry.cells[0]?.column === column)
			?.number;
	}

	function buildDocument(): PuzzleDocument {
		return {
			schemaVersion: 1,
			id: puzzleID,
			title,
			author,
			type: 'cryptic',
			status: 'draft',
			grid: {
				rows: size,
				cols: size,
				cells: grid.map((row) =>
					row.map((cell) => ({
						block: cell.block || undefined,
						solution: cell.solution || undefined
					}))
				)
			},
			entries: entries.map((entry) => {
				const answer = answerFor(entry);
				return {
					...entry,
					answer: answer.includes('?') ? undefined : answer,
					enumeration: String(entry.cells.length),
					clue: clueText[entry.id] ?? ''
				};
			})
		};
	}

	function exportPuzzle() {
		const json = JSON.stringify(buildDocument(), null, 2);
		const url = URL.createObjectURL(new Blob([json], { type: 'application/json' }));
		const anchor = document.createElement('a');
		anchor.href = url;
		anchor.download = `${title.toLowerCase().replaceAll(/[^a-z0-9]+/g, '-') || 'puzzle'}.json`;
		anchor.click();
		URL.revokeObjectURL(url);
		statusMessage = `Exported ${entries.length} entries`;
	}

	async function importPuzzle(file: File | undefined) {
		if (!file) return;
		try {
			const document = parsePuzzleDocument(await file.text());
			recordHistory();
			loadDocument(document);
			statusMessage = `Imported ${file.name}`;
		} catch (error) {
			statusMessage = error instanceof Error ? error.message : 'Could not import puzzle';
		} finally {
			if (fileInput) fileInput.value = '';
		}
	}

	function loadDocument(document: PuzzleDocument) {
		puzzleID = document.id;
		title = document.title;
		author = document.author;
		size = document.grid.rows;
		grid = document.grid.cells.map((row, rowIndex) =>
			row.map((cell, columnIndex) => ({
				row: rowIndex,
				column: columnIndex,
				block: Boolean(cell.block),
				solution: cell.solution?.toUpperCase() ?? ''
			}))
		);
		clueText = Object.fromEntries(document.entries.map((entry) => [entry.id, entry.clue ?? '']));
		selected = { row: 0, column: 0 };
	}

	async function saveDraft() {
		saving = true;
		try {
			const saved = await savePuzzle(buildDocument());
			puzzleID = saved.id;
			statusMessage = `Saved draft ${saved.id}`;
		} catch (error) {
			statusMessage = error instanceof Error ? error.message : 'Could not save draft';
		} finally {
			saving = false;
		}
	}

	async function analyseEntry(entry: (typeof entries)[number]) {
		const clue = clueText[entry.id]?.trim();
		if (!clue) {
			statusMessage = 'Write a clue before analysing it';
			return;
		}
		analysing[entry.id] = true;
		try {
			const fullClue = /\([\d, -]+\)\s*$/.test(clue) ? clue : `${clue} (${entry.cells.length})`;
			analyses[entry.id] = await analyseClue(fullClue, answerFor(entry));
			statusMessage = `Analysed ${entry.number} ${entry.direction}`;
		} catch (error) {
			statusMessage = error instanceof Error ? error.message : 'Could not analyse clue';
		} finally {
			analysing[entry.id] = false;
		}
	}
</script>

<svelte:head>
	<title>Cryptic Foundry</title>
	<meta name="description" content="A focused workspace for constructing cryptic crosswords." />
</svelte:head>

<svelte:window onkeydown={handleKeydown} />

<main>
	<header>
		<div>
			<p class="eyebrow">Cryptic Foundry</p>
			<h1>Set the grid. Shape the clue.</h1>
			<a class="library-link" href={resolve('/library')}>View saved drafts →</a>
		</div>
		<div class="document-meta">
			<label>Title <input bind:value={title} onfocus={recordHistory} /></label>
			<label>Setter <input bind:value={author} onfocus={recordHistory} /></label>
		</div>
	</header>

	<section class="toolbar" aria-label="Editor controls">
		<div class="size-control">
			<span>Grid</span>
			{#each [5, 7, 9, 15] as option (option)}
				<button class:active={size === option} onclick={() => resize(option)}
					>{option}×{option}</button
				>
			{/each}
		</div>
		<div class="direction-control">
			<span>Writing</span>
			<button onclick={() => (direction = direction === 'across' ? 'down' : 'across')}>
				{direction === 'across' ? '→ Across' : '↓ Down'}
			</button>
		</div>
		<label class="symmetry-control">
			<input type="checkbox" bind:checked={symmetry} /> 180° symmetry
		</label>
		<div class="history-control">
			<button onclick={undo} disabled={undoStack.length === 0} title="Undo">↶</button>
			<button onclick={redo} disabled={redoStack.length === 0} title="Redo">↷</button>
		</div>
		<input
			class="file-input"
			type="file"
			accept="application/json,.json"
			bind:this={fileInput}
			onchange={(event) => importPuzzle(event.currentTarget.files?.[0])}
		/>
		<button onclick={() => fileInput?.click()}>Import JSON</button>
		<button
			onclick={saveDraft}
			disabled={saving || issues.some((issue) => issue.severity === 'error')}
		>
			{saving ? 'Saving…' : puzzleID ? 'Update draft' : 'Save draft'}
		</button>
		<button class="export" onclick={exportPuzzle}>Export JSON</button>
	</section>

	<section class="workspace">
		<div class="board-panel">
			<div class="board" style={`--size: ${size}`} aria-label={`${size} by ${size} crossword grid`}>
				{#each grid as row, rowIndex (rowIndex)}
					{#each row as cell, columnIndex (columnIndex)}
						<button
							class:block={cell.block}
							class:selected={selected.row === rowIndex && selected.column === columnIndex}
							aria-label={`Row ${rowIndex + 1}, column ${columnIndex + 1}${cell.block ? ', block' : ''}`}
							onclick={() => {
								if (selected.row === rowIndex && selected.column === columnIndex) {
									direction = direction === 'across' ? 'down' : 'across';
								}
								selected = { row: rowIndex, column: columnIndex };
							}}
							oncontextmenu={(event) => {
								event.preventDefault();
								toggleBlock(rowIndex, columnIndex);
							}}
						>
							{#if numberAt(rowIndex, columnIndex)}
								<span class="number">{numberAt(rowIndex, columnIndex)}</span>
							{/if}
							<span class="letter">{cell.solution}</span>
						</button>
					{/each}
				{/each}
			</div>
			<div class="keyboard-help">
				<span><kbd>Letters</kbd> fill</span><span><kbd>Space</kbd> block</span>
				<span><kbd>Enter</kbd> turn</span><span><kbd>Right click</kbd> block</span>
			</div>
			<p class="status">{statusMessage}</p>
			<div class="validation" class:valid={issues.length === 0}>
				<strong
					>{issues.length === 0 ? 'Draft checks passed' : `${issues.length} draft checks`}</strong
				>
				{#each issues as issue, index (`${issue.severity}-${index}`)}
					<p class:error={issue.severity === 'error'}>{issue.message}</p>
				{/each}
			</div>
		</div>

		<aside>
			{#each ['across', 'down'] as group (group)}
				<section class="clue-group">
					<h2>{group}</h2>
					{#each entries.filter((entry) => entry.direction === group) as entry (entry.id)}
						{@const analysis = analyses[entry.id]}
						<label class="clue-row">
							<span class="clue-number">{entry.number}</span>
							<span class="answer">{answerFor(entry)} <small>({entry.cells.length})</small></span>
							<input
								value={clueText[entry.id] ?? ''}
								onfocus={recordHistory}
								oninput={(event) => (clueText[entry.id] = event.currentTarget.value)}
								placeholder="Write the clue…"
							/>
							<button
								class="analyse"
								type="button"
								disabled={analysing[entry.id]}
								onclick={() => analyseEntry(entry)}
							>
								{analysing[entry.id] ? 'Thinking…' : 'Analyse'}
							</button>
						</label>
						{#if analysis}
							<div class="analysis-result">
								<strong>Suggested parses</strong>
								{#each analysis.candidates.slice(0, 4) as candidate (candidate.answer)}
									<span class:match={candidate.matches_pattern}>{candidate.answer}</span>
								{:else}
									<em>No candidates in the current word list.</em>
								{/each}
							</div>
						{/if}
					{:else}
						<p class="empty">Add open runs of two or more cells.</p>
					{/each}
				</section>
			{/each}
		</aside>
	</section>
</main>

<style>
	:global(*) {
		box-sizing: border-box;
	}
	:global(body) {
		margin: 0;
		background: #f3efe5;
		color: #1f2925;
		font-family: Georgia, 'Times New Roman', serif;
	}
	:global(button),
	:global(input) {
		font: inherit;
	}
	main {
		min-height: 100vh;
		padding: clamp(1.25rem, 4vw, 4rem);
		background:
			radial-gradient(circle at 12% 8%, #fffdf6 0, transparent 32%),
			linear-gradient(135deg, transparent 0 74%, rgba(176, 65, 49, 0.07) 74%);
	}
	header {
		max-width: 1280px;
		margin: 0 auto 2rem;
		display: flex;
		justify-content: space-between;
		gap: 2rem;
		align-items: end;
		border-bottom: 1px solid #b9b2a2;
		padding-bottom: 1.4rem;
	}
	.eyebrow {
		margin: 0 0 0.4rem;
		color: #a23d2e;
		font:
			700 0.74rem/1.2 system-ui,
			sans-serif;
		letter-spacing: 0.18em;
		text-transform: uppercase;
	}
	.library-link {
		display: inline-block;
		margin-top: 0.8rem;
		color: #a23d2e;
		font:
			700 0.74rem system-ui,
			sans-serif;
		text-decoration: none;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}
	h1 {
		margin: 0;
		font-size: clamp(2rem, 4vw, 4.4rem);
		font-weight: 500;
		line-height: 0.98;
		letter-spacing: -0.035em;
	}
	.document-meta {
		display: grid;
		gap: 0.7rem;
		min-width: min(100%, 300px);
	}
	label {
		font:
			650 0.72rem/1.2 system-ui,
			sans-serif;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}
	input {
		display: block;
		width: 100%;
		margin-top: 0.3rem;
		padding: 0.62rem 0.7rem;
		border: 1px solid #aaa292;
		background: rgba(255, 255, 255, 0.52);
		color: inherit;
		border-radius: 2px;
		text-transform: none;
		letter-spacing: 0;
	}
	.toolbar {
		max-width: 1280px;
		margin: 0 auto 1rem;
		display: flex;
		gap: 1rem;
		align-items: center;
		font:
			650 0.75rem system-ui,
			sans-serif;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}
	.toolbar div {
		display: flex;
		gap: 0.35rem;
		align-items: center;
	}
	.toolbar span {
		margin-right: 0.35rem;
		color: #69675f;
	}
	.symmetry-control {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		white-space: nowrap;
	}
	.symmetry-control input {
		width: auto;
		margin: 0;
		accent-color: #a23d2e;
	}
	.file-input {
		display: none;
	}
	.toolbar button:disabled {
		cursor: not-allowed;
		opacity: 0.35;
	}
	button {
		border: 1px solid #938c7e;
		background: #fcfaf4;
		color: inherit;
		cursor: pointer;
	}
	.toolbar button {
		padding: 0.5rem 0.7rem;
	}
	.toolbar button:hover,
	.toolbar button.active {
		color: #fff;
		background: #263c34;
	}
	.toolbar .export {
		margin-left: auto;
		background: #a23d2e;
		color: #fff;
		border-color: #a23d2e;
		padding-inline: 1rem;
	}
	.workspace {
		max-width: 1280px;
		margin: 0 auto;
		display: grid;
		grid-template-columns: minmax(340px, 0.95fr) minmax(340px, 1.05fr);
		gap: clamp(1.5rem, 4vw, 4rem);
		align-items: start;
	}
	.board-panel {
		padding: clamp(1rem, 3vw, 2.2rem);
		background: #263c34;
		box-shadow: 0 20px 50px rgba(33, 38, 33, 0.16);
	}
	.board {
		display: grid;
		grid-template-columns: repeat(var(--size), 1fr);
		aspect-ratio: 1;
		background: #1f2925;
		border: 3px solid #1f2925;
		gap: 1px;
	}
	.board button {
		position: relative;
		min-width: 0;
		padding: 0;
		border: 0;
		background: #fffdf7;
		aspect-ratio: 1;
	}
	.board button.block {
		background: #1f2925;
	}
	.board button.selected:not(.block) {
		background: #f0ca83;
		outline: 3px solid #b33f2d;
		outline-offset: -3px;
		z-index: 1;
	}
	.number {
		position: absolute;
		top: 4%;
		left: 6%;
		font:
			700 clamp(0.43rem, 1vw, 0.72rem) system-ui,
			sans-serif;
	}
	.letter {
		font:
			600 clamp(0.9rem, 3.2vw, 2.25rem)/1 system-ui,
			sans-serif;
	}
	.keyboard-help {
		display: flex;
		flex-wrap: wrap;
		gap: 0.8rem 1.2rem;
		margin-top: 1.1rem;
		color: #d8dfd7;
		font:
			0.7rem system-ui,
			sans-serif;
	}
	kbd {
		color: #f0ca83;
		font:
			700 0.67rem system-ui,
			sans-serif;
	}
	.status {
		color: #f0ca83;
		margin: 0.8rem 0 0;
		font:
			0.75rem system-ui,
			sans-serif;
	}
	.validation {
		margin-top: 1rem;
		padding: 0.9rem;
		border-left: 3px solid #f0ca83;
		background: rgba(255, 255, 255, 0.08);
		color: #f3efe5;
		font:
			0.75rem/1.4 system-ui,
			sans-serif;
	}
	.validation.valid {
		border-color: #91b89b;
	}
	.validation p {
		margin: 0.35rem 0 0;
		color: #f0ca83;
	}
	.validation p.error {
		color: #ff9d8d;
	}
	aside {
		display: grid;
		gap: 2rem;
	}
	.clue-group h2 {
		margin: 0 0 0.8rem;
		padding-bottom: 0.5rem;
		border-bottom: 3px solid #263c34;
		font:
			750 0.78rem system-ui,
			sans-serif;
		letter-spacing: 0.16em;
		text-transform: uppercase;
	}
	.clue-row {
		display: grid;
		grid-template-columns: 2rem minmax(6rem, 0.35fr) 1fr auto;
		gap: 0.7rem;
		align-items: center;
		margin-bottom: 0.55rem;
	}
	.clue-row input {
		margin: 0;
	}
	.clue-number {
		font:
			750 1rem system-ui,
			sans-serif;
		color: #a23d2e;
	}
	.analyse {
		align-self: stretch;
		padding: 0 0.7rem;
		border-color: #a23d2e;
		color: #a23d2e;
		background: transparent;
		font:
			700 0.67rem system-ui,
			sans-serif;
		text-transform: uppercase;
	}
	.analysis-result {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem;
		align-items: center;
		margin: -0.1rem 0 0.8rem 2.7rem;
		font:
			0.72rem system-ui,
			sans-serif;
	}
	.analysis-result span {
		padding: 0.2rem 0.4rem;
		background: #ded8c9;
		text-transform: uppercase;
	}
	.analysis-result span.match {
		background: #263c34;
		color: #fff;
	}
	.answer {
		overflow: hidden;
		font:
			700 0.8rem system-ui,
			sans-serif;
		letter-spacing: 0.08em;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.answer small {
		color: #777369;
		font-weight: 500;
	}
	.empty {
		color: #777369;
		font-style: italic;
	}
	@media (max-width: 820px) {
		header {
			align-items: stretch;
			flex-direction: column;
		}
		.workspace {
			grid-template-columns: 1fr;
		}
		.toolbar {
			align-items: stretch;
			flex-wrap: wrap;
		}
		.toolbar .export {
			margin-left: 0;
		}
	}
</style>
