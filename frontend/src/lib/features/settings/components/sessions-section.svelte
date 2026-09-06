<script lang="ts">
	/** Sesiones activas del usuario, con cierre individual o de todas las demás. */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import SettingsSection from './settings-section.svelte';
	import {
		actionError,
		actionSucceeded,
		countOtherSessions,
		describeDevice,
		formatSessionDate,
		type ActiveSession,
		type SettingsForm
	} from '../settings';

	interface Props {
		sessions: ActiveSession[] | undefined;
		form: SettingsForm;
	}

	let { sessions, form }: Props = $props();

	let revokingSessionId = $state<string | null>(null);
	let revokeOthersLoading = $state(false);

	const sessionList = $derived(sessions ?? []);
	const otherSessionsCount = $derived(countOtherSessions(sessions));

	const sessionsError = $derived(
		actionError(form, 'revokeSession') || actionError(form, 'revokeOtherSessions')
	);
	const sessionsSuccess = $derived(
		actionSucceeded(form, 'revokeSession') || actionSucceeded(form, 'revokeOtherSessions')
	);
</script>

<SettingsSection
	title="Sesiones abiertas"
	description="Los dispositivos que pueden entrar a tu cuenta ahora mismo. Cierra cualquiera que no reconozcas."
>
	{#if sessionList.length === 0}
		<p class="hint">No se pudieron cargar las sesiones activas.</p>
	{:else}
		<ul class="session-list">
			{#each sessionList as session (session.id)}
				<li class="session-item">
					<div class="session-info">
						<p class="session-name">
							{describeDevice(session.userAgent)}{#if session.current}<span class="here">
									, el que estás usando</span
								>{/if}
						</p>
						<!-- La IP y la hora en mono porque son cadenas de máquina; lo que
						     las nombra, en prosa, porque lo escribió alguien. -->
						<p class="session-meta">
							Desde <span class="mono">{session.ipAddress ?? 'una IP desconocida'}</span
							>{#if session.location}, {session.location}{/if}. Última actividad
							<span class="mono">{formatSessionDate(session.lastActiveAt)}</span>.
						</p>
					</div>
					{#if !session.current}
						<form
							method="POST"
							action="?/revokeSession"
							use:enhance={() => {
								revokingSessionId = session.id;
								return async ({ update }) => {
									await update();
									revokingSessionId = null;
								};
							}}
						>
							<input type="hidden" name="sessionId" value={session.id} />
							<button
								type="submit"
								class="row-action danger"
								disabled={revokingSessionId === session.id}
							>
								{revokingSessionId === session.id ? 'Cerrando…' : 'Cerrar sesión'}
							</button>
						</form>
					{/if}
				</li>
			{/each}
		</ul>

		{#if sessionsError}
			<p class="feedback error">{sessionsError}</p>
		{/if}
		{#if sessionsSuccess}
			<p class="feedback success">Sesión cerrada correctamente.</p>
		{/if}

		{#if otherSessionsCount > 0}
			<div class="form-actions">
				<form
					method="POST"
					action="?/revokeOtherSessions"
					use:enhance={() => {
						revokeOthersLoading = true;
						return async ({ update }) => {
							await update();
							revokeOthersLoading = false;
						};
					}}
				>
					<Button type="submit" variant="secondary" size="sm" loading={revokeOthersLoading}>
						{revokeOthersLoading
							? 'Cerrando sesiones…'
							: `Cerrar las demás sesiones (${otherSessionsCount})`}
					</Button>
				</form>
			</div>
		{/if}
	{/if}
</SettingsSection>

<style>
	/* Una lista con filetes, no cajas: son filas de datos, como las del resto
	   del panel. */
	.session-list {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.session-item {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.8rem 0;
		border-bottom: 1px solid var(--border);
	}

	.session-item:first-child {
		padding-top: 0;
	}

	.session-item:last-child {
		border-bottom: none;
	}

	.session-info {
		min-width: 0;
	}

	.session-name {
		margin: 0;
		font-size: 0.88rem;
		color: var(--text);
	}

	/* «el que estás usando» era una píldora ESTE DISPOSITIVO. Dice lo mismo
	   dentro de la frase y sin pedir un renglón para ella sola. */
	.here {
		color: var(--text-muted);
	}

	.session-meta {
		margin: 0.25rem 0 0;
		font-size: 0.78rem;
		line-height: 1.5;
		color: var(--text-dim);
		overflow-wrap: anywhere;
	}

	.mono {
		font-family: var(--font-mono);
		font-size: 0.74rem;
	}
</style>
