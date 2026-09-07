<script lang="ts">
	/**
	 * Invitaciones enviadas, con reenvío y revocación por fila.
	 *
	 * La columna de caducidad va en palabras («caduca en 3 días») y no con la
	 * fecha: una invitación no se reenvía porque venza un martes, se reenvía
	 * porque le quedan dos días.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminBlock from './admin-block.svelte';
	import {
		INVITE_ROLES,
		formatDay,
		invitationStatusLabel,
		invitationStatusTone,
		type InvitationItem
	} from '../admin';
	import { describeInvitations, daysSince, formatDeadline } from '../desk';

	interface Props {
		invitations: InvitationItem[];
		/** `form` de la página, para el error por fila. */
		form: Record<string, unknown> | null;
	}

	let { invitations, form }: Props = $props();

	/*
	 * La columna de estado solo aparece cuando hay algo que no esté pendiente.
	 * En una tabla titulada «invitaciones pendientes», cuya frase de arriba ya
	 * cuenta cuántas hay sin aceptar, una insignia «PENDIENTE» en cada fila era
	 * la tercera vez que se decía lo mismo en la misma pantalla.
	 */
	const mixed = $derived(invitations.some((inv) => inv.status !== 'pending'));

	let resendingId = $state<string | null>(null);
	let revokingId = $state<string | null>(null);

	/** «customer» es el nombre del rol en el backend, no el que se lee aquí. */
	const roleLabel = (role: string) => INVITE_ROLES.find((r) => r.value === role)?.label ?? role;
</script>

<AdminBlock title="Invitaciones pendientes" summary={describeInvitations(invitations)}>
	<DataTable caption="Invitaciones enviadas y en qué estado está cada una">
		<thead>
			<tr>
				<th>Correo</th>
				<th>Rol</th>
				{#if mixed}
					<th>Estado</th>
				{/if}
				<th>Caducidad</th>
				<th><span class="sr-only">Acciones</span></th>
			</tr>
		</thead>
		<tbody>
			{#each invitations as inv (inv.id)}
				{@const pending = inv.status === 'pending'}
				{@const expired = (daysSince(inv.expiresAt) ?? 0) >= 0}
				<tr>
					<td class="cell-email">{inv.email}</td>
					<td>{roleLabel(inv.role)}</td>
					{#if mixed}
						<td>
							<Badge tone={invitationStatusTone(inv.status)}>
								{invitationStatusLabel(inv.status)}
							</Badge>
						</td>
					{/if}
					<!-- La caducidad solo dice algo mientras la invitación sigue en pie:
					     una aceptada o revocada no caduca, y su estado ya lo cuenta. -->
					<td
						class="cell-age"
						class:aged={pending && expired}
						title={pending ? formatDay(inv.expiresAt) : undefined}
					>
						{pending ? formatDeadline(inv.expiresAt) : '—'}
					</td>
					<td class="cell-actions">
						<div class="row-actions">
							<form
								method="POST"
								action="?/resendInvitation"
								use:enhance={() => {
									resendingId = inv.id;
									return async ({ update }) => {
										resendingId = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="id" value={inv.id} />
								<button class="row-action" type="submit" disabled={resendingId === inv.id}>
									{resendingId === inv.id ? 'Reenviando…' : 'Reenviar'}
								</button>
							</form>
							<form
								method="POST"
								action="?/revokeInvitation"
								use:enhance={() => {
									revokingId = inv.id;
									return async ({ update }) => {
										revokingId = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="id" value={inv.id} />
								<button class="row-action danger" type="submit" disabled={revokingId === inv.id}>
									Revocar
								</button>
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
</AdminBlock>
