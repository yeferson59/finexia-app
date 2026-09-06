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
		return scopes.map((scope) => SCOPE_LABELS[scope] ?? scope).join(', ');
	}
</script>

<SettingsSection
	title="Aplicaciones conectadas"
	description="Aplicaciones a las que autorizaste el acceso a tus carteras. Al desconectar una, su acceso deja de funcionar al instante."
>
	{#if grantList.length === 0}
		<p class="hint empty">
			No hay ninguna. Las que conectes desde otra aplicación —el conector de claude.ai, por ejemplo—
			aparecerán aquí para que puedas cortarles el acceso.
		</p>
	{:else}
		<ul class="grant-list">
			{#each grantList as grant (grant.id)}
				<li class="grant-item">
					<div class="grant-info">
						<p class="grant-name">{grant.clientName}</p>
						<p class="grant-meta">
							{describeScopes(grant.scopes)}. Conectada el {formatMCPTokenDate(grant.createdAt)},
							con último uso {formatMCPTokenDate(grant.lastUsedAt)}.
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
						<button type="submit" class="row-action danger" disabled={revokingId === grant.id}>
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
	.empty {
		max-width: 58ch;
	}

	/* Filas con filete, como las sesiones: son la misma clase de dato —algo que
	   tiene acceso a la cuenta— y se leen igual. */
	.grant-list {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.grant-item {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.8rem 0;
		border-bottom: 1px solid var(--border);
	}

	.grant-item:first-child {
		padding-top: 0;
	}

	.grant-item:last-child {
		border-bottom: none;
	}

	.grant-info {
		min-width: 0;
	}

	.grant-name {
		margin: 0;
		font-size: 0.88rem;
		color: var(--text);
	}

	.grant-meta {
		margin: 0.25rem 0 0;
		font-size: 0.78rem;
		line-height: 1.5;
		color: var(--text-dim);
	}
</style>
