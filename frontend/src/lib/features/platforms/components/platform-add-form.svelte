<script lang="ts">
	/*
	 * Alta de una plataforma.
	 *
	 * Era una tarjeta con sombra y tres campos dentro, con las etiquetas en
	 * Mayúsculas De Título y un asterisco rojo en las obligatorias. Ahora es una
	 * columna: tres campos no necesitan el carril de configuración, que gana su
	 * sitio cuando hay varios bloques que hojear. Lo que sí hacía falta es decir
	 * antes de empezar qué es una plataforma aquí.
	 *
	 * Dos cosas que no eran de estilo:
	 *
	 *  - El desplegable arrancaba en «Bróker», que es la *etiqueta* del tipo; los
	 *    valores de las opciones son las claves (`broker`, `neobank`…). Como no
	 *    coincidía con ninguna, el `<select>` abría sin nada seleccionado y, al
	 *    llevar `required`, el navegador cortaba el envío con su propio globo.
	 *    Ahora arranca en la clave.
	 *  - La action devolvía `{ error }` y este formulario no lo leía en ningún
	 *    sitio, así que un alta rechazada por el backend no dejaba rastro en la
	 *    pantalla.
	 */
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/button.svelte';
	import { PLATFORM_TYPES } from '../platforms';
	import type { ActionForm } from '$lib/shared/form';

	let { form }: { form: ActionForm } = $props();

	let name = $state('');
	let description = $state('');
	// La clave, no la etiqueta: es lo que llevan los `value` de las opciones.
	let type = $state('broker');
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
	<!--
		La duda real de esta pantalla, antes del primer campo. Finexia no se
		conecta a ningún bróker ni pide credenciales: una plataforma aquí es solo
		el nombre del sitio donde tienes el dinero, para poder repartir tus
		posiciones entre ellos. Sin decirlo, quien llega espera que el siguiente
		paso sea escribir la contraseña de su bróker.
	-->
	<p class="lead">
		Una plataforma es el sitio donde guardas tu dinero: un bróker, una casa de bolsa, una billetera
		cripto. Finexia no se conecta con ella ni te pedirá sus claves; solo le pone nombre para que
		puedas repartir tus posiciones.
	</p>

	<div class="fields rail-fields">
		<div class="field">
			<label for="name">Nombre</label>
			<input
				id="name"
				name="name"
				type="text"
				bind:value={name}
				placeholder="Interactive Brokers"
				disabled={isSubmitting}
				required
				minlength="2"
			/>
			<p class="hint">Como la llamas tú. Es el nombre con el que aparecerá en tus posiciones.</p>
		</div>

		<div class="field">
			<label for="type">Tipo</label>
			<select id="type" name="type" bind:value={type} disabled={isSubmitting} required>
				{#each PLATFORM_TYPES.entries() as [key, label] (key)}
					<option value={key}>{label}</option>
				{/each}
			</select>
		</div>

		<div class="field">
			<!-- Se marca lo opcional, no lo obligatorio: de tres campos, dos lo
				     son, así que dos asteriscos rojos señalaban casi todo. -->
			<label for="description">Notas <span class="optional">(opcional)</span></label>
			<textarea
				id="description"
				name="description"
				bind:value={description}
				placeholder="Qué tienes aquí, en qué moneda opera, lo que te sirva para reconocerla."
				disabled={isSubmitting}
				rows="3"></textarea>
		</div>

		{#if error}
			<p class="feedback error">{error}</p>
		{/if}

		<div class="actions">
			<Button type="submit" loading={isSubmitting}>
				{isSubmitting ? 'Creando…' : 'Crear plataforma'}
			</Button>
			<a class="cancel" href={resolve('/dashboard/platforms')}>Cancelar</a>
		</div>
	</div>
</form>

<style>
	.lead {
		max-width: 62ch;
		margin: 0 0 2rem;
		font-size: 0.9rem;
		line-height: 1.65;
		color: var(--text-muted);
	}

	/*
	 * La columna de campos es la compartida (`rail-fields`, en
	 * `routes/layout.css`): aquí estaba copiada entera, con sus etiquetas, sus
	 * `input` y sus estados, idéntica a la del carril hasta el último píxel. Lo
	 * único de esta pantalla es dónde se corta: el nombre de un bróker no se
	 * escribe en una caja de mil píxeles.
	 */
	.fields {
		max-width: 34rem;
	}

	/* Solo el margen: la forma y el color de un aviso los pone
	   `routes/layout.css`. */
	.feedback {
		margin: 0;
	}

	/*
	 * Los botones a la izquierda y en el orden en que se leen, no anclados al
	 * borde derecho: el ojo viene bajando por los campos y ahí es donde termina.
	 * Y «Cancelar» es un enlace de verdad al listado, no un botón con `goto`.
	 */
	.actions {
		display: flex;
		align-items: center;
		gap: 1.25rem;
		margin-top: 0.5rem;
	}

	/* El botón sin el halo ámbar que `ui/button` le pone, como en configuración:
	   en una pantalla de un solo formulario el halo no jerarquiza nada. */
	form :global(.btn-primary) {
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

	@media (max-width: 900px) {
		.fields {
			max-width: none;
		}
	}
</style>
