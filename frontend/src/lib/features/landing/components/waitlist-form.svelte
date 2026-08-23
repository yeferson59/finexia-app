<script lang="ts">
	import CountdownInline from './countdown-inline.svelte';

	interface Props {
		/** Ancla del formulario. Solo lo lleva el del hero, al que apunta el menú. */
		anchor?: string;
		/** El del CTA final va centrado dentro de su panel. */
		centered?: boolean;
		/** El aviso de spam sobra donde ya se ha leído una vez. */
		note?: boolean;
		/** Texto que precede a la cuenta atrás. */
		countdownLabel?: string;
	}

	let { anchor, centered = false, note = true, countdownLabel = 'Faltan' }: Props = $props();

	let waitlistEmail = $state('');
	let waitlistError = $state(false);
	let waitlistErrorMessage = $state('');
	let waitlistSuccess = $state(false);
	let submitting = $state(false);

	async function submitWaitlist(event: SubmitEvent) {
		event.preventDefault();
		const formEl = event.currentTarget as HTMLFormElement;
		waitlistError = false;
		waitlistErrorMessage = '';
		submitting = true;

		try {
			const res = await fetch('/api/waitlist', {
				method: 'POST',
				headers: { accept: 'application/json' },
				body: new FormData(formEl)
			});
			const data = await res.json();

			if (data.success) {
				waitlistSuccess = true;
			} else {
				waitlistError = true;
				waitlistErrorMessage = data.error ?? 'Ocurrió un error. Inténtalo de nuevo.';
			}
		} catch {
			waitlistError = true;
			waitlistErrorMessage = 'Ocurrió un error. Inténtalo de nuevo.';
		} finally {
			submitting = false;
		}
	}
</script>

<div class="waitlist reveal" class:centered id={anchor}>
	{#if waitlistSuccess}
		<div class="wl-success">
			<span class="check-ico" aria-hidden="true">
				<svg
					width="12"
					height="12"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="3.2"
					stroke-linecap="round"
					stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg
				>
			</span>
			<span>¡Listo! Te avisaremos en cuanto Finexia esté disponible.</span>
		</div>
	{:else}
		<form
			class="wl-form"
			class:error={waitlistError}
			method="POST"
			action="/api/waitlist"
			onsubmit={submitWaitlist}
			novalidate
		>
			<input
				type="email"
				bind:value={waitlistEmail}
				placeholder="tu@email.com"
				autocomplete="email"
				name="email"
				required
				aria-label="Correo electrónico"
			/>
			<button type="submit" class="btn-amber" disabled={submitting}>
				{submitting ? 'Enviando...' : 'Acceso anticipado'}
			</button>
		</form>
		{#if waitlistError && waitlistErrorMessage}
			<p class="wl-error" role="alert">{waitlistErrorMessage}</p>
		{/if}
		<div class="wl-meta">
			{#if note}
				<div class="wl-note">
					<svg
						width="13"
						height="13"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						><rect x="3" y="11" width="18" height="11" rx="2" /><path
							d="M7 11V7a5 5 0 0 1 10 0v4"
						/></svg
					>
					Sin spam. Solo te escribimos el día del lanzamiento.
				</div>
			{/if}
			<CountdownInline label={countdownLabel} />
		</div>
	{/if}
</div>

<style>
	.waitlist {
		width: 100%;
		max-width: 480px;
	}

	.waitlist.centered {
		margin: 0 auto;
	}

	.wl-form {
		display: flex;
		gap: 8px;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		padding: 6px;
		transition:
			border-color 0.2s,
			box-shadow 0.2s;
	}

	.wl-form:focus-within {
		border-color: rgba(212, 145, 42, 0.5);
		box-shadow: 0 0 0 3px rgba(212, 145, 42, 0.07);
	}

	.wl-form.error {
		border-color: var(--red);
	}

	.wl-form input {
		flex: 1;
		background: transparent;
		border: none;
		outline: none;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 15px;
		padding: 0 14px;
		min-width: 0;
	}

	.wl-form input::placeholder {
		color: var(--text-dim);
	}

	/*
	 * Aviso y cuenta atrás comparten fila: la información secundaria del
	 * formulario ocupa una línea en vez de dos bloques apilados.
	 */
	.wl-meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin-top: 12px;
	}

	.waitlist.centered .wl-meta {
		justify-content: center;
		margin-top: 14px;
	}

	.wl-note {
		font-size: 12px;
		color: var(--text-dim);
		display: flex;
		align-items: center;
		gap: 7px;
	}

	.wl-error {
		margin-top: 10px;
		font-size: 12.5px;
		color: var(--red);
	}

	.wl-success {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 20px;
		border-radius: 8px;
		background: rgba(34, 201, 126, 0.06);
		border: 1px solid rgba(34, 201, 126, 0.22);
		font-size: 15px;
		font-weight: 300;
	}

	.check-ico {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--green);
		color: #06140c;
	}

	@media (max-width: 640px) {
		.wl-form {
			flex-direction: column;
		}
		.wl-form input {
			height: 44px;
		}
		.wl-meta {
			flex-direction: column;
			align-items: flex-start;
			gap: 8px;
		}
		.waitlist.centered .wl-meta {
			align-items: center;
		}
	}
</style>
