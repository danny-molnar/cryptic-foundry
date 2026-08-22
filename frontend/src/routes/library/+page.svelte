<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import { deleteDraft, listDrafts, type DraftSummary } from '$lib/api';
	import { onMount } from 'svelte';

	let drafts = $state<DraftSummary[]>([]);
	let loading = $state(true);
	let message = $state('');

	onMount(() => void refresh());

	async function refresh() {
		loading = true;
		try {
			drafts = await listDrafts();
			message = '';
		} catch (error) {
			message = error instanceof Error ? error.message : 'Could not load drafts';
		} finally {
			loading = false;
		}
	}

	async function remove(draft: DraftSummary) {
		if (!confirm(`Delete “${draft.title}”? This cannot be undone.`)) return;
		try {
			await deleteDraft(draft.id);
			drafts = drafts.filter((item) => item.id !== draft.id);
		} catch (error) {
			message = error instanceof Error ? error.message : 'Could not delete draft';
		}
	}
</script>

<svelte:head><title>Saved drafts · Cryptic Workshop</title></svelte:head>

<main>
	<header>
		<div>
			<p>Cryptic Workshop</p>
			<h1>Saved drafts</h1>
		</div>
		<a href={resolve('/')}>+ New puzzle</a>
	</header>

	{#if message}<div class="message">{message}</div>{/if}
	{#if loading}
		<p class="empty">Loading your workshop…</p>
	{:else if drafts.length === 0}
		<section class="empty">
			<h2>No persistent drafts yet.</h2>
			<p>Start a puzzle, then choose Save draft to keep it here.</p>
			<a href={resolve('/')}>Open the editor</a>
		</section>
	{:else}
		<section class="drafts">
			{#each drafts as draft (draft.id)}
				<article>
					<div class="grid-mark" aria-hidden="true">{draft.rows}<span>×</span>{draft.cols}</div>
					<div class="details">
						<p>{draft.author || 'Unknown setter'}</p>
						<h2>{draft.title}</h2>
						<time datetime={draft.updatedAt}
							>Updated {new Date(draft.updatedAt).toLocaleString()}</time
						>
					</div>
					<div class="actions">
						<a
							href={resolve('/')}
							onclick={(event) => {
								event.preventDefault();
								void goto(resolve(`/?id=${encodeURIComponent(draft.id)}` as '/'));
							}}>Continue editing</a
						>
						<button onclick={() => remove(draft)}>Delete</button>
					</div>
				</article>
			{/each}
		</section>
	{/if}
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
	main {
		min-height: 100vh;
		padding: clamp(1.5rem, 5vw, 5rem);
		background: linear-gradient(115deg, #fffdf6 0 38%, transparent 38%);
	}
	header {
		max-width: 1050px;
		margin: 0 auto 3rem;
		display: flex;
		align-items: end;
		justify-content: space-between;
		border-bottom: 1px solid #aaa292;
		padding-bottom: 1.5rem;
	}
	header p {
		margin: 0 0 0.35rem;
		color: #a23d2e;
		font:
			700 0.72rem system-ui,
			sans-serif;
		letter-spacing: 0.16em;
		text-transform: uppercase;
	}
	h1 {
		margin: 0;
		font-size: clamp(2.5rem, 7vw, 5.5rem);
		font-weight: 500;
		letter-spacing: -0.04em;
	}
	a {
		color: #a23d2e;
		font:
			700 0.75rem system-ui,
			sans-serif;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}
	.drafts {
		max-width: 1050px;
		margin: 0 auto;
		display: grid;
		gap: 1rem;
	}
	article {
		display: grid;
		grid-template-columns: 90px 1fr auto;
		gap: 1.5rem;
		align-items: center;
		padding: 1.25rem;
		background: rgba(255, 255, 255, 0.66);
		border: 1px solid #c8c0b0;
	}
	.grid-mark {
		display: grid;
		place-items: center;
		aspect-ratio: 1;
		background: #263c34;
		color: #fff;
		font:
			700 1.2rem system-ui,
			sans-serif;
	}
	.grid-mark span {
		color: #f0ca83;
	}
	.details p {
		margin: 0 0 0.25rem;
		color: #777369;
		font:
			700 0.66rem system-ui,
			sans-serif;
		text-transform: uppercase;
		letter-spacing: 0.1em;
	}
	.details h2 {
		margin: 0 0 0.4rem;
		font-size: 1.55rem;
		font-weight: 500;
	}
	time {
		color: #777369;
		font:
			0.72rem system-ui,
			sans-serif;
	}
	.actions {
		display: grid;
		gap: 0.65rem;
		text-align: right;
	}
	button {
		padding: 0;
		border: 0;
		background: none;
		color: #777369;
		cursor: pointer;
		font:
			650 0.7rem system-ui,
			sans-serif;
		text-decoration: underline;
	}
	.empty,
	.message {
		max-width: 1050px;
		margin: 4rem auto;
	}
	.empty h2 {
		font-size: 2rem;
		font-weight: 500;
	}
	@media (max-width: 650px) {
		article {
			grid-template-columns: 65px 1fr;
		}
		.actions {
			grid-column: 1 / -1;
			display: flex;
			justify-content: space-between;
		}
	}
</style>
