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

<SettingsSection
	title="Asistentes"
	description="Conecta Claude —u otro cliente MCP— a tus carteras para preguntarle por tus posiciones, tu distribución o tus últimos movimientos."
>
	{#snippet aside()}
		<p class="readonly">
			El acceso es de solo lectura: un asistente consulta tus datos, nunca los modifica.
		</p>
	{/snippet}

	<p class="hint">
		Crea un token, pégalo en la configuración de tu cliente y apúntalo a esta dirección:
	</p>
	<p class="endpoint"><code>{mcpUrl}</code></p>

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
			<Button type="submit" size="sm" loading={creating}>Crear token</Button>
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
						<p class="token-head">
							<span class="token-name">{token.name}</span>
							<code class="token-last4">····{token.last4}</code>
						</p>
						<!-- El estado en la misma frase que las fechas: era una insignia
						     («CADUCADO», «SIN USAR») encima de un pie con tres datos
						     encadenados por puntos medios. -->
						<p class="token-meta" class:is-expired-note={token.expired}>
							{#if token.expired}
								Caducó el {formatMCPTokenDate(token.expiresAt)} y ya no sirve.
							{:else if !token.lastUsedAt}
								Sin usar todavía.
							{:else}
								Último uso {formatMCPTokenDate(token.lastUsedAt)}.
							{/if}
							{#if !token.expired}
								{token.expiresAt
									? `Caduca el ${formatMCPTokenDate(token.expiresAt)}.`
									: 'No caduca.'}
							{/if}
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
							<button type="submit" class="row-action" disabled={rotatingId === token.id}>
								{rotatingId === token.id ? 'Rotando…' : 'Rotar'}
							</button>
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
							<button type="submit" class="row-action danger" disabled={deletingId === token.id}>
								{deletingId === token.id ? 'Eliminando…' : 'Eliminar'}
							</button>
						</form>
					</div>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="hint empty">Todavía no hay ninguno. Crea el primero para conectar tu asistente.</p>
	{/if}
</SettingsSection>

<style>
	/* El único bloque enmarcado de la sección, como los códigos de recuperación:
	   un secreto que solo se enseña una vez y hay que copiar antes de irse. */
	.issued {
		margin: 1.25rem 0;
		padding: 1rem 1.1rem;
		border: 1px solid rgba(212, 145, 42, 0.35);
		border-radius: 10px;
		background: rgba(212, 145, 42, 0.06);
	}

	.issued-title {
		margin: 0 0 0.3rem;
		font-size: 0.88rem;
		font-weight: 500;
		color: var(--amber-light);
	}

	.issued-note {
		max-width: 58ch;
		margin: 0 0 0.85rem;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
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

	/* El mismo aspecto que el campo de al lado, que es de `ui/input`: eran dos
	   controles de la misma fila con dos alturas, dos bordes y dos grises. */
	.field-label {
		display: block;
		margin-bottom: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		letter-spacing: 0.3px;
		color: var(--text);
	}

	.field-select {
		padding: 0.875rem 1rem;
		border-radius: 8px;
		border: 1px solid rgba(212, 145, 42, 0.2);
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.95rem;
		box-sizing: border-box;
		cursor: pointer;
	}

	.field-select option {
		background: var(--bg);
		color: var(--text);
	}

	/* Filas con filete, como las sesiones y las aplicaciones: los tres son lo
	   mismo —algo que tiene acceso a la cuenta— y se leen igual. */
	.token-list {
		margin: 1.25rem 0 0;
		padding: 0;
		list-style: none;
	}

	.token {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		align-items: baseline;
		justify-content: space-between;
		padding: 0.8rem 0;
		border-bottom: 1px solid var(--border);
	}

	.token:last-child {
		border-bottom: none;
	}

	.token-head {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: baseline;
		margin: 0;
	}

	.token-name {
		font-size: 0.88rem;
		color: var(--text);
	}

	.token-last4 {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.token-meta {
		margin: 0.25rem 0 0;
		font-size: 0.78rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	/* Un token caducado no está apagado: hay que verlo para borrarlo. Lo que
	   cambia es lo que dice de sí mismo. */
	.is-expired-note {
		color: var(--red);
	}

	.token-actions {
		display: flex;
		gap: 0.5rem;
	}

	.empty {
		max-width: 58ch;
	}

	/* La dirección del servidor: se copia a mano en la configuración del
	   cliente, así que va suelta y en mono, no metida en mitad de una frase. */
	.endpoint {
		margin: 0.5rem 0 0;
		overflow-x: auto;
	}

	.endpoint code {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		color: var(--amber-light);
		white-space: nowrap;
	}

	.readonly {
		max-width: 40ch;
		margin: 0.7rem 0 0;
		font-size: 0.78rem;
		line-height: 1.55;
		color: var(--text-dim);
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
