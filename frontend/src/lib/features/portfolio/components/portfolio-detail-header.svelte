<script lang="ts">
	/*
	 * Quién es este portafolio y qué se puede hacer con él.
	 *
	 * La vuelta al listado es un enlace y no un botón con una flecha suelta:
	 * dice a dónde lleva, se abre en otra pestaña y el teclado la encuentra.
	 */
	import { resolve } from '$app/paths';

	let {
		name,
		description,
		riskName,
		holdingsCount,
		portfolioId,
		onEdit
	}: {
		name: string;
		description?: string | null;
		riskName?: string;
		holdingsCount: number;
		portfolioId: string;
		onEdit: () => void;
	} = $props();
</script>

<header class="head">
	<a class="back" href={resolve('/dashboard/portfolios')}>Volver a portafolios</a>

	<div class="row">
		<div class="who">
			<h1>{name}</h1>
			{#if description}
				<p class="description">{description}</p>
			{/if}
			<p class="meta">
				{riskName ? `Riesgo ${riskName.toLowerCase()}` : 'Sin nivel de riesgo'},
				{holdingsCount === 1 ? '1 activo' : `${holdingsCount} activos`}
			</p>
		</div>

		<div class="actions">
			<button type="button" class="edit" onclick={onEdit}>Editar</button>
			<a class="add" href={resolve('/dashboard/portfolios/[id]/add', { id: portfolioId })}>
				Añadir activo
			</a>
		</div>
	</div>
</header>

<style>
	.head {
		margin-bottom: 2rem;
		padding-bottom: 1.75rem;
		border-bottom: 1px solid var(--border);
	}

	.back {
		display: inline-block;
		margin-bottom: 1.1rem;
		font-size: 0.82rem;
		color: var(--text-muted);
		text-decoration: none;
	}

	.back::before {
		content: '←';
		margin-right: 0.4rem;
	}

	.back:hover {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.who {
		min-width: 0;
	}

	h1 {
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(2rem, 4vw, 2.75rem);
		font-weight: 300;
		line-height: 1.05;
		letter-spacing: -0.02em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.description {
		max-width: 58ch;
		margin: 0.6rem 0 0;
		font-size: 0.95rem;
		font-weight: 300;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.meta {
		margin: 0.35rem 0 0;
		font-size: 0.85rem;
		color: var(--text-dim);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		flex-shrink: 0;
	}

	.edit,
	.add {
		display: inline-flex;
		align-items: center;
		padding: 0.6rem 1.15rem;
		border-radius: 9px;
		font-family: var(--font-body);
		font-size: 0.88rem;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.edit {
		border: 1px solid var(--border-strong);
		background: none;
		color: var(--text);
	}

	.edit:hover {
		border-color: var(--text-dim);
		background: var(--panel);
	}

	.add {
		border: 1px solid var(--amber);
		background: var(--amber);
		color: #0d0800;
	}

	.add:hover {
		border-color: var(--amber-light);
		background: var(--amber-light);
	}

	@media (prefers-reduced-motion: reduce) {
		.edit,
		.add {
			transition: none;
		}
	}
</style>
