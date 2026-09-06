<script lang="ts">
	import type { Asset } from '$lib/api/types';

	/**
	 * Alta de un activo desde el buscador, cuando el catálogo no tiene lo que el
	 * usuario está registrando.
	 *
	 * Antes esta situación no tenía salida: crear activos era solo de admin y el
	 * buscador terminaba en "no se encontró nada", así que la única forma de
	 * registrar un instrumento ausente era fabricar un archivo e importarlo. El
	 * activo que se crea aquí lo ve solo quien lo creó, hasta que el operador lo
	 * cure (API §2.8).
	 */
	let {
		ticker,
		oncreated,
		oncancel
	}: {
		ticker: string;
		oncreated: (asset: Asset) => void;
		oncancel: () => void;
	} = $props();

	const ASSET_TYPES = [
		{ value: 'stock', label: 'Acción' },
		{ value: 'etf', label: 'ETF' },
		{ value: 'crypto', label: 'Cripto' },
		{ value: 'bond', label: 'Bono' },
		{ value: 'real_estate', label: 'Bienes raíces' },
		{ value: 'commodity', label: 'Materias primas' },
		{ value: 'cash', label: 'Efectivo' },
		{ value: 'other', label: 'Otro' }
	];

	let name = $state('');
	let assetType = $state('stock');
	let currency = $state('USD');
	let exchange = $state('');
	let submitting = $state(false);
	let error = $state('');

	const normalizedTicker = $derived(ticker.trim().toUpperCase());

	async function submit(event: SubmitEvent) {
		event.preventDefault();

		if (submitting) return;

		submitting = true;
		error = '';

		try {
			const res = await fetch('/api/assets', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					ticker: normalizedTicker,
					name: name.trim() || normalizedTicker,
					assetType,
					currency: currency.trim().toUpperCase(),
					exchange: exchange.trim()
				})
			});

			const json = await res.json().catch(() => null);

			if (!res.ok || !json?.success || !json.data) {
				error = json?.message ?? 'No se pudo crear el activo';
				return;
			}

			oncreated(json.data as Asset);
		} catch {
			error = 'No se pudo conectar. Revisa tu conexión e inténtalo de nuevo.';
		} finally {
			submitting = false;
		}
	}
</script>

<form class="create-asset" onsubmit={submit}>
	<!-- Dice qué se está creando y qué pasa después: es un alta en el catálogo,
	     no la posición, y al terminar se vuelve al formulario con el activo ya
	     elegido. -->
	<p class="create-asset-title">
		Nuevo activo <strong>{normalizedTicker}</strong>. Al crearlo queda elegido y sigues con la
		posición.
	</p>

	<div class="create-asset-grid">
		<div class="create-asset-field create-asset-field-wide">
			<label class="create-asset-label" for="new-asset-name">Nombre</label>
			<input
				id="new-asset-name"
				type="text"
				class="create-asset-input"
				placeholder={normalizedTicker}
				bind:value={name}
				maxlength="255"
			/>
		</div>

		<div class="create-asset-field">
			<label class="create-asset-label" for="new-asset-type">Tipo</label>
			<select id="new-asset-type" class="create-asset-input" bind:value={assetType}>
				{#each ASSET_TYPES as type (type.value)}
					<option value={type.value}>{type.label}</option>
				{/each}
			</select>
		</div>

		<div class="create-asset-field">
			<label class="create-asset-label" for="new-asset-currency">Moneda</label>
			<input
				id="new-asset-currency"
				type="text"
				class="create-asset-input"
				placeholder="USD"
				bind:value={currency}
				maxlength="3"
				required
			/>
		</div>

		<div class="create-asset-field">
			<label class="create-asset-label" for="new-asset-exchange">
				Mercado <span class="create-asset-optional">(opcional)</span>
			</label>
			<input
				id="new-asset-exchange"
				type="text"
				class="create-asset-input"
				placeholder="NASDAQ"
				bind:value={exchange}
				maxlength="100"
			/>
		</div>
	</div>

	{#if error}
		<p class="create-asset-error" role="alert">{error}</p>
	{/if}

	<div class="create-asset-actions">
		<button type="button" class="create-asset-cancel" onclick={oncancel}>Cancelar</button>
		<button type="submit" class="create-asset-submit" disabled={submitting}>
			{submitting ? 'Creando…' : 'Crear activo'}
		</button>
	</div>
</form>

<style>
	/*
	 * El alta de catálogo, dentro del buscador. Sin el borde ámbar de 1,5 px que
	 * lo hacía competir con el total: es un desvío del camino principal, no el
	 * camino.
	 */
	.create-asset {
		margin-top: 0.65rem;
		padding: 1rem;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
		background: var(--panel, rgba(255, 255, 255, 0.04));
	}

	.create-asset-title {
		max-width: 56ch;
		margin: 0 0 0.9rem;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.create-asset-title strong {
		font-family: var(--font-mono);
		font-weight: 500;
		color: var(--text);
	}

	.create-asset-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.85rem;
	}

	.create-asset-field {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		min-width: 0;
	}

	.create-asset-field-wide {
		grid-column: 1 / -1;
	}

	.create-asset-label {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text);
	}

	.create-asset-optional {
		font-weight: 400;
		color: var(--text-dim);
	}

	/* Más apretados que los del formulario que los rodea: son un paso dentro de
	   un campo, y con la misma altura la caja parecía otra pantalla. */
	.create-asset input.create-asset-input,
	.create-asset select.create-asset-input {
		width: 100%;
		padding: 0.55rem 0.7rem;
		border: 1px solid var(--border-strong);
		border-radius: 7px;
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.87rem;
		box-sizing: border-box;
		transition: border-color 0.2s ease;
	}

	.create-asset input.create-asset-input::placeholder {
		color: var(--text-dim);
	}

	.create-asset input.create-asset-input:focus,
	.create-asset select.create-asset-input:focus {
		border-color: var(--amber);
	}

	.create-asset-error {
		margin: 0.85rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid var(--red);
		font-size: 0.82rem;
		line-height: 1.5;
		color: var(--red);
	}

	.create-asset-actions {
		display: flex;
		align-items: center;
		gap: 1.1rem;
		margin-top: 1rem;
	}

	.create-asset-submit {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: none;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.83rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.create-asset-submit:hover:not(:disabled) {
		border-color: var(--text-dim);
		background: var(--surface-2);
	}

	.create-asset-submit:disabled {
		color: var(--text-dim);
		cursor: default;
	}

	.create-asset-cancel {
		order: 1;
		padding: 0;
		border: none;
		background: none;
		color: var(--text-muted);
		font-family: var(--font-body);
		font-size: 0.83rem;
		cursor: pointer;
		transition: color 0.2s ease;
	}

	.create-asset-cancel:hover {
		color: var(--text);
	}

	@media (prefers-reduced-motion: reduce) {
		.create-asset input.create-asset-input,
		.create-asset select.create-asset-input,
		.create-asset-submit,
		.create-asset-cancel {
			transition: none;
		}
	}

	@media (max-width: 520px) {
		.create-asset-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
