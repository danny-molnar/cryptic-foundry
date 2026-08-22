<script lang="ts">
	import { createGrid, deriveEntries } from '$lib/puzzle/grid';
	import type { Direction, PuzzleDocument } from '$lib/puzzle/types';

	let title = $state('Untitled cryptic');
	let author = $state('Danny Molnar');
	let size = $state(7);
	let grid = $state(createGrid(7, 7));
	let selected = $state({ row: 0, column: 0 });
	let direction = $state<Direction>('across');
	let clueText = $state<Record<string, string>>({});
	let exportMessage = $state('');
	let entries = $derived(deriveEntries(grid));

	function resize(nextSize: number) {
		size = nextSize;
		grid = createGrid(size, size);
		selected = { row: 0, column: 0 };
		clueText = {};
	}

	function toggleBlock(row = selected.row, column = selected.column) {
		const cell = grid[row][column];
		cell.block = !cell.block;
		cell.solution = '';
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
		if ((event.target as HTMLElement).matches('input, textarea')) return;
		const cell = grid[selected.row][selected.column];
		if (/^[a-zA-Z]$/.test(event.key) && !cell.block) {
			cell.solution = event.key.toUpperCase();
			advance();
			event.preventDefault();
			return;
		}
		if (event.key === 'Backspace' && !cell.block) {
			if (cell.solution) cell.solution = '';
			else advance(true);
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
		exportMessage = `Exported ${entries.length} entries`;
	}
</script>

<svelte:head>
	<title>Cryptic Workshop</title>
	<meta name="description" content="A focused workspace for constructing cryptic crosswords." />
</svelte:head>

<svelte:window onkeydown={handleKeydown} />

<main>
	<header>
		<div>
			<p class="eyebrow">Cryptic Workshop</p>
			<h1>Set the grid. Shape the clue.</h1>
		</div>
		<div class="document-meta">
			<label>Title <input bind:value={title} /></label>
			<label>Setter <input bind:value={author} /></label>
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
			{#if exportMessage}<p class="status">{exportMessage}</p>{/if}
		</div>

		<aside>
			{#each ['across', 'down'] as group (group)}
				<section class="clue-group">
					<h2>{group}</h2>
					{#each entries.filter((entry) => entry.direction === group) as entry (entry.id)}
						<label class="clue-row">
							<span class="clue-number">{entry.number}</span>
							<span class="answer">{answerFor(entry)} <small>({entry.cells.length})</small></span>
							<input
								value={clueText[entry.id] ?? ''}
								oninput={(event) => (clueText[entry.id] = event.currentTarget.value)}
								placeholder="Write the clue…"
							/>
						</label>
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
		grid-template-columns: 2rem minmax(6rem, 0.42fr) 1fr;
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
