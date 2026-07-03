<script lang="ts">
	import { resolve } from '$app/paths';
	import { CONTACT_EMAIL, SITE_NAME, absoluteUrl } from '$lib/seo';

	const canonical = absoluteUrl('/terminos');
</script>

<svelte:head>
	<title>Términos y Condiciones — {SITE_NAME}</title>
	<meta name="description" content="Términos y Condiciones de uso del servicio Finexia." />
	<link rel="canonical" href={canonical} />
	<meta name="robots" content="index,follow" />
</svelte:head>

<h1>Términos y Condiciones</h1>
<p class="legal-updated">Última actualización: 3 de julio de 2026</p>

<p class="legal-intro">
	Estos Términos y Condiciones regulan el uso del servicio {SITE_NAME}. Al crear una cuenta o
	utilizar la aplicación, aceptas estos términos. Si no estás de acuerdo, por favor no utilices el
	servicio.
</p>

<h2>1. Qué es Finexia</h2>
<p>
	{SITE_NAME} es una herramienta que te permite registrar manualmente tus activos y agruparlos en portafolios
	para visualizar tu patrimonio de forma consolidada. Finexia <strong>no</strong> se conecta a tus brokers
	ni plataformas, no te pide credenciales de terceros y no ejecuta operaciones financieras en tu nombre.
</p>

<h2>2. No es asesoría financiera</h2>
<p>
	La información que muestra {SITE_NAME} tiene fines meramente informativos y de organización personal.
	<strong>No constituye asesoría financiera, de inversión, contable ni tributaria.</strong> Las decisiones
	que tomes con base en la información registrada son de tu exclusiva responsabilidad. Te recomendamos
	consultar a un profesional antes de tomar decisiones de inversión.
</p>

<h2>3. Tu cuenta</h2>
<ul>
	<li>Debes proporcionar información veraz al registrarte y mantenerla actualizada.</li>
	<li>
		Eres responsable de mantener la confidencialidad de tu contraseña y de la actividad en tu
		cuenta.
	</li>
	<li>Debes ser mayor de edad o contar con autorización legal para usar el servicio.</li>
	<li>Notifícanos de inmediato ante cualquier uso no autorizado de tu cuenta.</li>
</ul>

<h2>4. Uso aceptable</h2>
<p>Al usar {SITE_NAME}, te comprometes a no:</p>
<ul>
	<li>Utilizar el servicio con fines ilícitos o no autorizados.</li>
	<li>Intentar acceder a cuentas, datos o sistemas que no te pertenezcan.</li>
	<li>Interferir con el funcionamiento, la seguridad o la disponibilidad del servicio.</li>
	<li>Registrar información que infrinja derechos de terceros.</li>
</ul>

<h2>5. Datos personales</h2>
<p>
	El tratamiento de tus datos personales se rige por nuestra
	<a href={resolve('/privacidad')}>Política de Tratamiento de Datos Personales</a>, conforme a la
	Ley 1581 de 2012 de Colombia. Te recomendamos leerla junto con estos términos.
</p>

<h2>6. Disponibilidad del servicio</h2>
<p>
	Nos esforzamos por mantener el servicio disponible, pero puede presentar interrupciones por
	mantenimiento, mejoras o causas fuera de nuestro control. El servicio se ofrece «tal cual», sin
	garantías de disponibilidad ininterrumpida.
</p>

<h2>7. Limitación de responsabilidad</h2>
<p>
	En la máxima medida permitida por la ley, {SITE_NAME} no será responsable por pérdidas o daños derivados
	del uso o la imposibilidad de uso del servicio, ni por decisiones tomadas con base en la información
	que tú mismo registras.
</p>

<h2>8. Cambios en los términos</h2>
<p>
	Podemos actualizar estos términos en cualquier momento. Publicaremos la versión vigente en esta
	página con su fecha de última actualización. El uso continuado del servicio implica la aceptación
	de los cambios.
</p>

<h2>9. Contacto</h2>
<p>
	Para cualquier duda sobre estos términos, escríbenos a
	<a href="mailto:{CONTACT_EMAIL}">{CONTACT_EMAIL}</a>.
</p>
