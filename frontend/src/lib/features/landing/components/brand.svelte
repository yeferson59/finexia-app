<script lang="ts">
	/*
	 * La marca, como la escribe el panel: el cuadro ámbar y «Finexia» en la
	 * serif, en caja normal. Aquí iba en versalitas espaciadas, así que la
	 * portada y el menú del panel firmaban de dos maneras. El color lo hereda,
	 * para servir igual sobre el papel que sobre el verde del pie.
	 */
	interface Props {
		iconSize?: number;
		nameSize?: number;
	}

	let { iconSize = 28, nameSize = 22 }: Props = $props();
</script>

<span class="brand">
	<svg class="brand-icon" width={iconSize} height={iconSize} viewBox="0 0 30 30" fill="none">
		<rect width="30" height="30" rx="7" fill="#d4912a" />
		<path
			d="M7 22L12.5 14.5L16.5 18.5L23 9"
			stroke="#10231e"
			stroke-width="2.6"
			stroke-linecap="round"
			stroke-linejoin="round"
		/>
	</svg>
	<span class="brand-name" style="font-size:{nameSize}px">Finexia</span>
</span>

<style>
	.brand {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		color: inherit;
	}
	.brand-icon {
		flex-shrink: 0;
	}
	.brand-name {
		font-family: var(--font-display);
		font-weight: 500;
		letter-spacing: -0.01em;
		line-height: 1;
	}
</style>
