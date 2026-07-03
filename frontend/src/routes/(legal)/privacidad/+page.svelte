<script lang="ts">
	import { resolve } from '$app/paths';
	import { CONTACT_EMAIL, SITE_NAME, absoluteUrl } from '$lib/seo';

	const canonical = absoluteUrl('/privacidad');
</script>

<svelte:head>
	<title>Política de Tratamiento de Datos Personales — {SITE_NAME}</title>
	<meta
		name="description"
		content="Política de Tratamiento de Datos Personales de Finexia conforme a la Ley 1581 de 2012 (Habeas Data) de Colombia."
	/>
	<link rel="canonical" href={canonical} />
	<meta name="robots" content="index,follow" />
</svelte:head>

<h1>Política de Tratamiento de Datos Personales</h1>
<p class="legal-updated">Última actualización: 3 de julio de 2026</p>

<p class="legal-intro">
	En {SITE_NAME} respetamos tu privacidad. Esta política explica qué datos personales recolectamos, con
	qué finalidad, cómo los protegemos y cuáles son tus derechos como titular, en cumplimiento de la
	<strong>Ley 1581 de 2012</strong> y el Decreto 1377 de 2013 de la República de Colombia (régimen de
	Habeas Data).
</p>

<h2>1. Responsable del tratamiento</h2>
<p>
	El responsable del tratamiento de tus datos personales es {SITE_NAME}. Puedes contactarnos en
	cualquier momento a través del correo <a href="mailto:{CONTACT_EMAIL}">{CONTACT_EMAIL}</a> para asuntos
	relacionados con tus datos personales.
</p>

<h2>2. Qué datos recolectamos y para qué (finalidad)</h2>
<p>Recolectamos únicamente los datos necesarios para prestarte el servicio:</p>
<ul>
	<li>
		<strong>Correo electrónico:</strong> para crear tu cuenta, autenticarte, enviarte notificaciones del
		servicio y, si te uniste a la lista de espera, avisarte del lanzamiento.
	</li>
	<li>
		<strong>Nombre:</strong> para identificarte dentro de la aplicación y personalizar tu experiencia.
	</li>
	<li>
		<strong>Información de tus portafolios y activos:</strong> los datos que tú registras manualmente
		(nombres de portafolios, activos, montos y plataformas) para mostrarte tu patrimonio consolidado.
		Nunca accedemos a tus cuentas ni te pedimos credenciales de terceros.
	</li>
	<li>
		<strong>Datos técnicos mínimos:</strong> información necesaria para el funcionamiento seguro del servicio
		(por ejemplo, datos de sesión).
	</li>
</ul>
<p>
	No vendemos tus datos personales ni los usamos para finalidades distintas a las aquí descritas sin
	tu autorización previa.
</p>

<h2>3. Cómo protegemos tus datos</h2>
<p>
	Aplicamos medidas técnicas, humanas y administrativas razonables para proteger tu información
	frente a acceso no autorizado, pérdida, alteración o divulgación. Entre otras: cifrado de las
	comunicaciones (HTTPS), almacenamiento seguro de credenciales, control de acceso y buenas
	prácticas de desarrollo. Ningún sistema es infalible, pero trabajamos para mantener tu información
	protegida.
</p>

<h2>4. Tus derechos como titular</h2>
<p>Como titular de los datos, la ley colombiana te garantiza los siguientes derechos:</p>
<ul>
	<li><strong>Conocer</strong> los datos personales que tenemos sobre ti.</li>
	<li>
		<strong>Actualizar</strong> y <strong>rectificar</strong> tus datos cuando sean inexactos, incompletos
		o desactualizados.
	</li>
	<li>
		<strong>Suprimir</strong> tus datos cuando consideres que no se están tratando conforme a la ley.
	</li>
	<li><strong>Revocar la autorización</strong> que otorgaste para el tratamiento de tus datos.</li>
	<li>Solicitar prueba de la autorización otorgada, salvo las excepciones legales.</li>
	<li>
		Presentar quejas ante la Superintendencia de Industria y Comercio (SIC) por infracciones a la
		ley.
	</li>
</ul>

<h2>5. Cómo ejercer tus derechos</h2>
<p>
	Para conocer, actualizar, rectificar o suprimir tus datos, o para revocar tu autorización,
	escríbenos a <a href="mailto:{CONTACT_EMAIL}">{CONTACT_EMAIL}</a>. Atenderemos tu solicitud en los
	plazos que establece la ley. Indícanos tu nombre, el correo asociado a tu cuenta y la solicitud
	concreta que deseas realizar.
</p>

<h2>6. Autorización</h2>
<p>
	Al registrarte y marcar la casilla correspondiente, autorizas de manera libre, previa, expresa e
	informada el tratamiento de tus datos personales conforme a esta política. Puedes revocar esta
	autorización en cualquier momento por el canal de contacto indicado.
</p>

<h2>7. Vigencia y cambios</h2>
<p>
	Esta política puede actualizarse para reflejar cambios legales o del servicio. Publicaremos
	siempre la versión vigente en esta página con su fecha de última actualización. Tus datos se
	conservarán mientras exista la relación con el servicio o mientras la ley lo exija.
</p>

<div class="legal-contact">
	<p><strong>Canal de contacto para tus datos personales</strong></p>
	<p>Correo: <a href="mailto:{CONTACT_EMAIL}">{CONTACT_EMAIL}</a></p>
	<p>
		Consulta también nuestros <a href={resolve('/terminos')}>Términos y Condiciones</a> y el
		<a href={resolve('/cookies')}>Aviso de Cookies</a>.
	</p>
</div>
