<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Badge from '$components/ui/badge.svelte';
	import Button from '$components/ui/button.svelte';
	import Input from '$components/ui/input.svelte';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	let showInviteForm = $state(false);
	let inviteName = $state('');
	let inviteEmail = $state('');
	let inviteRole = $state('customer');
	let inviting = $state(false);
	let deleting = $state<string | null>(null);
	let banning = $state<string | null>(null);
	let invitingId = $state<string | null>(null);

	function formatDate(iso: string): string {
		return new Intl.DateTimeFormat('es', { dateStyle: 'medium' }).format(new Date(iso));
	}

	const statusLabels: Record<string, string> = {
		pending: 'Pendiente',
		expired: 'Expirada',
		accepted: 'Aceptada',
		revoked: 'Revocada'
	};

	function statusTone(status: string): 'amber' | 'neutral' | 'success' | 'danger' {
		if (status === 'accepted') return 'success';
		if (status === 'revoked') return 'danger';
		if (status === 'expired') return 'neutral';
		return 'amber';
	}

	$effect(() => {
		if (form && 'success' in form && form.success) {
			showInviteForm = false;
			inviteName = '';
			inviteEmail = '';
			inviteRole = 'customer';
		}
	});
</script>

<svelte:head>
	<title>Usuarios — Admin — FINEXIA</title>
</svelte:head>

<PageHeader
	eyebrow="Administración"
	title="Usuarios"
	subtitle="Invita, gestiona y controla el acceso a la plataforma."
>
	{#snippet actions()}
		<Button
			variant="secondary"
			size="sm"
			type="button"
			onclick={() => (showInviteForm = !showInviteForm)}
		>
			{showInviteForm ? 'Cancelar' : '+ Invitar usuario'}
		</Button>
	{/snippet}
</PageHeader>

{#if showInviteForm}
	<div class="create-form-card">
		<Card padding="md">
			<h2 class="form-title">Invitar a un nuevo usuario</h2>
			<p class="form-hint">
				Enviaremos un enlace seguro de un solo uso para que la persona cree su propia contraseña.
			</p>
			<form
				method="POST"
				action="?/inviteUser"
				use:enhance={() => {
					inviting = true;
					return async ({ update }) => {
						inviting = false;
						await update();
					};
				}}
			>
				<div class="form-row">
					<Input label="Nombre (opcional)" name="name" bind:value={inviteName} />
					<Input
						label="Correo electrónico"
						name="email"
						type="email"
						bind:value={inviteEmail}
						required
					/>
					<div class="field">
						<span class="field-label">Rol</span>
						<select class="select" name="role" bind:value={inviteRole}>
							<option value="customer">Usuario</option>
							<option value="admin">Administrador</option>
						</select>
					</div>
				</div>
				{#if form && 'error' in form && form.error}
					<p class="form-error">{form.error}</p>
				{/if}
				{#if form && 'invited' in form && form.invited}
					<p class="form-success">Invitación enviada a {form.invited}.</p>
				{/if}
				<div class="form-actions">
					<Button type="submit" loading={inviting}>Enviar invitación</Button>
				</div>
			</form>
		</Card>
	</div>
{/if}

{#if data.invitations.length > 0}
	<section class="section">
		<h2 class="section-title">Invitaciones pendientes</h2>
		<Card padding="none">
			<div class="table-wrapper">
				<table class="users-table">
					<thead>
						<tr>
							<th>Correo</th>
							<th>Rol</th>
							<th>Estado</th>
							<th>Expira</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{#each data.invitations as inv (inv.id)}
							<tr>
								<td class="cell-email">{inv.email}</td>
								<td>
									<Badge tone={inv.role === 'admin' ? 'amber' : 'neutral'}>{inv.role}</Badge>
								</td>
								<td
									><Badge tone={statusTone(inv.status)}
										>{statusLabels[inv.status] ?? inv.status}</Badge
									></td
								>
								<td class="cell-date">{formatDate(inv.expiresAt)}</td>
								<td class="cell-actions">
									<div class="action-row">
										<form
											method="POST"
											action="?/resendInvitation"
											use:enhance={() => {
												invitingId = inv.id;
												return async ({ update }) => {
													invitingId = null;
													await update({ reset: false });
												};
											}}
										>
											<input type="hidden" name="id" value={inv.id} />
											<Button
												variant="secondary"
												size="sm"
												type="submit"
												loading={invitingId === inv.id}
											>
												Reenviar
											</Button>
										</form>
										<form
											method="POST"
											action="?/revokeInvitation"
											use:enhance={() => {
												return async ({ update }) => {
													await update({ reset: false });
												};
											}}
										>
											<input type="hidden" name="id" value={inv.id} />
											<Button variant="ghost" size="sm" type="submit">
												<span class="delete-label">Revocar</span>
											</Button>
										</form>
									</div>
									{#if form && 'inviteError' in form && form.inviteError && form.inviteId === inv.id}
										<p class="row-error">{form.inviteError}</p>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	</section>
{/if}

{#if data.waitlist.length > 0}
	<section class="section">
		<h2 class="section-title">Lista de espera</h2>
		<Card padding="none">
			<div class="table-wrapper">
				<table class="users-table">
					<thead>
						<tr>
							<th>Correo</th>
							<th>En lista desde</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{#each data.waitlist as entry (entry.id)}
							<tr>
								<td class="cell-email">{entry.email}</td>
								<td class="cell-date">{formatDate(entry.createdAt)}</td>
								<td class="cell-actions">
									<form
										method="POST"
										action="?/inviteUser"
										use:enhance={() => {
											invitingId = entry.id;
											return async ({ update }) => {
												invitingId = null;
												await update({ reset: false });
											};
										}}
									>
										<input type="hidden" name="email" value={entry.email} />
										<input type="hidden" name="role" value="customer" />
										<Button
											variant="secondary"
											size="sm"
											type="submit"
											loading={invitingId === entry.id}
										>
											Invitar
										</Button>
									</form>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	</section>
{/if}

<section class="section">
	<h2 class="section-title">Usuarios registrados</h2>
	<Card padding="none">
		{#if data.users.length === 0}
			<p class="empty-state">No hay usuarios registrados.</p>
		{:else}
			<div class="table-wrapper">
				<table class="users-table">
					<thead>
						<tr>
							<th>Nombre</th>
							<th>Correo</th>
							<th>Rol</th>
							<th>Estado</th>
							<th>Verificado</th>
							<th>Miembro desde</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{#each data.users as user (user.id)}
							{@const isAdmin = user.role?.name === 'admin'}
							{@const isBanned = !!user.bannedAt}
							<tr class:row-banned={isBanned} class:row-admin={isAdmin}>
								<td class="cell-name">{user.name}</td>
								<td class="cell-email">{user.email}</td>
								<td>
									<Badge tone={isAdmin ? 'amber' : 'neutral'}>
										{user.role?.name ?? '—'}
									</Badge>
								</td>
								<td>
									{#if isBanned}
										<Badge tone="danger">Baneado</Badge>
									{:else}
										<Badge tone="success">Activo</Badge>
									{/if}
								</td>
								<td>
									<span class="verified-dot" class:verified={user.emailVerified}>
										{user.emailVerified ? 'Sí' : 'No'}
									</span>
								</td>
								<td class="cell-date">{formatDate(user.createdAt)}</td>
								<td class="cell-actions">
									{#if !isAdmin}
										<div class="action-row">
											<!-- Ban / Unban -->
											<form
												method="POST"
												action="?/banUser"
												use:enhance={() => {
													banning = user.id;
													return async ({ update }) => {
														banning = null;
														await update({ reset: false });
													};
												}}
											>
												<input type="hidden" name="id" value={user.id} />
												<input type="hidden" name="ban" value={isBanned ? 'false' : 'true'} />
												<Button
													variant={isBanned ? 'secondary' : 'ghost'}
													size="sm"
													type="submit"
													loading={banning === user.id}
												>
													{isBanned ? 'Desbanear' : 'Banear'}
												</Button>
											</form>

											<!-- Delete -->
											<form
												method="POST"
												action="?/deleteUser"
												use:enhance={() => {
													deleting = user.id;
													return async ({ update }) => {
														deleting = null;
														await update();
													};
												}}
											>
												<input type="hidden" name="id" value={user.id} />
												<Button
													variant="ghost"
													size="sm"
													type="submit"
													loading={deleting === user.id}
												>
													<span class="delete-label">Eliminar</span>
												</Button>
											</form>
										</div>
										{#if form && 'banError' in form && form.banError && form.banId === user.id}
											<p class="row-error">{form.banError}</p>
										{/if}
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			{#if data.meta.totalPages > 1}
				<form class="pagination" method="GET">
					{#if data.meta.previous}
						<button type="submit" name="page" value={data.meta.currentPage - 1} class="page-btn"
							>← Anterior</button
						>
					{/if}
					<span class="page-info">Página {data.meta.currentPage} de {data.meta.totalPages}</span>
					{#if data.meta.next}
						<button type="submit" name="page" value={data.meta.currentPage + 1} class="page-btn"
							>Siguiente →</button
						>
					{/if}
				</form>
			{/if}
		{/if}
	</Card>
</section>

<style>
	.create-form-card {
		margin-bottom: 1.5rem;
	}

	.form-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
		margin: 0 0 0.35rem 0;
	}

	.form-hint {
		font-size: 0.82rem;
		color: var(--text-dim);
		margin: 0 0 1.25rem 0;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr auto;
		gap: 1rem;
		margin-bottom: 1rem;
		align-items: end;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.field-label {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text-muted);
	}

	.select {
		appearance: none;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		color: var(--text);
		font-size: 0.875rem;
		padding: 0.6rem 0.75rem;
		min-width: 9rem;
		cursor: pointer;
	}

	.select:focus {
		outline: none;
		border-color: var(--amber);
	}

	.form-error {
		font-size: 0.82rem;
		color: var(--red);
		margin: 0 0 0.75rem 0;
	}

	.form-success {
		font-size: 0.82rem;
		color: var(--green);
		margin: 0 0 0.75rem 0;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
	}

	.section {
		margin-bottom: 1.75rem;
	}

	.section-title {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--text-dim);
		margin: 0 0 0.75rem 0;
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.users-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.users-table th {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--text-dim);
		padding: 0.875rem 1.25rem;
		text-align: left;
		border-bottom: 1px solid var(--border);
		white-space: nowrap;
	}

	.users-table td {
		padding: 0.75rem 1.25rem;
		color: var(--text-muted);
		border-bottom: 1px solid var(--border);
		vertical-align: middle;
	}

	.users-table tbody tr:last-child td {
		border-bottom: none;
	}

	.users-table tbody tr:hover td {
		background: var(--surface-2);
	}

	.row-banned td {
		background: rgba(239, 68, 68, 0.04) !important;
		opacity: 0.75;
	}

	.row-admin .cell-name {
		color: var(--amber-light) !important;
	}

	.cell-name {
		color: var(--text) !important;
		font-weight: 500;
		white-space: nowrap;
	}

	.cell-email {
		font-family: var(--font-mono);
		font-size: 0.8rem;
	}

	.cell-date {
		white-space: nowrap;
		font-family: var(--font-mono);
		font-size: 0.8rem;
	}

	.cell-actions {
		text-align: right;
		white-space: nowrap;
	}

	.action-row {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 0.25rem;
	}

	.delete-label {
		color: var(--red);
	}

	.row-error {
		font-size: 0.75rem;
		color: var(--red);
		margin: 0.25rem 0 0;
		text-align: right;
	}

	.verified-dot {
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.verified-dot.verified {
		color: var(--green);
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-dim);
		font-size: 0.9rem;
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1.5rem;
		padding: 1rem 1.25rem;
		border-top: 1px solid var(--border);
	}

	.page-btn {
		font-size: 0.85rem;
		color: var(--amber);
		text-decoration: none;
		font-weight: 500;
		transition: color 0.2s ease;
	}

	.page-btn:hover {
		color: var(--amber-light);
	}

	.page-info {
		font-family: var(--font-mono);
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
