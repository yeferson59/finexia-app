<script lang="ts">
	import { resolve } from '$app/paths';
	import { SITE_NAME, CONTACT_EMAIL, OG_IMAGE, absoluteUrl } from '$lib/seo';
	import '$lib/ui/public.css';
	import Brand from '$lib/ui/brand.svelte';
	import { Footer } from '$lib/features/landing';
	import { BoldPayment, PaymentResult, SupportNotes } from '$lib/features/support';

	let { data, form } = $props();

	const title = `Apoyar el proyecto — ${SITE_NAME}`;
	const description =
		'Finexia es gratis y la construye una persona en Colombia. Puedes apoyarla con un pago por Bold.';
	const canonical = absoluteUrl('/apoyar');
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description} />
	<link rel="canonical" href={canonical} />
	<meta name="theme-color" content="#eef0ec" />
	<meta property="og:type" content="website" />
	<meta property="og:url" content={canonical} />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:image" content={OG_IMAGE} />
</svelte:head>

<div class="lp">
	<header class="top">
		<div class="lp-wrap nav">
			<a href={resolve('/')} class="home" aria-label="Finexia, volver al inicio">
				<Brand />
			</a>
			{#if data.signedIn}
				<a href={resolve('/dashboard')} class="back">Volver al panel</a>
			{:else}
				<a href={resolve('/')} class="back">Volver al inicio</a>
			{/if}
		</div>
	</header>

	<main>
		<div class="lp-wrap intro">
			<h1>Apoyar Finexia</h1>
			<p>
				Finexia es gratis y la construye una sola persona en Colombia. Si te sirve para ordenar tu
				patrimonio, puedes ayudar a pagar lo que cuesta mantenerla abierta.
			</p>
		</div>

		{#if data.result}
			<div class="lp-wrap outcome">
				<PaymentResult result={data.result} contactEmail={CONTACT_EMAIL} />
			</div>
		{/if}

		<BoldPayment enabled={data.enabled} error={form?.error} contactEmail={CONTACT_EMAIL} />

		<div class="lp-wrap rest">
			<SupportNotes contactEmail={CONTACT_EMAIL} />
		</div>
	</main>

	<Footer />
</div>

<style>
	.top {
		border-bottom: 1px solid var(--lp-rule);
	}

	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		height: 72px;
	}

	.home {
		display: inline-flex;
		color: var(--lp-ink);
	}

	.back {
		font-size: 15px;
		color: var(--lp-ink-2);
	}

	.back:hover {
		color: var(--lp-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.intro {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 440px);
		align-items: end;
		gap: 24px 96px;
		padding-block: clamp(48px, 10vh, 112px) 56px;
	}

	h1 {
		margin: 0;
		font-size: clamp(40px, 6vw, 80px);
		font-stretch: 118%;
		font-weight: 640;
		line-height: 0.95;
		letter-spacing: -0.03em;
		text-wrap: balance;
	}

	.intro p {
		margin: 0;
		font-size: var(--lp-fs-lead);
		line-height: 1.55;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.outcome {
		padding-bottom: 48px;
	}

	.rest {
		padding-block: 88px 112px;
	}

	@media (max-width: 860px) {
		.intro {
			grid-template-columns: minmax(0, 1fr);
			padding-block: 40px;
		}
		.rest {
			padding-block: 56px 72px;
		}
	}

	@media (max-width: 640px) {
		.nav {
			height: 60px;
		}
	}
</style>
