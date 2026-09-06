<script lang="ts">
	/** El canal de avisos dentro del panel: anunciado, todavía no disponible. */
	import NotificationSection from './notification-section.svelte';
</script>

<NotificationSection
	title="En la app"
	description="Avisos dentro del panel, sin pasar por la bandeja de entrada."
>
	<!--
		Una frase, no una tarjeta con su icono y su píldora. Lo que no existe
		todavía no necesita la mitad de la pantalla para decirlo, y quien viene a
		esta página viene a decidir sobre lo que sí puede recibir hoy.
	-->
	<p class="pending">
		Todavía no están disponibles. Cuando lleguen te enterarás aquí mismo de los cambios de precio y
		los movimientos de tus portafolios; hasta entonces, todo lo que Finexia te avisa sale por
		correo.
	</p>
</NotificationSection>

<style>
	.pending {
		max-width: 58ch;
		margin: 0;
		font-size: 0.87rem;
		line-height: 1.6;
		color: var(--text-muted);
	}
</style>
