<script lang="ts">
	/** Los correos que Finexia envía: alertas de actividad y resumen semanal. */
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { untrack } from 'svelte';
	import Button from '$lib/ui/button.svelte';
	import Checkbox from '$lib/ui/checkbox.svelte';
	import NotificationSection from './notification-section.svelte';
	import { actionError, actionSucceeded, type ActionForm } from '$lib/shared/form';
	import type { UserPreferences } from '$lib/api/types';

	interface Props {
		preferences: UserPreferences;
		/** Quién recibe los correos; de la sesión, no de las preferencias. */
		user: App.Locals['user'];
		form: ActionForm;
	}

	let { preferences, user, form }: Props = $props();

	// Seeded from the server once; the user toggles locally from there.
	let emailAlerts = $state(untrack(() => preferences.emailAlerts));
	let weeklySummary = $state(untrack(() => preferences.weeklySummary));
	let prefsLoading = $state(false);

	const prefsSuccess = $derived(actionSucceeded(form, 'updatePreferences'));
	const prefsError = $derived(actionError(form, 'updatePreferences'));
</script>

<NotificationSection title="Correo" description="Lo que Finexia te manda a tu bandeja de entrada.">
	<!--
		A dónde llegan, antes de decidir qué llega: es lo único que esta página
		sabe y no se puede averiguar desde ninguna otra, y sin ello los dos
		interruptores de abajo son una preferencia en el vacío.

		Y si la dirección no está verificada, ese es el dato de la página: da igual
		lo que se marque aquí abajo, no va a salir nada. Estaba disponible en la
		sesión y no lo contaba nadie.
	-->
	{#if user}
		{#if user.emailVerified}
			<p class="destination">
				Llegan a <span class="address">{user.email}</span>.
			</p>
		{:else}
			<p class="destination unverified">
				<span class="address">{user.email}</span> todavía no está verificada, así que no podemos
				enviarte nada: marca lo que quieras recibir y confirma la dirección para que empiece a
				llegar.
				<a href={resolve('/auth/verify-email')}>Pide un enlace nuevo</a>.
			</p>
		{/if}
	{/if}

	<form
		method="POST"
		action="?/updatePreferences"
		use:enhance={() => {
			prefsLoading = true;
			return async ({ update }) => {
				await update();
				prefsLoading = false;
			};
		}}
	>
		<!--
			La casilla delante y no al fondo de la fila: son opciones que se marcan
			antes de guardar, no interruptores que actúan solos, y puestas en
			columna se ve de un vistazo qué está encendido. Al final de la fila
			quedaban a un palmo de la frase que las nombra, y en el móvil se
			alineaban con el renglón de en medio del párrafo.

			La fila entera es la etiqueta. El nombre y la explicación van nombrando
			y describiendo la casilla por separado, para que quien la oiga leída no
			se trague la frase entera como nombre del control.
		-->
		<div class="messages">
			<label class="message">
				<span class="control">
					<Checkbox
						name="emailAlerts"
						bind:checked={emailAlerts}
						aria-labelledby="alerts-name"
						aria-describedby="alerts-hint"
					/>
				</span>
				<span class="text">
					<span class="name" id="alerts-name">Alertas de actividad</span>
					<span class="hint" id="alerts-hint">
						Un correo cuando pase algo que deberías saber en tu cuenta.
					</span>
				</span>
			</label>

			<label class="message">
				<span class="control">
					<Checkbox
						name="weeklySummary"
						bind:checked={weeklySummary}
						aria-labelledby="weekly-name"
						aria-describedby="weekly-hint"
					/>
				</span>
				<span class="text">
					<span class="name" id="weekly-name">Resumen semanal</span>
					<span class="hint" id="weekly-hint">
						Cada semana, cómo se movieron tus portafolios.
					</span>
				</span>
			</label>
		</div>

		{#if prefsError}
			<p class="feedback error">{prefsError}</p>
		{/if}
		{#if prefsSuccess}
			<p class="feedback success">Preferencias guardadas correctamente.</p>
		{/if}

		<div class="form-actions">
			<Button type="submit" size="sm" loading={prefsLoading}>
				{prefsLoading ? 'Guardando…' : 'Guardar preferencias'}
			</Button>
		</div>
	</form>
</NotificationSection>

<style>
	.destination {
		max-width: 62ch;
		margin: 0 0 0.85rem;
		font-size: 0.87rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	/* La dirección en la tipografía de máquina, como las IP y las horas de las
	   sesiones abiertas: es una cadena que hay que leer carácter a carácter. */
	.address {
		font-family: var(--font-mono);
		font-size: 0.85em;
		color: var(--text);
	}

	.destination.unverified {
		padding-left: 0.75rem;
		border-left: 2px solid var(--amber);
		color: var(--text);
	}

	.destination a {
		color: var(--amber);
	}

	.message {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.9rem;
		padding: 0.95rem 0;
		border-bottom: 1px solid var(--border);
		cursor: pointer;
	}

	/* Empujada al renglón del nombre: la casilla mide veinte píxeles y el nombre
	   dieciocho, así que alineadas por arriba quedan un pelo descuadradas. */
	.control {
		padding-top: 0.1rem;
	}

	.message:last-child {
		border-bottom: none;
	}

	.text {
		min-width: 0;
	}

	.name {
		display: block;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.hint {
		display: block;
		max-width: 46ch;
		margin-top: 0.2rem;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-muted);
	}
</style>
