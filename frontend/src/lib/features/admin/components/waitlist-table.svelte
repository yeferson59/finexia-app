<script lang="ts">
	/**
	 * Lista de espera: correos que pidieron acceso y aún no tienen invitación.
	 * Invitar desde aquí reutiliza la misma action que el formulario de alta;
	 * eliminar saca la entrada de la lista (typos, duplicados o bajas), y el
	 * correo queda libre para volver a apuntarse.
	 *
	 * Es el primer bloque de la pantalla porque es el único donde alguien está
	 * esperando por ti: lo demás son cuentas que ya existen.
	 */
	import { enhance } from '$app/forms';
	import DataTable from '$lib/ui/data-table.svelte';
	import AdminBlock from './admin-block.svelte';
	import { formatDay, type WaitlistItem } from '../admin';
	import { describeWaitlist, formatAge, isStale } from '../desk';

	interface Props {
		waitlist: WaitlistItem[];
		/** `form` de la página, para el error por fila. */
		form: Record<string, unknown> | null;
	}

	let { waitlist, form }: Props = $props();

	let invitingId = $state<string | null>(null);
	let deletingId = $state<string | null>(null);
</script>

<AdminBlock title="Lista de espera" summary={describeWaitlist(waitlist)}>
	<DataTable caption="Correos que pidieron acceso y siguen esperando una invitación">
		<thead>
			<tr>
				<th>Correo</th>
				<th>Pidió acceso</th>
				<th><span class="sr-only">Acciones</span></th>
			</tr>
		</thead>
		<tbody>
			{#each waitlist as entry (entry.id)}
				<tr>
					<td class="cell-email">{entry.email}</td>
					<!-- La fecha exacta se queda en el `title`: lo que se decide con
					     esta columna es a quién lleva más tiempo esperando. -->
					<td
						class="cell-age"
						class:aged={isStale(entry.createdAt)}
						title={formatDay(entry.createdAt)}
					>
						{formatAge(entry.createdAt)}
					</td>
					<td class="cell-actions">
						<div class="row-actions">
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
								<button class="row-action" type="submit" disabled={invitingId === entry.id}>
									{invitingId === entry.id ? 'Invitando…' : 'Invitar'}
								</button>
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
								<button class="row-action danger" type="submit" disabled={deletingId === entry.id}>
									Eliminar
								</button>
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
</AdminBlock>
