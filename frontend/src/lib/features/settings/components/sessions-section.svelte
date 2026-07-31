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

<SettingsSection title="Sesiones activas">
	<p class="hint sessions-intro">
		Estos son los dispositivos con acceso a tu cuenta y a la información de tu patrimonio. Cierra
		cualquier sesión que no reconozcas.
	</p>

	{#if sessionList.length === 0}
		<p class="hint">No se pudieron cargar las sesiones activas.</p>
	{:else}
		<ul class="session-list">
			{#each sessionList as session (session.id)}
				<li class="session-item">
					<div class="session-info">
						<div class="session-device">
							<span class="session-name">{describeDevice(session.userAgent)}</span>
							{#if session.current}
								<span class="session-badge">Este dispositivo</span>
							{/if}
						</div>
						<p class="session-meta">
							{session.ipAddress ?? 'IP desconocida'}{session.location
								? ` · ${session.location}`
								: ''} · Última actividad: {formatSessionDate(session.lastActiveAt)}
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
							<button type="submit" class="btn-revoke" disabled={revokingSessionId === session.id}>
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
					<Button type="submit" variant="secondary" loading={revokeOthersLoading}>
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
	.sessions-intro {
		margin-bottom: 1.25rem;
	}

	.session-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.session-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.875rem 1rem;
		border: 1px solid rgba(212, 145, 42, 0.12);
		border-radius: 8px;
		background: var(--surface-2, rgba(255, 255, 255, 0.02));
	}

	.session-info {
		min-width: 0;
	}

	.session-device {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.session-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text);
	}

	.session-badge {
		font-size: 0.675rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--amber);
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 20px;
		padding: 0.15rem 0.55rem;
	}

	.session-meta {
		margin: 0.3rem 0 0;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.45);
		font-family: var(--font-mono);
		overflow-wrap: anywhere;
	}

	.btn-revoke {
		flex-shrink: 0;
		padding: 0.4rem 0.875rem;
		border-radius: 6px;
		border: 1px solid rgba(224, 90, 90, 0.35);
		background: rgba(224, 90, 90, 0.06);
		color: var(--red, #e05a5a);
		font-size: 0.775rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.btn-revoke:hover:not(:disabled) {
		background: rgba(224, 90, 90, 0.14);
		border-color: rgba(224, 90, 90, 0.6);
	}

	.btn-revoke:disabled {
		opacity: 0.6;
		cursor: default;
	}
</style>
