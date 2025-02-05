<script lang="ts">
	import '../app.css';
	import Navbar from '$lib/components/layout/Navbar.svelte';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';

	let { children } = $props();
</script>

<div class="min-h-screen dark:bg-gray-900">
	<Navbar />
	<div class="flex">
		<Sidebar />
		<main class="flex-1 p-8">
			{@render children()}
		</main>
	</div>
</div>
