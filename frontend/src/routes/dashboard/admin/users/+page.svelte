<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Badge from '$components/ui/badge.svelte';
	import Button from '$components/ui/button.svelte';
	import Input from '$components/ui/input.svelte';
	import { resolve } from '$app/paths';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	let showCreateForm = $state(false);
	let createName = $state('');
	let createEmail = $state('');
	let creating = $state(false);
	let deleting = $state<string | null>(null);

	function formatDate(iso: string): string {
		return new Intl.DateTimeFormat('es', { dateStyle: 'medium' }).format(new Date(iso));
	}

	$effect(() => {
		if (form?.success) {
			showCreateForm = false;
			createName = '';
			createEmail = '';
		}
	});
</script>

<svelte:head>
	<title>Usuarios — Admin — FINEXIA</title>
</svelte:head>

<PageHeader eyebrow="Administración" title="Usuarios" subtitle="Gestiona los usuarios del sistema.">
	{#snippet actions()}
		<Button
			variant="secondary"
			size="sm"
			type="button"
			onclick={() => (showCreateForm = !showCreateForm)}
		>
			{showCreateForm ? 'Cancelar' : '+ Crear Usuario'}
		</Button>
	{/snippet}
</PageHeader>

{#if showCreateForm}
	<div class="create-form-card">
	<Card padding="md">
		<h2 class="form-title">Nuevo usuario</h2>
		<form
			method="POST"
			action="?/createUser"
			use:enhance={() => {
				creating = true;
				return async ({ update }) => {
					creating = false;
					await update();
				};
			}}
		>
			<div class="form-row">
				<Input label="Nombre" name="name" bind:value={createName} required />
				<Input label="Correo electrónico" name="email" type="email" bind:value={createEmail} required />
			</div>
			{#if form?.error}
				<p class="form-error">{form.error}</p>
			{/if}
			<div class="form-actions">
				<Button type="submit" loading={creating}>Crear usuario</Button>
			</div>
		</form>
	</Card>
	</div>
{/if}

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
						<th>Verificado</th>
						<th>Miembro desde</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each data.users as user (user.id)}
						<tr>
							<td class="cell-name">{user.name}</td>
							<td class="cell-email">{user.email}</td>
							<td>
								<Badge tone={user.role?.name === 'admin' ? 'amber' : 'neutral'}>
									{user.role?.name ?? '—'}
								</Badge>
							</td>
							<td>
								<span class="verified-dot" class:verified={user.emailVerified}>
									{user.emailVerified ? 'Sí' : 'No'}
								</span>
							</td>
							<td class="cell-date">{formatDate(user.createdAt)}</td>
							<td class="cell-action">
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
										Eliminar
									</Button>
								</form>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		{#if data.meta.totalPages > 1}
			<div class="pagination">
				{#if data.meta.previous}
					<a href={`?page=${data.meta.currentPage - 1}`} class="page-btn">← Anterior</a>
				{/if}
				<span class="page-info">Página {data.meta.currentPage} de {data.meta.totalPages}</span>
				{#if data.meta.next}
					<a href={`?page=${data.meta.currentPage + 1}`} class="page-btn">Siguiente →</a>
				{/if}
			</div>
		{/if}
	{/if}
</Card>

<style>
	.create-form-card {
		margin-bottom: 1.5rem;
	}

	.form-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
		margin: 0 0 1.25rem 0;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.form-error {
		font-size: 0.82rem;
		color: var(--red);
		margin: 0 0 0.75rem 0;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
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
		padding: 0.875rem 1.25rem;
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

	.cell-action {
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
