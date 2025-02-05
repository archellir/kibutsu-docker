<script lang="ts">
	import '../app.css';
	import '@xterm/xterm/css/xterm.css';
	import Navbar from '$lib/components/layout/Navbar.svelte';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import NotificationContainer from '$lib/components/notifications/NotificationContainer.svelte';
	import { theme } from '$lib/stores/theme';
	import { onMount } from 'svelte';

	let { children } = $props();

	onMount(() => {
		// Apply theme class to HTML element
		$: if ($theme === 'dark') {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	});
</script>

<div class="min-h-screen">
	<NotificationContainer />
	<Navbar />
	<div class="flex">
		<Sidebar />
		<main class="flex-1 p-8">
			{@render children()}
		</main>
	</div>
</div>
