<script lang="ts">
	/**
	 * Usuarios registrados: estado, verificación y acciones de baneo/borrado.
	 *
	 * A los administradores no se les ofrece ninguna acción: la fila se pinta
	 * distinta y se queda sin botones, para no poder banearse entre sí desde
	 * aquí. La paginación es del servidor (un GET con `?page=`), así que va en
	 * un formulario y no en el `Pagination` cliente de `lib/ui`.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminSection from './admin-section.svelte';
	import { formatDay, type PageMeta, type UserItem } from '../admin';

	interface Props {
		users: UserItem[];
		meta: PageMeta;
		/** `form` de la página, para el error por fila. */
		form: Record<string, unknown> | null;
	}

	let { users, meta, form }: Props = $props();

	let deleting = $state<string | null>(null);
	let banning = $state<string | null>(null);
</script>

<AdminSection title="Usuarios registrados">
	{#if users.length === 0}
		<p class="empty-state">No hay usuarios registrados.</p>
	{:else}
		<DataTable>
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
				{#each users as user (user.id)}
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
						<td class="cell-date">{formatDay(user.createdAt)}</td>
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
										<Button variant="ghost" size="sm" type="submit" loading={deleting === user.id}>
											<span class="delete-label">Eliminar</span>
										</Button>
									</form>
								</div>
								{#if form?.banError && form?.banId === user.id}
									<p class="row-error">{form.banError}</p>
								{/if}
								{#if form?.deleteError && form?.deleteId === user.id}
									<p class="row-error">{form.deleteError}</p>
								{/if}
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</DataTable>

		{#if meta.totalPages > 1}
			<form class="pagination" method="GET">
				{#if meta.previous}
					<button type="submit" name="page" value={meta.currentPage - 1} class="page-btn"
						>← Anterior</button
					>
				{/if}
				<span class="page-info">Página {meta.currentPage} de {meta.totalPages}</span>
				{#if meta.next}
					<button type="submit" name="page" value={meta.currentPage + 1} class="page-btn"
						>Siguiente →</button
					>
				{/if}
			</form>
		{/if}
	{/if}
</AdminSection>

<style>
	.row-banned td {
		background: rgba(239, 68, 68, 0.04) !important;
		opacity: 0.75;
	}

	.row-admin .cell-name {
		color: var(--amber-light) !important;
	}

	.verified-dot {
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.verified-dot.verified {
		color: var(--green);
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
</style>
