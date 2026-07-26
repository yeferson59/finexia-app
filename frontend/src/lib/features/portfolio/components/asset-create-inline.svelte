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
	<p class="create-asset-title">
		Crear <strong>{normalizedTicker}</strong>
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
			{submitting ? 'Creando…' : 'Crear y seleccionar'}
		</button>
	</div>
</form>

<style>
	.create-asset {
		margin-top: 0.6rem;
		padding: 0.9rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 10px;
		background: var(--surface);
	}

	.create-asset-title {
		margin: 0 0 0.75rem;
		font-size: 0.88rem;
		color: rgba(236, 234, 229, 0.75);
	}

	.create-asset-title strong {
		font-family: var(--font-mono);
		color: var(--amber);
	}

	.create-asset-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.7rem;
	}

	.create-asset-field {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		min-width: 0;
	}

	.create-asset-field-wide {
		grid-column: 1 / -1;
	}

	.create-asset-label {
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.4px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.5);
	}

	.create-asset-optional {
		text-transform: none;
		letter-spacing: 0;
		font-weight: 400;
	}

	.create-asset-input {
		width: 100%;
		padding: 0.6rem 0.75rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		transition: all 0.3s ease;
	}

	.create-asset-input:focus {
		outline: none;
		border-color: var(--amber);
		box-shadow: 0 0 0 3px var(--border);
	}

	.create-asset-error {
		margin: 0.7rem 0 0;
		font-size: 0.82rem;
		color: #e5736b;
	}

	.create-asset-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 0.9rem;
	}

	.create-asset-cancel,
	.create-asset-submit {
		padding: 0.5rem 0.9rem;
		border-radius: 8px;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.create-asset-cancel {
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		background: transparent;
		color: rgba(236, 234, 229, 0.7);
	}

	.create-asset-cancel:hover {
		color: var(--text);
	}

	.create-asset-submit {
		border: 1.5px solid var(--amber);
		background: rgba(212, 145, 42, 0.15);
		color: var(--amber);
	}

	.create-asset-submit:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	@media (max-width: 480px) {
		.create-asset-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
