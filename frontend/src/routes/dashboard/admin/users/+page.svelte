<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
	import { InvitationsTable, InviteUserForm, UsersTable, WaitlistTable } from '$lib/features/admin';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	let showInviteForm = $state(false);
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
		<Button variant="secondary" size="sm" type="button" onclick={() => (showInviteForm = true)}>
			Invitar usuario
		</Button>
	{/snippet}
</PageHeader>

<Modal
	open={showInviteForm}
	title="Invitar a un nuevo usuario"
	description="Enviaremos un enlace seguro de un solo uso para que la persona cree su propia contraseña."
	onClose={() => (showInviteForm = false)}
>
	<InviteUserForm
		{form}
		onCancel={() => (showInviteForm = false)}
		onSuccess={() => (showInviteForm = false)}
	/>
</Modal>

{#if data.invitations.length > 0}
	<InvitationsTable invitations={data.invitations} {form} />
{/if}

{#if data.waitlist.length > 0}
	<WaitlistTable waitlist={data.waitlist} {form} />
{/if}

<UsersTable users={data.users} meta={data.meta} {form} />
