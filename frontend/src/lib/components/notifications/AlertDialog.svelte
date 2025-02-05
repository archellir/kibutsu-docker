<script lang="ts">
	let {
		dispatch,
		title,
		message,
		type = 'error',
		confirmLabel = 'Confirm',
		cancelLabel = 'Cancel'
	} = $props<{
		title: string;
		message: string;
		type?: 'error' | 'warning';
		confirmLabel?: string;
		cancelLabel?: string;
	}>();

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			dispatch('cancel');
		}
	}
</script>

<div
	role="presentation"
	class="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 p-4"
	onkeydown={handleKeydown}
>
	<div
		role="dialog"
		aria-modal="true"
		aria-labelledby="dialog-title"
		class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl dark:bg-gray-800"
	>
		<h2 id="dialog-title" class="mb-4 text-xl font-bold">
			<span aria-hidden="true" class="mr-2">
				{type === 'error' ? '❌' : '⚠️'}
			</span>
			{title}
		</h2>
		<p class="mb-6 text-gray-600 dark:text-gray-300">{message}</p>
		<div class="flex justify-end gap-3">
			<button
				class="rounded-md bg-gray-200 px-4 py-2 hover:bg-opacity-90 dark:bg-gray-700"
				onclick={() => dispatch('cancel')}
			>
				{cancelLabel}
			</button>
			<button
				class="rounded-md px-4 py-2 text-white hover:bg-opacity-90"
				class:bg-red-500={type === 'error'}
				class:bg-yellow-500={type === 'warning'}
				onclick={() => dispatch('confirm')}
				autofocus
			>
				{confirmLabel}
			</button>
		</div>
	</div>
</div>
