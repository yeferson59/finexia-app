<script lang="ts">
	/** Envoltorio de pruebas: `Modal` recibe su contenido por snippet. */
	import Modal from './modal.svelte';

	let {
		open = false,
		title = 'Título',
		description = '',
		onClose = () => {}
	}: {
		open?: boolean;
		title?: string;
		description?: string;
		onClose?: () => void;
	} = $props();
</script>

<Modal {open} {title} {description} {onClose}>
	<p>contenido del modal</p>
</Modal>
