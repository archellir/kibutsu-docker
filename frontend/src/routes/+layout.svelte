<script lang="ts">
	import '../app.css';
	import { theme } from '$lib/stores/theme';

	let { children } = $props();
</script>

<div class="min-h-screen" data-theme={$theme}>
	<header class="bg-white shadow dark:bg-gray-800">
		<nav class="container mx-auto px-4 py-4">
			<button
				class="rounded-lg bg-gray-200 p-2 dark:bg-gray-700"
				onclick={() => theme.set($theme === 'dark' ? 'light' : 'dark')}
			>
				Toggle theme
			</button>
		</nav>
	</header>

	<main class="container mx-auto px-4 py-8">
		{@render children()}
	</main>
</div>
