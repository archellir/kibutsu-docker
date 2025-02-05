<script lang="ts">
	import { notifications } from '$lib/stores/notification';
	import { slide } from 'svelte/transition';

	export let message: string;
	export let progress: number | undefined = undefined;
	export let type: 'loading' | 'success' | 'error' = 'loading';

	const colors = {
		loading: 'bg-blue-500',
		success: 'bg-green-500',
		error: 'bg-red-500'
	};
</script>

<div
	role="status"
	class="fixed bottom-0 left-0 right-0 border-t bg-white p-4 dark:border-gray-700 dark:bg-gray-800"
	transition:slide
>
	<div class="container mx-auto flex items-center justify-between">
		<div class="flex items-center gap-3">
			{#if type === 'loading'}
				<div
					class="h-4 w-4 animate-spin rounded-full border-2 border-blue-500 border-t-transparent"
				></div>
			{/if}
			<span>{message}</span>
		</div>

		{#if progress !== undefined}
			<div class="h-2 w-48 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
				<div
					class={`h-full ${colors[type]} transition-all duration-300`}
					style={`width: ${progress}%`}
				></div>
			</div>
		{/if}
	</div>
</div>
