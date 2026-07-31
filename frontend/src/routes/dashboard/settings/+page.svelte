<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import {
		AppearanceSection,
		MarketCredentials,
		PasswordSection,
		ProfileSection,
		SessionsSection,
		TwoFactorSection
	} from '$lib/features/settings';

	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Configuración - FINEXIA</title>
	<meta name="description" content="Preferencias de cuenta y notificaciones de FINEXIA" />
</svelte:head>

<PageHeader
	title="Configuración"
	subtitle="Ajusta preferencias de experiencia, alertas y seguridad de cuenta."
/>

<div class="settings-layout">
	<ProfileSection user={data.user} {form} />
	<AppearanceSection />
	<PasswordSection {form} />
	<TwoFactorSection twoFactor={data.twoFactor} {form} />
	<MarketCredentials credentials={data.marketCredentials} {form} />
	<SessionsSection sessions={data.sessions} {form} />
</div>

<style>
	.settings-layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	@media (max-width: 1024px) {
		.settings-layout {
			grid-template-columns: 1fr;
		}
	}
</style>
