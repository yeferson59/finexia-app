<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
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
		<Button
			variant="secondary"
			size="sm"
			type="button"
			onclick={() => (showInviteForm = !showInviteForm)}
		>
			{showInviteForm ? 'Cancelar' : '+ Invitar usuario'}
		</Button>
	{/snippet}
</PageHeader>

{#if showInviteForm}
	<InviteUserForm {form} onSuccess={() => (showInviteForm = false)} />
{/if}

{#if data.invitations.length > 0}
	<InvitationsTable invitations={data.invitations} {form} />
{/if}

{#if data.waitlist.length > 0}
	<WaitlistTable waitlist={data.waitlist} {form} />
{/if}

<UsersTable users={data.users} meta={data.meta} {form} />
