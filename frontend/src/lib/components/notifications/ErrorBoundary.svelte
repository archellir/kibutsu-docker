<script lang="ts">
	import { notifications } from '$lib/stores/notification';

	let error: Error | null = null;

	function handleError(e: unknown, reset: () => void) {
		error = e as Error;
		notifications.error((e as Error).message, {
			title: 'Rendering Error',
			priority: 'high',
			autoDismiss: false
		});
	}

	function resetError() {
		error = null;
	}
</script>

<svelte:boundary onerror={handleError}>
	{#if error}
		<div
			role="alert"
			class="rounded-lg border border-red-400 bg-red-100 p-6 dark:bg-red-900 dark:bg-opacity-20"
		>
			<h2 class="mb-2 text-xl font-bold">Component Error</h2>
			<p class="mb-4 font-mono text-sm">{error.message}</p>
			<button
				class="rounded-md bg-red-500 px-4 py-2 text-white hover:bg-red-600"
				on:click={resetError}
			>
				Try Again
			</button>
		</div>
	{:else}
		<slot />
	{/if}
</svelte:boundary>
