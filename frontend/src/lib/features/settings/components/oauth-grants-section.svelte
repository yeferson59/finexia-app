<script lang="ts">
	/**
	 * Aplicaciones externas conectadas por OAuth (el conector remoto de
	 * claude.ai y cualquier otro cliente MCP que se autorice a sí mismo).
	 *
	 * Es el contrapeso de la pantalla de consentimiento: allí se concede el
	 * acceso, aquí se retira. Y tiene que estar, porque un token OAuth no se
	 * puede revocar de ninguna otra forma desde la interfaz — el cliente lo
	 * renueva solo mientras nadie corte el permiso.
	 *
	 * `clientName` lo eligió quien registró el cliente y el registro es abierto,
	 * así que se pinta como texto y nunca como enlace.
	 */
	import { enhance } from '$app/forms';
	import SettingsSection from './settings-section.svelte';
	import {
		actionError,
		actionSucceeded,
		formatMCPTokenDate,
		type OAuthGrant,
		type SettingsForm
	} from '../settings';

	interface Props {
		grants: OAuthGrant[] | undefined;
		form: SettingsForm;
	}

	let { grants, form }: Props = $props();

	let revokingId = $state<string | null>(null);

	const grantList = $derived(grants ?? []);
	const error = $derived(actionError(form, 'revokeOAuthGrant'));
	const success = $derived(actionSucceeded(form, 'revokeOAuthGrant'));

	/** Los ámbitos, en lo que significan; ver la pantalla de consentimiento. */
	const SCOPE_LABELS: Record<string, string> = {
		'mcp:read': 'Solo lectura'
	};

	function describeScopes(scopes: string[]): string {
		return scopes.map((scope) => SCOPE_LABELS[scope] ?? scope).join(' · ');
	}
</script>

<SettingsSection title="Aplicaciones conectadas">
	<p class="hint">
		Aplicaciones que pediste conectar y a las que autorizaste el acceso a tus carteras. Desconecta
		cualquiera que ya no uses: su acceso deja de funcionar al instante.
	</p>

	{#if grantList.length === 0}
		<p class="hint empty">No tienes ninguna aplicación conectada.</p>
	{:else}
		<ul class="grant-list">
			{#each grantList as grant (grant.id)}
				<li class="grant-item">
					<div class="grant-info">
						<span class="grant-name">{grant.clientName}</span>
						<p class="grant-meta">
							{describeScopes(grant.scopes)} · Conectada el {formatMCPTokenDate(grant.createdAt)} · Último
							uso: {formatMCPTokenDate(grant.lastUsedAt)}
						</p>
					</div>
					<form
						method="POST"
						action="?/revokeOAuthGrant"
						use:enhance={() => {
							revokingId = grant.id;

							return async ({ update }) => {
								await update();
								revokingId = null;
							};
						}}
					>
						<input type="hidden" name="grantId" value={grant.id} />
						<button type="submit" class="btn-revoke" disabled={revokingId === grant.id}>
							{revokingId === grant.id ? 'Desconectando…' : 'Desconectar'}
						</button>
					</form>
				</li>
			{/each}
		</ul>
	{/if}

	{#if error}
		<p class="feedback error">{error}</p>
	{/if}
	{#if success}
		<p class="feedback success">Aplicación desconectada.</p>
	{/if}
</SettingsSection>

<style>
	.hint {
		font-size: 0.8125rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.empty {
		margin-top: 1rem;
	}

	.grant-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-top: 1rem;
	}

	.grant-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.75rem 0.875rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		background: var(--surface);
	}

	.grant-info {
		min-width: 0;
	}

	.grant-name {
		font-size: 0.9375rem;
		font-weight: 500;
	}

	.grant-meta {
		margin-top: 0.2rem;
		font-size: 0.75rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.btn-revoke {
		flex-shrink: 0;
		padding: 0.4rem 0.7rem;
		border: 1px solid var(--border-strong);
		border-radius: 0.375rem;
		background: transparent;
		font-family: var(--font-body);
		font-size: 0.75rem;
		color: var(--text-muted);
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-revoke:hover:not(:disabled) {
		border-color: var(--red);
		color: var(--red);
	}

	.btn-revoke:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.feedback {
		margin-top: 0.75rem;
		font-size: 0.8125rem;
	}

	.error {
		color: var(--red);
	}

	.success {
		color: var(--green);
	}
</style>
