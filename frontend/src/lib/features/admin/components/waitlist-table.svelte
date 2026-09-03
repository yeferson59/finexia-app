<script lang="ts">
	/**
	 * Lista de espera: correos que pidieron acceso y aún no tienen invitación.
	 * Invitar desde aquí reutiliza la misma action que el formulario de alta;
	 * eliminar saca la entrada de la lista (typos, duplicados o bajas), y el
	 * correo queda libre para volver a apuntarse.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminSection from './admin-section.svelte';
	import { formatDay, type WaitlistItem } from '../admin';

	interface Props {
		waitlist: WaitlistItem[];
		/** `form` de la página, para el error por fila. */
		form: Record<string, unknown> | null;
	}

	let { waitlist, form }: Props = $props();

	let invitingId = $state<string | null>(null);
	let deletingId = $state<string | null>(null);
</script>

<AdminSection title="Lista de espera">
	<DataTable>
		<thead>
			<tr>
				<th>Correo</th>
				<th>En lista desde</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each waitlist as entry (entry.id)}
				<tr>
					<td class="cell-email">{entry.email}</td>
					<td class="cell-date">{formatDay(entry.createdAt)}</td>
					<td class="cell-actions">
						<div class="action-row">
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
							<form
								method="POST"
								action="?/deleteWaitlist"
								use:enhance={() => {
									deletingId = entry.id;
									return async ({ update }) => {
										deletingId = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="id" value={entry.id} />
								<Button variant="ghost" size="sm" type="submit" loading={deletingId === entry.id}>
									<span class="delete-label">Eliminar</span>
								</Button>
							</form>
						</div>
						{#if form?.waitlistError && form?.waitlistId === entry.id}
							<p class="row-error">{form.waitlistError}</p>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</DataTable>
</AdminSection>
