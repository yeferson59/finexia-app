<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Card from '$lib/ui/card.svelte';
	import Button from '$lib/ui/button.svelte';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { flash } from '$lib/shared/flash.svelte';
	import { formatCurrency as formatMoney } from '$lib/shared/format/money';
	import PlatformStatsGrid from './platform-stats-grid.svelte';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import { formatSourceType, type Platform } from '../platforms';
	import PlatformEditForm from './platform-edit-form.svelte';
	import PlatformDeleteConfirm from './platform-delete-confirm.svelte';

	let { platform }: { platform: Platform } = $props();

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('es-CO', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	// La moneda del total la informa el backend: es la suma de posiciones
	// compradas en varias monedas, convertida a la de la cuenta, y etiquetarla
	// siempre con "$" le ponía el símbolo equivocado.
	const currency = $derived(platform.displayCurrency || FALLBACK_CURRENCY);

	// Posiciones que entraron al total sin convertir por no haber tasa.
	const unconverted = $derived(platform.positionsUnconverted ?? 0);

	// Ausente no es cero: un backend anterior a estas métricas no dice que la
	// plataforma no haya ganado nada, dice que no lo sabe. Se calla en vez de
	// afirmar.
	const gain = $derived(platform.gainLoss !== undefined ? parseFloat(platform.gainLoss) : null);

	function formatCurrency(value: string): string {
		return privacy.money(formatMoney(parseFloat(value) || 0, currency));
	}

	let isEditing = $state(false);
	let showDeleteConfirm = $state(false);
	const saved = flash(3000);

	function goBack() {
		goto(resolve('/dashboard/platforms'));
	}

	// El tipo y la fecha de alta son una línea de identificación bajo el título,
	// no dos bloques con su etiqueta en versalitas cada uno.
	const identity = $derived(
		`${formatSourceType(platform.sourceType)} · registrada el ${formatDate(platform.createdAt)}`
	);
</script>

<button onclick={goBack} class="btn-back" type="button">
	<svg
		width="16"
		height="16"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
	>
		<path d="M19 12H5M12 19l-7-7 7-7" />
	</svg>
	Plataformas
</button>

<PageHeader title={platform.name} subtitle={identity}>
	{#snippet actions()}
		{#if !platform.isActive}
			<span class="inactive">Inactiva</span>
		{/if}
		<Button type="button" variant="secondary" size="sm" onclick={() => (isEditing = true)}>
			Editar
		</Button>
		<Button type="button" variant="tertiary" size="sm" onclick={() => (showDeleteConfirm = true)}>
			Eliminar
		</Button>
	{/snippet}
</PageHeader>

{#if saved.text}
	<p class="saved">{saved.text}</p>
{/if}

<Modal open={isEditing} title="Editar plataforma" onClose={() => (isEditing = false)} size="lg">
	<PlatformEditForm
		{platform}
		onCancel={() => (isEditing = false)}
		onSaved={() => {
			isEditing = false;
			saved.show('Plataforma actualizada correctamente.');
		}}
	/>
</Modal>

<Card variant="elevated" padding="none">
	<div class="panel-body">
		<PlatformStatsGrid {platform} {unconverted} {gain} {formatCurrency} />

		{#if platform.description}
			<p class="description">{platform.description}</p>
		{/if}
	</div>
</Card>

<Modal
	open={showDeleteConfirm}
	title="Confirmar eliminación"
	onClose={() => (showDeleteConfirm = false)}
	size="sm"
>
	<PlatformDeleteConfirm
		platformName={platform.name}
		onCancel={() => (showDeleteConfirm = false)}
	/>
</Modal>

<style>
	.btn-back {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		margin-bottom: 1.25rem;
		padding: 0;
		border: none;
		background: none;
		color: var(--text-muted);
		font-family: var(--font-body);
		font-size: 0.85rem;
		cursor: pointer;
		transition: color 0.2s ease;
	}

	.btn-back:hover {
		color: var(--amber);
	}

	/* Activa es el caso normal y no lleva marca; el rojo de esta pantalla queda
	   para las pérdidas y para el botón que borra. */
	.inactive {
		padding: 0.2rem 0.5rem;
		border: 1px solid var(--border-strong);
		border-radius: 4px;
		font-size: 0.72rem;
		color: var(--text-muted);
	}

	.saved {
		margin: 0 0 1.5rem;
		padding: 0.7rem 1rem;
		border-left: 2px solid var(--green);
		background: rgba(34, 201, 126, 0.08);
		color: var(--green);
		font-size: 0.875rem;
	}

	.panel-body {
		padding: 1.75rem;
	}

	/* La nota del dueño va bajo las cifras, separada por una línea y sin título:
	   un `<h2>Descripción</h2>` sobre un párrafo es etiquetar lo evidente. */
	.description {
		margin: 1.75rem 0 0;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border);
		max-width: 68ch;
		color: rgba(236, 234, 229, 0.75);
		line-height: 1.65;
		font-size: 0.9rem;
	}
</style>
