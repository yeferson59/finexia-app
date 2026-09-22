<script lang="ts">
	/*
	 * El aporte con Bold, en la banda verde: es lo único que se hace en la
	 * página. Los montos van en el Archivo ancho de los titulares, porque la
	 * decisión aquí es una cifra.
	 *
	 * Al enviar, el servidor crea la orden y la firma con la llave secreta
	 * (acción `pagar`); el navegador carga la librería de Bold y abre la
	 * pasarela con esa orden. Bold devuelve a quien paga a esta misma página,
	 * que consulta el resultado.
	 *
	 * Sin llaves configuradas, la banda dice a quién escribir en vez de enseñar
	 * un formulario que no llevaría a ninguna parte.
	 */
	import { enhance } from '$app/forms';
	import type { SubmitFunction } from '@sveltejs/kit';
	import { formatCurrency } from '$lib/shared/format/money';
	import { AMOUNT_PRESETS, DEFAULT_AMOUNT, MAX_AMOUNT, MIN_AMOUNT, parsePesos } from '../support';
	import { openBoldCheckout, type BoldCheckoutOptions } from '../bold-checkout';

	interface Props {
		enabled: boolean;
		/** El error que devolvió la acción `pagar`, si lo hubo. */
		error?: string;
		contactEmail: string;
	}

	let { enabled, error = '', contactEmail }: Props = $props();

	let preset = $state(String(DEFAULT_AMOUNT));
	let custom = $state('');
	let pending = $state(false);
	let clientError = $state('');
	let customInput = $state<HTMLInputElement>();

	const cop = (value: number) => formatCurrency(value, 'COP');

	const amount = $derived(preset === 'otro' ? parsePesos(custom) : Number(preset));
	const inRange = $derived(amount >= MIN_AMOUNT && amount <= MAX_AMOUNT);
	const cta = $derived(inRange ? `Aportar ${cop(amount)} con Bold` : 'Aportar con Bold');
	const message = $derived(clientError || error);

	$effect(() => {
		if (preset === 'otro') customInput?.focus();
	});

	const submit: SubmitFunction = () => {
		pending = true;
		clientError = '';
		return async ({ result, update }) => {
			const checkout =
				result.type === 'success'
					? (result.data?.checkout as BoldCheckoutOptions | undefined)
					: undefined;
			if (checkout) {
				try {
					// La pasarela navega fuera de la página: el botón sigue ocupado.
					await openBoldCheckout(checkout);
				} catch {
					clientError =
						'No se pudo abrir la pasarela de Bold. Revisa tu conexión e inténtalo de nuevo.';
					pending = false;
				}
				return;
			}
			pending = false;
			await update({ reset: false });
		};
	};
</script>

<!-- Al volver con «Atrás» desde Bold, la página sale de la caché con el botón
     aún ocupado. -->
<svelte:window onpageshow={() => (pending = false)} />

<section class="bold lp-forest" aria-labelledby="bold-title">
	<div class="lp-wrap">
		<h2 id="bold-title">¿Cuánto quieres aportar?</h2>

		{#if enabled}
			<form method="POST" action="?/pagar" use:enhance={submit}>
				<fieldset>
					<legend class="sr-only">Monto del aporte</legend>
					<div class="amounts">
						{#each AMOUNT_PRESETS as value (value)}
							<label class="amount">
								<input type="radio" name="preset" value={String(value)} bind:group={preset} />
								<span>{cop(value)}</span>
							</label>
						{/each}
						<label class="amount other">
							<input type="radio" name="preset" value="otro" bind:group={preset} />
							<span>Otro monto</span>
						</label>
					</div>
				</fieldset>

				{#if preset === 'otro'}
					<div class="custom">
						<label for="support-custom">Monto en pesos colombianos</label>
						<div class="peso-field">
							<span aria-hidden="true">$</span>
							<input
								id="support-custom"
								name="custom"
								inputmode="numeric"
								autocomplete="off"
								placeholder="15.000"
								aria-describedby="support-custom-hint"
								bind:value={custom}
								bind:this={customInput}
							/>
						</div>
						<p id="support-custom-hint">Desde {cop(MIN_AMOUNT)} hasta {cop(MAX_AMOUNT)}.</p>
					</div>
				{/if}

				{#if message}
					<p class="error" role="alert">{message}</p>
				{/if}

				<div class="submit">
					<button type="submit" class="lp-btn pay" disabled={pending}>
						{pending ? 'Abriendo Bold…' : cta}
					</button>
					<p class="note">
						Pagas en la pasarela de Bold con el medio que prefieras y al terminar vuelves aquí. No
						necesitas cuenta en Bold ni en Finexia.
					</p>
				</div>
				<noscript>
					<p class="note">Para abrir la pasarela de Bold necesitas JavaScript activo.</p>
				</noscript>
			</form>
		{:else}
			<p class="unavailable">
				Los aportes con Bold se abren aquí en los próximos días. Mientras tanto, escribe a
				<a href="mailto:{contactEmail}">{contactEmail}</a>.
			</p>
		{/if}
	</div>
</section>

<style>
	.bold {
		padding-block: 80px 88px;
	}

	h2 {
		margin: 0 0 40px;
		font-size: clamp(28px, 3.6vw, 44px);
		font-stretch: 112%;
		font-weight: 620;
		line-height: 1.05;
		letter-spacing: -0.02em;
	}

	fieldset {
		margin: 0;
		padding: 0;
		border: 0;
	}

	/*
	 * Los montos, en fila y a tamaño de titular: se elige una cifra, no una
	 * tarjeta. El elegido se enciende en papel y lleva el subrayado ámbar de
	 * la marca; los demás ceden en el verde claro.
	 */
	.amounts {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 40px;
		padding-bottom: 28px;
		border-bottom: 1px solid var(--lp-forest-rule);
	}

	.amount {
		position: relative;
		padding-bottom: 6px;
		border-bottom: 3px solid transparent;
		color: var(--lp-forest-ink-2);
		font-size: clamp(30px, 4.2vw, 54px);
		font-stretch: 122%;
		font-weight: 640;
		line-height: 1;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
		cursor: pointer;
		transition: color 0.15s ease;
	}

	.amount.other {
		align-self: flex-end;
		font-size: clamp(20px, 2.4vw, 28px);
		font-stretch: 110%;
		letter-spacing: -0.01em;
	}

	.amount:hover,
	.amount:has(input:checked) {
		color: var(--lp-forest-ink);
	}

	.amount:has(input:checked) {
		border-bottom-color: var(--lp-jubilacion);
	}

	.amount:has(input:focus-visible) {
		outline: 2px solid var(--lp-forest-ink);
		outline-offset: 6px;
	}

	.amount input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}

	.custom {
		max-width: 360px;
		margin-top: 28px;
	}

	.custom label {
		font-size: var(--lp-fs-sm);
		font-weight: 600;
	}

	.peso-field {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-top: 8px;
		padding: 0 16px;
		border: 1px solid var(--lp-forest-ink-2);
		border-radius: 6px;
		font-size: 24px;
		font-stretch: 115%;
		font-weight: 600;
	}

	.peso-field:focus-within {
		outline: 2px solid var(--lp-forest-ink);
		outline-offset: 2px;
	}

	.peso-field input {
		width: 100%;
		min-height: 56px;
		border: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		font-variant-numeric: tabular-nums;
		outline: none;
	}

	.peso-field input::placeholder {
		color: var(--lp-forest-ink-2);
		opacity: 0.6;
	}

	.custom p {
		margin: 8px 0 0;
		font-size: var(--lp-fs-sm);
		color: var(--lp-forest-ink-2);
	}

	.error {
		margin: 24px 0 0;
		padding: 12px 16px;
		border-left: 3px solid #f2b8b5;
		background: rgba(242, 184, 181, 0.12);
		color: var(--lp-forest-ink);
	}

	.submit {
		display: flex;
		align-items: center;
		gap: 16px 32px;
		margin-top: 32px;
	}

	.pay {
		min-width: 18rem;
		min-height: 56px;
		border-color: var(--lp-forest-ink);
		background: var(--lp-forest-ink);
		color: var(--lp-forest);
		font-size: var(--lp-fs-body);
	}

	.pay:hover {
		background: #ffffff;
	}

	.note {
		max-width: 48ch;
		margin: 0;
		font-size: var(--lp-fs-sm);
		color: var(--lp-forest-ink-2);
	}

	.unavailable {
		max-width: 56ch;
		margin: 0;
		font-size: var(--lp-fs-lead);
		color: var(--lp-forest-ink-2);
	}

	.unavailable a {
		color: var(--lp-forest-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
		white-space: nowrap;
	}

	@media (max-width: 760px) {
		.bold {
			padding-block: 56px;
		}
		h2 {
			margin-bottom: 28px;
		}
		.amounts {
			gap: 12px 28px;
		}
		.submit {
			flex-direction: column;
			align-items: stretch;
		}
		.pay {
			min-width: 0;
			width: 100%;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.amount {
			transition: none;
		}
	}
</style>
