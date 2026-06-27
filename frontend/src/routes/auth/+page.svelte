<script lang="ts">
	import LoginRegister from '$components/auth/login-register.svelte';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<svelte:head>
	<title>Iniciar sesión - FINEXIA</title>
	<meta name="description" content="Inicia sesión o crea una cuenta en FINEXIA" />
</svelte:head>

<LoginRegister {form} />
