<script lang="ts">
	/**
	 * Tokens de acceso para clientes MCP (Claude Desktop, Claude Code…).
	 *
	 * Un cliente MCP se configura una vez y funciona solo: no puede usar la
	 * sesión del navegador, que dura minutos y se renueva con una cookie que él
	 * no tiene. Estos tokens son esa credencial, y solo valen para `/mcp`.
	 *
	 * El secreto se muestra una única vez, al crearlo o al rotarlo. El backend
	 * guarda su hash, así que no hay forma de volver a verlo desde aquí: quien lo
	 * pierda, rota el token y reconfigura el cliente.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import Button from '$lib/ui/button.svelte';
	import Input from '$lib/ui/input.svelte';
	import SettingsSection from './settings-section.svelte';
	import {
		actionError,
		formatMCPTokenDate,
		issuedMCPToken,
		MCP_TOKEN_EXPIRY_OPTIONS,
		type MCPToken,
		type SettingsForm
	} from '../settings';

	interface Props {
		tokens: MCPToken[] | undefined;
		/** URL del endpoint MCP, para el ejemplo de configuración. */
		mcpUrl: string;
		form: SettingsForm;
	}

	let { tokens, mcpUrl, form }: Props = $props();

	let name = $state('');
	let expiresInDays = $state(90);
	let creating = $state(false);
	let rotatingId = $state<string | null>(null);
	let deletingId = $state<string | null>(null);
	let copied = $state(false);

	const tokenList = $derived(tokens ?? []);
	const issued = $derived(issuedMCPToken(form));
	const error = $derived(
		actionError(form, 'createMcpToken') ||
			actionError(form, 'rotateMcpToken') ||
			actionError(form, 'deleteMcpToken')
	);

	/** Ejemplo listo para pegar, con el token recién emitido dentro. */
	const configSnippet = $derived(
		JSON.stringify(
			{
				mcpServers: {
					finexia: {
						type: 'http',
						url: mcpUrl,
						headers: { Authorization: `Bearer ${issued?.token ?? 'TU_TOKEN'}` }
					}
				}
			},
			null,
			2
		)
	);

	async function copy(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			// Sin permiso de portapapeles el token sigue visible y seleccionable;
			// no hay nada que recuperar, así que no se avisa de nada.
			copied = false;
		}
	}
</script>

<SettingsSection title="Acceso para asistentes (MCP)">
	<p class="hint">
		Conecta Claude —u otro cliente MCP— a tus carteras para poder preguntar por tus posiciones, tu
		distribución o tus últimos movimientos. El acceso es de <strong>solo lectura</strong>: un
		asistente puede consultar tus datos, nunca modificarlos.
	</p>
	<p class="hint">
		Crea un token, pégalo en la configuración de tu cliente y apúntalo a <code>{mcpUrl}</code>.
	</p>

	{#if issued}
		<div class="issued">
			<p class="issued-title">
				{form?.action === 'rotateMcpToken' ? 'Token rotado' : 'Token creado'}: cópialo ahora
			</p>
			<p class="issued-note">
				Es la única vez que se muestra. Si lo pierdes tendrás que rotarlo y volver a configurar el
				cliente.
				{#if form?.action === 'rotateMcpToken'}
					El token anterior ya dejó de funcionar.
				{/if}
			</p>

			<div class="secret-row">
				<code class="secret">{issued.token}</code>
				<Button type="button" size="sm" variant="secondary" onclick={() => copy(issued.token)}>
					{copied ? 'Copiado' : 'Copiar'}
				</Button>
			</div>

			<details class="snippet">
				<summary>Ver configuración de ejemplo</summary>
				<pre>{configSnippet}</pre>
			</details>
		</div>
	{/if}

	<form
		method="POST"
		action="?/createMcpToken"
		use:enhance={() => {
			creating = true;
			return async ({ update }) => {
				creating = false;
				// reset:false conserva el resto de la página de ajustes; el nombre
				// se limpia a mano para que el siguiente token no herede el suyo.
				await update({ reset: false });
				name = '';
			};
		}}
	>
		<div class="create-row">
			<Input
				label="Nombre del token"
				name="name"
				placeholder="Claude Desktop"
				autocomplete="off"
				bind:value={name}
				required
			/>
			<div class="field">
				<label class="field-label" for="mcpExpiresInDays">Caducidad</label>
				<select
					id="mcpExpiresInDays"
					name="expiresInDays"
					class="field-select"
					bind:value={expiresInDays}
				>
					{#each MCP_TOKEN_EXPIRY_OPTIONS as option (option.days)}
						<option value={option.days}>{option.label}</option>
					{/each}
				</select>
			</div>
			<Button type="submit" loading={creating}>Crear token</Button>
		</div>
	</form>

	{#if error}
		<p class="feedback error">{error}</p>
	{/if}

	{#if tokenList.length > 0}
		<ul class="token-list">
			{#each tokenList as token (token.id)}
				<li class="token" class:is-expired={token.expired}>
					<div class="token-info">
						<div class="token-head">
							<span class="token-name">{token.name}</span>
							<code class="token-last4">····{token.last4}</code>
							{#if token.expired}
								<Badge tone="danger">Caducado</Badge>
							{:else if !token.lastUsedAt}
								<Badge tone="neutral">Sin usar</Badge>
							{:else}
								<Badge tone="success">Activo</Badge>
							{/if}
						</div>
						<p class="token-meta">
							Último uso: {formatMCPTokenDate(token.lastUsedAt)} · Caduca: {token.expiresAt
								? formatMCPTokenDate(token.expiresAt)
								: 'nunca'} · Creado: {formatMCPTokenDate(token.createdAt)}
						</p>
					</div>

					<div class="token-actions">
						<form
							method="POST"
							action="?/rotateMcpToken"
							use:enhance={() => {
								rotatingId = token.id;
								return async ({ update }) => {
									rotatingId = null;
									await update({ reset: false });
								};
							}}
						>
							<input type="hidden" name="tokenId" value={token.id} />
							<input type="hidden" name="expiresInDays" value={expiresInDays} />
							<Button type="submit" variant="secondary" size="sm" loading={rotatingId === token.id}>
								Rotar
							</Button>
						</form>
						<form
							method="POST"
							action="?/deleteMcpToken"
							use:enhance={() => {
								deletingId = token.id;
								return async ({ update }) => {
									deletingId = null;
									await update({ reset: false });
								};
							}}
						>
							<input type="hidden" name="tokenId" value={token.id} />
							<Button type="submit" variant="ghost" size="sm" loading={deletingId === token.id}>
								Eliminar
							</Button>
						</form>
					</div>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="hint empty">Todavía no tienes ningún token. Crea uno para conectar tu asistente.</p>
	{/if}
</SettingsSection>

<style>
	.hint code {
		padding: 0.1rem 0.35rem;
		border-radius: 0.25rem;
		background: rgb(255 255 255 / 6%);
		font-size: 0.78rem;
	}

	.issued {
		margin: 1rem 0 1.25rem;
		padding: 1rem;
		border: 1px solid var(--color-accent, #d4af37);
		border-radius: 0.75rem;
		background: rgb(212 175 55 / 6%);
	}

	.issued-title {
		margin: 0 0 0.25rem;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--color-text-primary, #f5f5f5);
	}

	.issued-note {
		margin: 0 0 0.75rem;
		font-size: 0.8rem;
		color: var(--color-text-muted, #9a9a9a);
	}

	.secret-row {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	.secret {
		flex: 1;
		overflow-x: auto;
		padding: 0.5rem 0.6rem;
		border-radius: 0.4rem;
		background: rgb(0 0 0 / 35%);
		font-size: 0.78rem;
		white-space: nowrap;
	}

	.snippet {
		margin-top: 0.75rem;
		font-size: 0.8rem;
		color: var(--color-text-muted, #9a9a9a);
	}

	.snippet summary {
		cursor: pointer;
	}

	.snippet pre {
		overflow-x: auto;
		margin: 0.5rem 0 0;
		padding: 0.75rem;
		border-radius: 0.5rem;
		background: rgb(0 0 0 / 35%);
		font-size: 0.75rem;
		line-height: 1.5;
	}

	.create-row {
		display: flex;
		gap: 0.75rem;
		align-items: flex-end;
		margin: 1rem 0 0.5rem;
	}

	.create-row :global(.input-wrapper) {
		flex: 1;
	}

	.field-label {
		display: block;
		margin-bottom: 0.35rem;
		font-size: 0.8rem;
		color: var(--color-text-secondary, #c4c4c4);
	}

	.field-select {
		padding: 0.55rem 0.7rem;
		border: 1px solid var(--color-border, #2a2a2a);
		border-radius: 0.5rem;
		background: var(--color-surface, #141414);
		color: var(--color-text-primary, #f5f5f5);
		font-size: 0.85rem;
	}

	.token-list {
		display: grid;
		gap: 0.75rem;
		margin: 1rem 0 0;
		padding: 0;
		list-style: none;
	}

	.token {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		align-items: center;
		justify-content: space-between;
		padding: 0.85rem 1rem;
		border: 1px solid var(--color-border, #2a2a2a);
		border-radius: 0.75rem;
		background: var(--color-surface-subtle, rgb(255 255 255 / 2%));
	}

	.token.is-expired {
		opacity: 0.7;
	}

	.token-head {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
	}

	.token-name {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--color-text-primary, #f5f5f5);
	}

	.token-last4 {
		padding: 0.1rem 0.35rem;
		border-radius: 0.25rem;
		background: rgb(255 255 255 / 6%);
		font-size: 0.75rem;
		color: var(--color-text-secondary, #c4c4c4);
	}

	.token-meta {
		margin: 0.35rem 0 0;
		font-size: 0.78rem;
		color: var(--color-text-muted, #9a9a9a);
	}

	.token-actions {
		display: flex;
		gap: 0.5rem;
	}

	.empty {
		margin-top: 1rem;
	}

	.feedback.error {
		margin: 0.5rem 0 0;
		font-size: 0.82rem;
		color: var(--color-danger, #e05260);
	}

	@media (max-width: 640px) {
		.create-row {
			flex-direction: column;
			align-items: stretch;
		}

		.token {
			align-items: flex-start;
		}
	}
</style>
