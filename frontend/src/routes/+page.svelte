<script lang="ts">
	import { onMount } from 'svelte';
	import {
		SITE_URL,
		SITE_NAME,
		DEFAULT_TITLE,
		DEFAULT_DESCRIPTION,
		OG_IMAGE,
		LOCALE,
		absoluteUrl
	} from '$lib/seo';
	import '$lib/features/landing/landing.css';
	import {
		Ticker,
		Header,
		Hero,
		Benefits,
		HowItWorks,
		Metrics,
		Faq,
		FinalCta,
		Footer
	} from '$lib/features/landing';

	const canonical = absoluteUrl('/');

	const faqs = [
		{
			q: '¿Cuándo estará disponible Finexia?',
			a: 'El lanzamiento oficial está previsto para el 1 de octubre de 2026. Quienes se unan a la lista de espera recibirán acceso anticipado antes que el público general.'
		},
		{
			q: '¿Tiene algún costo unirse ahora?',
			a: 'No. Unirse a la lista de espera es totalmente gratuito y sin compromiso. Solo necesitamos tu correo para avisarte el día del lanzamiento.'
		},
		{
			q: '¿Finexia se conecta a mis brokers o plataformas?',
			a: 'No. Finexia nunca accede a tus plataformas ni te pide credenciales. Tú registras manualmente dónde tienes tus activos, así que la información siempre está bajo tu control.'
		},
		{
			q: '¿Puedo agrupar activos de varias plataformas en un mismo portafolio?',
			a: 'Sí, esa es la idea. Creas los portafolios que tienes en mente y añades a cada uno los activos que quieras, sin importar en qué broker o plataforma estén registrados.'
		},
		{
			q: '¿Habrá conexión automática con plataformas en el futuro?',
			a: 'Es algo que estamos contemplando para más adelante. Por ahora todo el registro es manual para que tú mantengas el control, pero tenemos en mente ofrecer en el futuro una zona centralizada y automatizada que conecte con plataformas. Lo evaluaremos próximamente según las condiciones técnicas y de seguridad que permitan hacerlo bien.'
		}
	];

	onMount(() => {
		const revealIo = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						entry.target.classList.add('in');
						revealIo.unobserve(entry.target);
					}
				});
			},
			{ threshold: 0.12 }
		);
		document.querySelectorAll('.reveal').forEach((el, i) => {
			(el as HTMLElement).style.transitionDelay = Math.min(i * 60, 240) + 'ms';
			revealIo.observe(el);
		});

		return () => revealIo.disconnect();
	});

	// Structured data — single @graph keeps Organization, WebSite and FAQPage
	// in one script and lets Google relate them via @id references.
	const jsonLd = JSON.stringify({
		'@context': 'https://schema.org',
		'@graph': [
			{
				'@type': 'Organization',
				'@id': `${SITE_URL}/#organization`,
				name: SITE_NAME,
				url: SITE_URL,
				logo: `${SITE_URL}/favicon.svg`,
				description: DEFAULT_DESCRIPTION
			},
			{
				'@type': 'WebSite',
				'@id': `${SITE_URL}/#website`,
				url: SITE_URL,
				name: SITE_NAME,
				inLanguage: 'es',
				publisher: { '@id': `${SITE_URL}/#organization` }
			},
			{
				'@type': 'FAQPage',
				'@id': `${SITE_URL}/#faq`,
				mainEntity: faqs.map((faq) => ({
					'@type': 'Question',
					name: faq.q,
					acceptedAnswer: { '@type': 'Answer', text: faq.a }
				}))
			}
		]
	});
</script>

<svelte:head>
	<title>{DEFAULT_TITLE}</title>
	<meta name="description" content={DEFAULT_DESCRIPTION} />
	<link rel="canonical" href={canonical} />

	<!-- Open Graph -->
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content={SITE_NAME} />
	<meta property="og:locale" content={LOCALE} />
	<meta property="og:url" content={canonical} />
	<meta property="og:title" content={DEFAULT_TITLE} />
	<meta property="og:description" content={DEFAULT_DESCRIPTION} />
	<meta property="og:image" content={OG_IMAGE} />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />

	<!-- Twitter -->
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={DEFAULT_TITLE} />
	<meta name="twitter:description" content={DEFAULT_DESCRIPTION} />
	<meta name="twitter:image" content={OG_IMAGE} />

	<!-- Structured data (app-controlled JSON-LD, no user input) -->
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html `<script type="application/ld+json">${jsonLd}</scr` + 'ipt>'}
</svelte:head>

<div class="scanlines" aria-hidden="true"></div>

<Ticker />
<Header />

<main>
	<Hero />

	<div class="wrap"><div class="divider"></div></div>
	<Benefits />

	<div class="wrap"><div class="divider"></div></div>
	<HowItWorks />

	<div class="wrap"><div class="divider"></div></div>
	<Metrics />

	<div class="wrap"><div class="divider"></div></div>
	<Faq {faqs} />

	<FinalCta />
</main>

<Footer />

<style>
	:global(html) {
		scroll-behavior: smooth;
	}
	:global(body) {
		line-height: 1.55;
		overflow-x: hidden;
	}

	.scanlines {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
		background: repeating-linear-gradient(
			0deg,
			transparent,
			transparent 3px,
			rgba(255, 255, 255, 0.006) 3px,
			rgba(255, 255, 255, 0.006) 4px
		);
	}

	main {
		position: relative;
		z-index: 1;
	}
</style>
