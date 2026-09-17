<script lang="ts">
	import { page } from '$app/state';
	import { LoginRegister } from '$lib/features/auth';
	import type { ActionData, PageData } from './$types';

	let { form, data }: { form: ActionData; data: PageData } = $props();

	/* Lo que dicen las pantallas que terminan aquí. Cada una vuelve con su marca
	   en la URL y la frase dice qué ha pasado y qué toca ahora. */
	const notice = $derived.by(() => {
		const params = page.url.searchParams;
		if (params.has('registered'))
			return 'Cuenta creada. Te enviamos un enlace para verificar tu correo; ábrelo antes de iniciar sesión.';
		if (params.has('verified')) return 'Correo verificado. Ya puedes iniciar sesión.';
		if (params.has('reset')) return 'Contraseña cambiada. Inicia sesión con la nueva.';
		if (params.has('invited'))
			return 'Cuenta activada. Inicia sesión con tu correo y la contraseña que elegiste.';
		return undefined;
	});
</script>

<svelte:head>
	<title>Iniciar sesión — Finexia</title>
	<meta name="description" content="Inicia sesión en Finexia o pide acceso a la beta." />
</svelte:head>

<LoginRegister {form} {notice} selfRegistrationEnabled={data.selfRegistrationEnabled} />
