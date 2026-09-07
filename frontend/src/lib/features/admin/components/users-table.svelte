<script lang="ts">
	/**
	 * Usuarios registrados: en qué estado está cada cuenta y qué se puede hacer
	 * con ella.
	 *
	 * A los administradores no se les ofrece ninguna acción: la fila se queda sin
	 * botones, para no poder banearse entre sí desde aquí. La paginación es del
	 * servidor (un GET con `?page=`), así que va en un formulario y no en el
	 * `Pagination` cliente de `lib/ui`.
	 *
	 * El estado de una cuenta era dos columnas —una insignia «Activo/Baneado» y
	 * un «Verificado: Sí/No»— que nunca decían dos cosas a la vez. Aquí es una
	 * sola, y solo se enciende cuando hay algo que mirar: lo normal es una cuenta
	 * activa y verificada, y noventa insignias verdes no informan de nada.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminBlock from './admin-block.svelte';
	import { INVITE_ROLES, formatDay, type PageMeta, type UserItem } from '../admin';
	import { describeUsers } from '../desk';

	interface Props {
		users: UserItem[];
		meta: PageMeta;
		/** `form` de la página, para el error por fila. */
		form: Record<string, unknown> | null;
	}

	let { users, meta, form }: Props = $props();

	let deleting = $state<string | null>(null);
	let banning = $state<string | null>(null);

	const total = $derived(Number(meta.totalUsers ?? users.length));
	const roleLabel = (role: string) => INVITE_ROLES.find((r) => r.value === role)?.label ?? role;
</script>

<AdminBlock title="Usuarios registrados" summary={describeUsers(users, total, meta.totalPages > 1)}>
	<DataTable caption="Cuentas registradas, su estado y las acciones sobre cada una">
		<thead>
			<tr>
				<th>Nombre</th>
				<th>Correo</th>
				<th>Rol</th>
				<th>Estado</th>
				<th>Alta</th>
				<th><span class="sr-only">Acciones</span></th>
			</tr>
		</thead>
		<tbody>
			{#each users as user (user.id)}
				{@const isAdmin = user.role?.name === 'admin'}
				{@const isBanned = !!user.bannedAt}
				<tr class:row-muted={isBanned}>
					<td class="cell-name">{user.name}</td>
					<td class="cell-email">{user.email}</td>
					<td class:role-admin={isAdmin}>{roleLabel(user.role?.name ?? '')}</td>
					<td>
						{#if isBanned}
							<Badge tone="danger">Baneada</Badge>
						{:else if !user.emailVerified}
							<Badge tone="warning">Sin verificar</Badge>
						{:else}
							<span class="plain-state">Activa</span>
						{/if}
					</td>
					<td>{formatDay(user.createdAt)}</td>
					<td class="cell-actions">
						{#if !isAdmin}
							<div class="row-actions">
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
									<button
										class="row-action"
										class:danger={!isBanned}
										type="submit"
										disabled={banning === user.id}
									>
										{isBanned ? 'Levantar el baneo' : 'Banear'}
									</button>
								</form>
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
									<button class="row-action danger" type="submit" disabled={deleting === user.id}>
										Eliminar
									</button>
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

	{#snippet footer()}
		{#if meta.totalPages > 1}
			<form class="pager" method="GET">
				<p class="pager-info">Página {meta.currentPage} de {meta.totalPages}</p>
				<div class="pager-controls">
					<button
						type="submit"
						name="page"
						value={meta.currentPage - 1}
						class="pager-btn"
						disabled={!meta.previous}
					>
						Anterior
					</button>
					<button
						type="submit"
						name="page"
						value={meta.currentPage + 1}
						class="pager-btn"
						disabled={!meta.next}
					>
						Siguiente
					</button>
				</div>
			</form>
		{/if}
	{/snippet}
</AdminBlock>

<style>
	/* El rol solo se destaca cuando cambia lo que se puede hacer con la fila:
	   una cuenta de administrador no se puede banear ni borrar desde aquí. */
	.role-admin {
		color: var(--text) !important;
		font-weight: 500;
	}

	.plain-state {
		font-size: 0.82rem;
		color: var(--text-dim);
	}

	.pager {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
	}

	.pager-info {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.pager-controls {
		display: flex;
		gap: 1.25rem;
	}

	.pager-btn {
		padding: 0;
		border: none;
		background: none;
		font-family: var(--font-body);
		font-size: 0.82rem;
		color: var(--text-muted);
		cursor: pointer;
	}

	.pager-btn:hover:not(:disabled) {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.pager-btn:disabled {
		color: var(--text-dim);
		opacity: 0.5;
		cursor: default;
	}
</style>
