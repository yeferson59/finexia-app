<script lang="ts">
	interface Props {
		iconSize?: number;
		nameSize?: number;
	}

	let { iconSize = 30, nameSize = 20 }: Props = $props();
</script>

<div class="brand">
	<svg class="brand-icon" width={iconSize} height={iconSize} viewBox="0 0 30 30" fill="none">
		<rect width="30" height="30" rx="7" fill="var(--amber)" />
		<path
			d="M7 22L12.5 14.5L16.5 18.5L23 9"
			stroke="#0c0a06"
			stroke-width="2.6"
			stroke-linecap="round"
			stroke-linejoin="round"
		/>
	</svg>
	<span class="brand-name" style="font-size:{nameSize}px">FINEXIA</span>
</div>

<style>
	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.brand-icon {
		flex-shrink: 0;
	}
	.brand-name {
		font-family: var(--font-display);
		font-weight: 600;
		letter-spacing: 0.1em;
		color: var(--text);
	}
</style>
