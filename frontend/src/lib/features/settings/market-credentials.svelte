<script lang="ts">
	/**
	 * Claves de proveedor de datos de mercado (BYO-key).
	 *
	 * La aplicación no tiene claves de proveedor: cada usuario aporta la suya y
	 * los precios que trae solo los ve quien la puso. Esta sección es el único
	 * sitio donde se introducen.
	 *
	 * La clave nunca vuelve del servidor: se guarda cifrada y de ella solo se
	 * recibe `last4`. Por eso el campo jamás se prellena, y cambiarla significa
	 * escribirla entera de nuevo.
	 */
	import { enhance } from '$app/forms';
	import Card from '$lib/ui/card.svelte';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import Badge from '$lib/ui/badge.svelte';
	import type { MarketCredential, MarketProvider } from '$lib/api/types';

	interface Props {
		credentials: MarketCredential[];
		/** `form` de la página, para leer el resultado de las acciones. */
		form: Record<string, unknown> | null;
	}

	let { credentials, form }: Props = $props();

	const PROVIDERS: {
		id: MarketProvider;
		name: string;
		signupUrl: string;
		hint: string;
	}[] = [
		{
			id: 'finnhub',
			name: 'Finnhub',
			signupUrl: 'https://finnhub.io/register',
			hint: 'Recomendado: su plan gratuito permite 60 consultas por minuto.'
		},
		{
			id: 'alphavantage',
			name: 'Alpha Vantage',
			signupUrl: 'https://www.alphavantage.co/support/#api-key',
			hint: 'Su plan gratuito permite 25 consultas al día, así que se usa como respaldo.'
		}
	];

	const byProvider = $derived(new Map(credentials.map((c) => [c.provider, c])));

	// Un campo por proveedor. Nunca se rellena con nada: no hay valor que leer.
	let keyInputs = $state<Record<string, string>>({ finnhub: '', alphavantage: '' });
	let savingProvider = $state<string | null>(null);
	let verifyingProvider = $state<string | null>(null);
	let deletingProvider = $state<string | null>(null);
	let syncing = $state(false);

	const STATUS: Record<
		MarketCredential['status'],
		{ tone: 'success' | 'danger' | 'warning'; label: string; note: string }
	> = {
		active: { tone: 'success', label: 'Activa', note: 'La clave funciona.' },
		invalid: {
			tone: 'danger',
			label: 'No válida',
			note: 'El proveedor rechazó esta clave. Vuelve a introducirla o genera otra.'
		},
		rate_limited: {
			tone: 'warning',
			label: 'Sin cuota',
			note: 'La cuota del plan se agotó. Se reintentará en la próxima sincronización.'
		}
	};

	function formatDate(iso: string | null): string {
		if (!iso) return 'nunca';
		return new Intl.DateTimeFormat('es', { dateStyle: 'short', timeStyle: 'short' }).format(
			new Date(iso)
		);
	}

	function errorFor(provider: MarketProvider): string | null {
		if (form?.marketProvider !== provider) return null;
		return (form?.marketError as string) ?? null;
	}

	function successFor(provider: MarketProvider): string | null {
		if (form?.marketProvider !== provider || !form?.marketSuccess) return null;
		return (form?.marketMessage as string) ?? null;
	}

	const hasAnyKey = $derived(credentials.length > 0);
</script>

<Card variant="elevated" padding="none">
	<div class="section" id="datos-de-mercado">
		<h2 class="section-title">Datos de mercado</h2>

		<p class="hint">
			Finexia no consulta precios con claves propias: usa la tuya, de modo que la cuota y los datos
			son tuyos. Los precios que trae tu clave solo los ves tú.
		</p>
		<p class="hint privacy">
			Tu clave se guarda cifrada y no se puede volver a leer, ni siquiera desde aquí: solo se
			muestran sus cuatro últimos caracteres. Para cambiarla, introduce una nueva.
		</p>

		{#if !hasAnyKey}
			<p class="feedback warning">
				Sin ninguna clave configurada tus posiciones se valoran a su precio de compra, no a precio
				de mercado.
			</p>
		{/if}

		<div class="providers">
			{#each PROVIDERS as provider (provider.id)}
				{@const stored = byProvider.get(provider.id)}
				{@const status = stored ? STATUS[stored.status] : null}
				{@const error = errorFor(provider.id)}
				{@const success = successFor(provider.id)}

				<div class="provider">
					<div class="provider-head">
						<div class="provider-id">
							<span class="provider-name">{provider.name}</span>
							{#if stored && status}
								<Badge tone={status.tone}>{status.label}</Badge>
							{:else}
								<Badge tone="neutral">Sin configurar</Badge>
							{/if}
						</div>
						<!-- eslint-disable svelte/no-navigation-without-resolve -- resolve() es para rutas internas; estas salen al sitio del proveedor -->
						<a
							class="provider-link"
							href={provider.signupUrl}
							target="_blank"
							rel="noopener noreferrer"
						>
							Obtener una clave
						</a>
						<!-- eslint-enable svelte/no-navigation-without-resolve -->
					</div>

					<p class="provider-hint">{provider.hint}</p>

					{#if stored && status}
						<p class="provider-state">
							Clave <code>····{stored.last4}</code> · verificada {formatDate(stored.lastVerifiedAt)}
						</p>
						<p class="provider-state-note" class:is-problem={stored.status !== 'active'}>
							{status.note}
						</p>
					{/if}

					<form
						method="POST"
						action="?/saveMarketKey"
						use:enhance={() => {
							savingProvider = provider.id;
							return async ({ update }) => {
								savingProvider = null;
								// reset:false conserva el resto del formulario de ajustes;
								// el campo de la clave se limpia abajo, a mano.
								await update({ reset: false });
								keyInputs[provider.id] = '';
							};
						}}
					>
						<input type="hidden" name="provider" value={provider.id} />
						<div class="key-row">
							<Input
								label={stored ? 'Reemplazar clave' : 'Clave de API'}
								type="password"
								name="apiKey"
								autocomplete="off"
								placeholder="Pega aquí tu clave"
								bind:value={keyInputs[provider.id]}
								required
							/>
							<Button type="submit" loading={savingProvider === provider.id}>
								{stored ? 'Reemplazar' : 'Guardar'}
							</Button>
						</div>
					</form>

					{#if error}
						<p class="feedback error">{error}</p>
					{/if}
					{#if success}
						<p class="feedback success">{success}</p>
					{/if}

					{#if stored}
						<div class="provider-actions">
							<form
								method="POST"
								action="?/verifyMarketKey"
								use:enhance={() => {
									verifyingProvider = provider.id;
									return async ({ update }) => {
										verifyingProvider = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="provider" value={provider.id} />
								<Button
									type="submit"
									variant="secondary"
									size="sm"
									loading={verifyingProvider === provider.id}
								>
									Verificar
								</Button>
							</form>
							<form
								method="POST"
								action="?/deleteMarketKey"
								use:enhance={() => {
									deletingProvider = provider.id;
									return async ({ update }) => {
										deletingProvider = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="provider" value={provider.id} />
								<Button
									type="submit"
									variant="ghost"
									size="sm"
									loading={deletingProvider === provider.id}
								>
									Eliminar
								</Button>
							</form>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		{#if hasAnyKey}
			<div class="sync-row">
				<div>
					<p class="sync-title">Sincronizar ahora</p>
					<p class="hint">
						Se actualizan los precios de los activos que tienes, con tu clave. Si no, se hace cada
						día automáticamente.
					</p>
				</div>
				<form
					method="POST"
					action="?/syncMarketData"
					use:enhance={() => {
						syncing = true;
						return async ({ update }) => {
							syncing = false;
							await update({ reset: false });
						};
					}}
				>
					<Button type="submit" loading={syncing}>Sincronizar</Button>
				</form>
			</div>

			{#if form?.marketSyncError}
				<p class="feedback error">{form.marketSyncError}</p>
			{/if}
			{#if form?.marketSyncSuccess}
				<p class="feedback success">
					{form.marketSyncCount} precio{form.marketSyncCount === 1 ? '' : 's'} actualizado{form.marketSyncCount ===
					1
						? ''
						: 's'}{form.marketSyncRateCount
						? ` y ${form.marketSyncRateCount} tasa${form.marketSyncRateCount === 1 ? '' : 's'} de cambio`
						: ''}.
				</p>
			{/if}
		{/if}
	</div>
</Card>

<style>
	.section {
		padding: 1.75rem;
	}

	.section-title {
		margin: 0 0 0.75rem;
		font-size: 1.05rem;
		font-weight: 600;
		color: var(--color-text-primary, #f5f5f5);
	}

	.hint {
		margin: 0 0 0.5rem;
		font-size: 0.85rem;
		line-height: 1.5;
		color: var(--color-text-muted, #9a9a9a);
	}

	.privacy {
		margin-bottom: 1.25rem;
	}

	.providers {
		display: grid;
		gap: 1.25rem;
	}

	.provider {
		padding: 1.25rem;
		border: 1px solid var(--color-border, #2a2a2a);
		border-radius: 0.75rem;
		background: var(--color-surface-subtle, rgb(255 255 255 / 2%));
	}

	.provider-head {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		align-items: center;
		justify-content: space-between;
	}

	.provider-id {
		display: flex;
		gap: 0.6rem;
		align-items: center;
	}

	.provider-name {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--color-text-primary, #f5f5f5);
	}

	.provider-link {
		font-size: 0.8rem;
		color: var(--color-accent, #d4af37);
		text-decoration: none;
	}

	.provider-link:hover,
	.provider-link:focus-visible {
		text-decoration: underline;
	}

	.provider-hint {
		margin: 0.5rem 0 0;
		font-size: 0.8rem;
		color: var(--color-text-muted, #9a9a9a);
	}

	.provider-state {
		margin: 0.75rem 0 0.15rem;
		font-size: 0.8rem;
		color: var(--color-text-secondary, #c4c4c4);
	}

	.provider-state code {
		padding: 0.1rem 0.35rem;
		border-radius: 0.25rem;
		background: rgb(255 255 255 / 6%);
		font-size: 0.78rem;
	}

	.provider-state-note {
		margin: 0 0 0.5rem;
		font-size: 0.78rem;
		color: var(--color-text-muted, #9a9a9a);
	}

	.provider-state-note.is-problem {
		color: var(--color-warning, #e0a800);
	}

	.key-row {
		display: flex;
		gap: 0.75rem;
		align-items: flex-end;
		margin-top: 0.75rem;
	}

	.key-row :global(.input-wrapper) {
		flex: 1;
	}

	.provider-actions {
		display: flex;
		gap: 0.5rem;
		margin-top: 0.75rem;
	}

	.sync-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: center;
		justify-content: space-between;
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--color-border, #2a2a2a);
	}

	.sync-title {
		margin: 0 0 0.2rem;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--color-text-primary, #f5f5f5);
	}

	.sync-row .hint {
		margin: 0;
		max-width: 42ch;
	}

	.feedback {
		margin: 0.75rem 0 0;
		font-size: 0.82rem;
	}

	.feedback.error {
		color: var(--color-danger, #e05260);
	}

	.feedback.success {
		color: var(--color-success, #4caf7d);
	}

	.feedback.warning {
		margin-bottom: 1rem;
		color: var(--color-warning, #e0a800);
	}

	@media (max-width: 640px) {
		.section {
			padding: 1.25rem;
		}

		.key-row {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
