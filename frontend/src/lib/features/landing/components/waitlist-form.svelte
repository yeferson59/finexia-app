<script lang="ts">
	/*
	 * El alta en la lista de espera. Lo usan el hero y el cierre.
	 *
	 * El botón decía «Acceso anticipado» mientras la barra decía «Unirme a la
	 * lista»: dos nombres para la misma acción. Ahora es uno solo, y lo que
	 * contesta al enviarlo usa el mismo verbo.
	 */
	import { onMount } from 'svelte';
	import { LAUNCH_DATE, daysUntil, launchCountdownText } from '../landing';

	interface Props {
		/** Ancla del formulario. Solo lo lleva el del hero, al que apunta el menú. */
		anchor?: string;
		/** Sobre el verde del cierre, los colores se invierten. */
		tone?: 'paper' | 'forest';
	}

	let { anchor, tone = 'paper' }: Props = $props();

	const inputId = $props.id();

	let email = $state('');
	let errorMessage = $state('');
	let success = $state(false);
	let submitting = $state(false);

	/*
	 * La página se prerenderiza: la cifra del HTML es la del día del build. Hasta
	 * que hidrata se enseña la frase sin días, que nunca es falsa.
	 */
	let days = $state(0);

	onMount(() => {
		days = daysUntil(new Date(LAUNCH_DATE).getTime(), Date.now());
	});

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const formEl = event.currentTarget as HTMLFormElement;
		errorMessage = '';

		if (!email.trim()) {
			errorMessage = 'Escribe tu correo para unirte a la lista.';
			return;
		}

		submitting = true;
		try {
			const res = await fetch('/api/waitlist', {
				method: 'POST',
				headers: { accept: 'application/json' },
				body: new FormData(formEl)
			});
			const data = await res.json();
			if (data.success) {
				success = true;
			} else {
				errorMessage = data.error ?? 'No pudimos guardar tu correo. Inténtalo de nuevo.';
			}
		} catch {
			errorMessage = 'No pudimos guardar tu correo. Revisa tu conexión e inténtalo de nuevo.';
		} finally {
			submitting = false;
		}
	}
</script>

<div class="waitlist" class:forest={tone === 'forest'} id={anchor}>
	{#if success}
		<p class="done" role="status">
			Te has unido a la lista. Te escribimos el 1 de octubre, cuando abramos.
		</p>
	{:else}
		<form
			class="form"
			class:invalid={errorMessage}
			method="POST"
			action="/api/waitlist"
			onsubmit={submit}
			novalidate
		>
			<label class="lp-sr-only" for={inputId}>Correo electrónico</label>
			<input
				id={inputId}
				type="email"
				name="email"
				bind:value={email}
				placeholder="tu@correo.com"
				autocomplete="email"
				required
				aria-invalid={errorMessage ? 'true' : undefined}
				aria-describedby={errorMessage ? `${inputId}-error` : `${inputId}-note`}
			/>
			<button type="submit" class="lp-btn" disabled={submitting}>
				{submitting ? 'Uniéndote…' : 'Unirme a la lista'}
			</button>
		</form>

		{#if errorMessage}
			<p class="error" id="{inputId}-error" role="alert">{errorMessage}</p>
		{/if}

		<p class="note" id="{inputId}-note">
			{launchCountdownText(days)} Solo te escribiremos ese día.
		</p>
	{/if}
</div>

<style>
	.waitlist {
		width: 100%;
		max-width: 520px;
	}

	.form {
		display: flex;
		gap: 6px;
		padding: 5px;
		border: 1px solid var(--lp-ink);
		border-radius: 10px;
		background: #fff;
	}

	.form.invalid {
		border-color: #b3261e;
	}

	/* El foco lo marca el borde de la caja entera, no el del campo suelto. */
	.form:focus-within {
		outline: 2px solid var(--lp-ink);
		outline-offset: 2px;
	}

	.form input {
		flex: 1;
		min-width: 0;
		padding: 0 14px;
		border: none;
		background: transparent;
		color: var(--lp-ink);
		font-family: var(--lp-font);
		font-size: 16px;
	}

	.form input:focus,
	.form input:focus-visible {
		outline: none;
		box-shadow: none;
	}

	.form input::placeholder {
		color: #6b7a75;
	}

	.note,
	.error,
	.done {
		margin: 12px 0 0;
		font-size: var(--lp-fs-sm);
		line-height: 1.5;
	}

	.note {
		color: var(--lp-ink-2);
	}

	.error {
		color: #b3261e;
	}

	.done {
		margin: 0;
		padding: 14px 0 14px 16px;
		border-left: 3px solid var(--lp-cripto);
		font-size: var(--lp-fs-lead);
		color: var(--lp-ink);
	}

	/* Sobre el bosque: la caja pasa a papel y el botón, a ámbar. */
	.forest .form {
		border-color: transparent;
		background: var(--lp-forest-ink);
	}

	.forest .form:focus-within {
		outline-color: var(--lp-forest-ink);
	}

	.forest .lp-btn {
		background: var(--lp-jubilacion);
		color: var(--lp-ink);
	}

	.forest .lp-btn:hover {
		background: #e3a03a;
	}

	.forest .note {
		color: var(--lp-forest-ink-2);
	}

	.forest .error {
		color: #ffb4a9;
	}

	.forest .done {
		border-left-color: var(--lp-jubilacion);
		color: var(--lp-forest-ink);
	}

	@media (max-width: 520px) {
		.form {
			flex-direction: column;
			padding: 6px;
		}
		.form input {
			min-height: 48px;
		}
	}
</style>
