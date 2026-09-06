<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import {
		MarketCredentials,
		MCPTokensSection,
		OAuthGrantsSection,
		PasswordSection,
		ProfileSection,
		SessionsSection,
		SettingsGroup,
		TwoFactorSection,
		describeAccess,
		describeConnections,
		describeIdentity
	} from '$lib/features/settings';

	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();

	/*
	 * Cada grupo abre diciendo cómo está lo suyo. Es lo que sustituye a recorrer
	 * ocho tarjetas para averiguar si la 2FA está puesta o si queda una sesión
	 * abierta en un ordenador que ya no usas.
	 */
	const identity = $derived(describeIdentity(data.user));
	const access = $derived(describeAccess(data.twoFactor, data.sessions));
	const connections = $derived(
		describeConnections(data.marketCredentials, data.mcpTokens, data.oauthGrants)
	);
</script>

<svelte:head>
	<title>Configuración - FINEXIA</title>
	<meta name="description" content="Tu cuenta, cómo entras a ella y qué tienes conectado" />
</svelte:head>

<!-- Sin filete: el que abre el primer grupo hace ese trabajo, y dos rayas a
     cuarenta píxeles una de otra no separan nada. -->
<PageHeader
	title="Configuración"
	subtitle="Tu cuenta, cómo entras a ella y qué le has dejado conectado."
	divider={false}
/>

<SettingsGroup title="Tu perfil" summary={identity}>
	<ProfileSection user={data.user} {form} />
</SettingsGroup>

<SettingsGroup title="Cómo entras" summary={access}>
	<PasswordSection {form} />
	<TwoFactorSection twoFactor={data.twoFactor} {form} />
	<SessionsSection sessions={data.sessions} {form} />
</SettingsGroup>

<SettingsGroup title="Lo que tienes conectado" summary={connections}>
	<MarketCredentials credentials={data.marketCredentials} {form} />
	<MCPTokensSection tokens={data.mcpTokens} mcpUrl={data.mcpUrl} {form} />
	<OAuthGrantsSection grants={data.oauthGrants} {form} />
</SettingsGroup>
