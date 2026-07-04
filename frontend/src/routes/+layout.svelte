<script lang="ts">
	import './layout.css';
	import CookieNotice from '$/components/cookie-notice.svelte';

	let { children } = $props();
</script>

{@render children()}
<CookieNotice />
