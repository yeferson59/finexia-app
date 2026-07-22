<script lang="ts">
	import './layout.css';
	import { CookieNotice } from '$lib/features/legal';

	let { children } = $props();
</script>

{@render children()}
<CookieNotice />
