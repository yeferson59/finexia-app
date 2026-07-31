<script lang="ts">
	/**
	 * Lista de espera: correos que pidieron acceso y aún no tienen invitación.
	 * Invitar desde aquí reutiliza la misma action que el formulario de alta.
	 */
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminSection from './admin-section.svelte';
	import { formatDay, type WaitlistItem } from '../admin';

	interface Props {
		waitlist: WaitlistItem[];
	}

	let { waitlist }: Props = $props();

	let invitingId = $state<string | null>(null);
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
							<Button variant="secondary" size="sm" type="submit" loading={invitingId === entry.id}>
								Invitar
							</Button>
						</form>
					</td>
				</tr>
			{/each}
		</tbody>
	</DataTable>
</AdminSection>
