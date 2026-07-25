<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { Platform } from '$lib/api/types';

	let { platforms, selected = $bindable('') }: { platforms: Platform[]; selected?: string } =
		$props();

	function createNewPlatform() {
		goto(resolve('/dashboard/platforms/add'));
	}
</script>

<section class="form-section">
	<h2 class="section-title">Plataforma de Inversión</h2>
	<div class="form-group">
		<label for="platformId" class="form-label"
			>Selecciona una Plataforma <span class="required">*</span></label
		>
		{#if platforms.length > 0}
			<select id="platformId" bind:value={selected} name="platformId" class="form-select" required>
				<option value="">-- Elige una plataforma --</option>
				{#each platforms as platform (platform.id)}
					<option value={platform.id}>{platform.name}</option>
				{/each}
			</select>
			<p class="field-hint">Selecciona dónde realizarás esta inversión</p>
		{:else}
			<div class="empty-platforms">
				<p class="empty-text">No tienes plataformas registradas</p>
				<button type="button" onclick={createNewPlatform} class="btn-link">
					<svg
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<path d="M12 5v14M5 12h14" />
					</svg>
					Crear tu primera plataforma
				</button>
			</div>
		{/if}
	</div>
</section>

<style>
	.form-section {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 1.75rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1.15rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
		margin-bottom: 1.35rem;
	}

	.form-group:last-child {
		margin-bottom: 0;
	}

	.form-label {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.required {
		color: var(--red);
	}

	.field-hint {
		margin: 0.4rem 0 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.4);
		font-style: italic;
	}

	.empty-platforms {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 1.5rem;
		border-radius: 10px;
		background: var(--surface);
		border: 1px dashed rgba(212, 145, 42, 0.2);
		text-align: center;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(236, 234, 229, 0.6);
		font-weight: 500;
	}

	.btn-link {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.65rem 1.2rem;
		border: none;
		background: var(--border-strong);
		color: var(--amber);
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: var(--font-body);
	}

	.btn-link:hover {
		background: rgba(212, 145, 42, 0.25);
		transform: translateX(2px);
	}

	.form-select {
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.95rem;
		font-family: var(--font-body);
		transition: all 0.3s ease;
	}

	.form-select:focus {
		outline: none;
		border-color: var(--amber);
		box-shadow: 0 0 0 3px var(--border);
	}
</style>
