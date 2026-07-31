<script lang="ts">
	/** Invitaciones pendientes, con reenvío y revocación por fila. */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import Button from '$lib/ui/button.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminSection from './admin-section.svelte';
	import {
		formatDay,
		invitationStatusLabel,
		invitationStatusTone,
		type InvitationItem
	} from '../admin';

	interface Props {
		invitations: InvitationItem[];
		/** `form` de la página, para el error por fila. */
		form: Record<string, unknown> | null;
	}

	let { invitations, form }: Props = $props();

	let invitingId = $state<string | null>(null);
</script>

<AdminSection title="Invitaciones pendientes">
	<DataTable>
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
			{#each invitations as inv (inv.id)}
				<tr>
					<td class="cell-email">{inv.email}</td>
					<td>
						<Badge tone={inv.role === 'admin' ? 'amber' : 'neutral'}>{inv.role}</Badge>
					</td>
					<td>
						<Badge tone={invitationStatusTone(inv.status)}>
							{invitationStatusLabel(inv.status)}
						</Badge>
					</td>
					<td class="cell-date">{formatDay(inv.expiresAt)}</td>
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
								<Button variant="secondary" size="sm" type="submit" loading={invitingId === inv.id}>
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
						{#if form?.inviteError && form?.inviteId === inv.id}
							<p class="row-error">{form.inviteError}</p>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</DataTable>
</AdminSection>
