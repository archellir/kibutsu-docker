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

<div class="flex h-screen bg-gray-100 dark:bg-gray-900">
	<Sidebar />

	<div class="flex flex-1 flex-col overflow-hidden">
		<Navbar />

		<main class="flex-1 overflow-y-auto p-4">
			{@render children()}
		</main>
	</div>
</div>
