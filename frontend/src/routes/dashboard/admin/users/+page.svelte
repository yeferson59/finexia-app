<script lang="ts">
	/*
	 * Los tres listados son un mismo recorrido: alguien pide acceso, se le
	 * invita y acaba con cuenta. Van en ese orden y no al revés —los registrados
	 * abrían la pantalla— porque lo único que espera por ti está al principio.
	 */
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
	title="Usuarios"
	subtitle="Quién puede entrar a Finexia, desde que lo pide hasta que tiene cuenta."
>
	{#snippet actions()}
		<Button type="button" onclick={() => (showInviteForm = true)}>Invitar a alguien</Button>
	{/snippet}
</PageHeader>

<Modal
	open={showInviteForm}
	title="Invitar a alguien"
	description="Le enviaremos un enlace de un solo uso para que cree su propia contraseña."
	onClose={() => (showInviteForm = false)}
>
	<InviteUserForm
		{form}
		onCancel={() => (showInviteForm = false)}
		onSuccess={() => (showInviteForm = false)}
	/>
</Modal>

{#if data.waitlist.length > 0}
	<WaitlistTable waitlist={data.waitlist} {form} />
{/if}

{#if data.invitations.length > 0}
	<InvitationsTable invitations={data.invitations} {form} />
{/if}

<UsersTable users={data.users} meta={data.meta} {form} />
