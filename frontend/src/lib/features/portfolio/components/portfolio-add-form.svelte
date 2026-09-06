<script lang="ts">
	/*
	 * Alta de un portafolio.
	 *
	 * Eran tres tarjetas apiladas y centradas en la columna, con las leyendas EN
	 * VERSALITAS ÁMBAR, un botón «Volver» flotando arriba a la izquierda sin
	 * alinear con nada, y su propia copia del CSS de los campos. Ahora es el
	 * carril de configuración, alineado a la izquierda como el resto del panel.
	 *
	 * Tres cosas que no eran de estilo:
	 *
	 *  - `errors` se declaraba y no se rellenaba nunca, así que el marcado de
	 *    error de cada campo no podía aparecer.
	 *  - `submitSuccess` tampoco se ponía nunca a `true`: el aviso verde de
	 *    «Portafolio creado exitosamente» era código muerto. Y sobra, porque al
	 *    crear se navega al listado, donde el portafolio ya está.
	 *  - La action devolvía `{ success: false }` y el formulario no leía `form`,
	 *    así que un alta rechazada no dejaba ni rastro en pantalla.
	 */
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/button.svelte';
	import { SUPPORTED_CURRENCIES, resolveDisplayCurrency } from '$lib/shared/currency';
	import type { ActionForm } from '$lib/shared/form';
	import { PORTFOLIO_TYPES } from '../portfolio';
	import PortfolioFormSection from './portfolio-form-section.svelte';
	import PortfolioRiskPicker from './portfolio-risk-picker.svelte';
	import PortfolioGoalFieldset from './portfolio-goal-fieldset.svelte';

	let {
		risks,
		/** Moneda de la cuenta: el arranque más probable para un portafolio nuevo. */
		defaultCurrency,
		form
	}: {
		risks: { id: string; name: string; description: string }[];
		defaultCurrency?: string;
		form: ActionForm;
	} = $props();

	let name = $state('');
	let description = $state('');
	let type = $state('stocks_etfs');
	// Semilla, no vínculo: a partir de aquí manda lo que elija el usuario.
	let currency = $state(untrack(() => resolveDisplayCurrency(defaultCurrency)));
	let riskId = $state('');
	let targetAmount = $state('');
	let isDefault = $state(false);
	let isSubmitting = $state(false);

	const error = $derived((form?.error as string) ?? '');
</script>

<form
	method="POST"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update();
			isSubmitting = false;
		};
	}}
>
	<PortfolioFormSection
		title="Cómo lo llamas"
		description="Puedes tener tantos portafolios como quieras: uno para el retiro, otro para cripto, otro para el fondo de emergencia. El nombre es con el que lo verás en el panel."
	>
		<div class="field">
			<label for="name">Nombre</label>
			<input
				type="text"
				id="name"
				name="name"
				bind:value={name}
				placeholder="Retiro"
				disabled={isSubmitting}
				required
				minlength="1"
			/>
		</div>

		<div class="field">
			<label for="description">Descripción <span class="optional">(opcional)</span></label>
			<textarea
				id="description"
				name="description"
				bind:value={description}
				placeholder="Para qué es este dinero y cuándo piensas tocarlo."
				disabled={isSubmitting}
				rows="3"></textarea>
		</div>
	</PortfolioFormSection>

	<PortfolioFormSection
		title="Qué guarda y en qué moneda"
		description="La moneda es en la que hablan sus totales: Finexia convierte a ella lo que compres en otra, así que elígela por dónde quieres leer las cifras, no por dónde compras."
	>
		<div class="pair">
			<div class="field">
				<label for="type">Tipo</label>
				<select id="type" name="type" bind:value={type} disabled={isSubmitting} required>
					{#each PORTFOLIO_TYPES as option (option.value)}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<div class="field">
				<label for="currency">Moneda</label>
				<select
					id="currency"
					name="currency"
					bind:value={currency}
					disabled={isSubmitting}
					required
				>
					{#each SUPPORTED_CURRENCIES as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
			</div>
		</div>
	</PortfolioFormSection>

	<PortfolioFormSection
		title="Cómo lo quieres seguir"
		description="El riesgo es una etiqueta tuya: aparece en la ficha del portafolio para reconocer de un vistazo qué es cada uno. La meta y el portafolio por defecto puedes dejarlos en blanco."
	>
		<PortfolioRiskPicker {risks} bind:selected={riskId} disabled={isSubmitting} />
		<PortfolioGoalFieldset {currency} bind:targetAmount bind:isDefault disabled={isSubmitting} />
	</PortfolioFormSection>

	<div class="close">
		{#if error}
			<p class="feedback error">{error}</p>
		{/if}

		<div class="actions">
			<Button type="submit" loading={isSubmitting}>
				{isSubmitting ? 'Creando…' : 'Crear portafolio'}
			</Button>
			<a class="cancel" href={resolve('/dashboard/portfolios')}>Cancelar</a>
		</div>
	</div>
</form>

<style>
	.close {
		padding-top: 2.25rem;
		border-top: 1px solid var(--border-strong);
	}

	/* Prosa con un filete rojo, no una caja de alerta: el idioma que ya hablan
	   configuración y notificaciones. */
	/* Solo el margen: la forma y el color de un aviso los pone
	   `routes/layout.css`. */
	.feedback {
		margin: 0 0 1.25rem;
	}

	/*
	 * A la izquierda y en el orden en que se leen, no anclados al borde derecho
	 * de la página: el ojo viene bajando por el carril de campos y ahí es donde
	 * termina. Y «Cancelar» es un enlace al listado, no un botón con `goto`.
	 */
	.actions {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	/* Sin el halo ámbar de `ui/button`, como en configuración. */
	.actions :global(.btn-primary) {
		box-shadow: none;
	}

	.cancel {
		font-size: 0.85rem;
		color: var(--text-muted);
		text-decoration: none;
		transition: color 0.2s ease;
	}

	.cancel:hover {
		color: var(--text);
	}

	@media (prefers-reduced-motion: reduce) {
		.cancel {
			transition: none;
		}
	}
</style>
