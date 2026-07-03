<script lang="ts">
	import { resolve } from '$app/paths';
	import { CONTACT_EMAIL, SITE_NAME, absoluteUrl } from '$lib/seo';

	const canonical = absoluteUrl('/cookies');
</script>

<svelte:head>
	<title>Aviso de Cookies — {SITE_NAME}</title>
	<meta name="description" content="Aviso de uso de cookies en Finexia." />
	<link rel="canonical" href={canonical} />
	<meta name="robots" content="index,follow" />
</svelte:head>

<h1>Aviso de Cookies</h1>
<p class="legal-updated">Última actualización: 3 de julio de 2026</p>

<p class="legal-intro">
	Este aviso explica cómo {SITE_NAME} utiliza cookies y tecnologías similares de almacenamiento en tu
	navegador.
</p>

<h2>1. Qué son las cookies</h2>
<p>
	Las cookies son pequeños archivos que un sitio web guarda en tu navegador. Sirven, por ejemplo,
	para mantener tu sesión iniciada o recordar preferencias. Tecnologías similares como
	<em>localStorage</em> cumplen funciones parecidas.
</p>

<h2>2. Qué cookies utilizamos</h2>
<h3>Cookies estrictamente necesarias</h3>
<p>
	Son imprescindibles para que el servicio funcione. Las usamos para mantener tu sesión iniciada de
	forma segura y recordar decisiones básicas, como que ya viste este aviso de cookies. No requieren
	tu consentimiento porque sin ellas el servicio no puede operar.
</p>

<h3>Cookies de analítica o de terceros</h3>
<p>
	Actualmente {SITE_NAME} no utiliza cookies de analítica ni de terceros con fines publicitarios o de
	seguimiento. Si en el futuro las incorporamos, actualizaremos este aviso y, cuando la ley lo exija,
	te pediremos tu consentimiento previo.
</p>

<h2>3. Cómo gestionar las cookies</h2>
<p>
	Puedes eliminar o bloquear las cookies desde la configuración de tu navegador. Ten en cuenta que
	si bloqueas las cookies estrictamente necesarias, algunas funciones del servicio (como iniciar
	sesión) podrían no funcionar correctamente.
</p>

<h2>4. Más información</h2>
<p>
	Para saber cómo tratamos tus datos personales, consulta nuestra
	<a href={resolve('/privacidad')}>Política de Tratamiento de Datos Personales</a>. Si tienes dudas,
	escríbenos a
	<a href="mailto:{CONTACT_EMAIL}">{CONTACT_EMAIL}</a>.
</p>
