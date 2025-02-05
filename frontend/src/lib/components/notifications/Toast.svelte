<script lang="ts">
	import { notifications } from '$lib/stores/notification';
	import { fly } from 'svelte/transition';
	import type { NotificationType } from '$lib/stores/notification';

	export let notification: {
		id: string;
		type: NotificationType;
		title: string;
		message: string;
		actions?: Array<{ label: string; onClick: () => void; style?: string }>;
	};

	const icons: Record<NotificationType, string> = {
		success: '✅',
		error: '❌',
		warning: '⚠️',
		info: 'ℹ️'
	};

	const colors = {
		success: 'bg-green-100 dark:bg-green-800 border-green-400',
		error: 'bg-red-100 dark:bg-red-800 border-red-400',
		warning: 'bg-yellow-100 dark:bg-yellow-800 border-yellow-400',
		info: 'bg-blue-100 dark:bg-blue-800 border-blue-400'
	};
</script>

<div
	role="alert"
	class={`flex items-center justify-between rounded-lg border p-4 shadow-lg ${colors[notification.type]}`}
	transition:fly={{ y: 50, duration: 300 }}
>
	<div class="flex items-center gap-3">
		<span class="text-xl" aria-hidden="true">{icons[notification.type]}</span>
		<div>
			<h3 class="font-semibold">{notification.title}</h3>
			<p class="text-sm">{notification.message}</p>
		</div>
	</div>

	{#if notification.actions}
		<div class="flex gap-2">
			{#each notification.actions as action}
				<button
					class="rounded-md px-3 py-1 text-sm hover:bg-opacity-90"
					class:bg-blue-500={action.style === 'primary'}
					class:bg-gray-500={action.style === 'secondary'}
					class:bg-red-500={action.style === 'danger'}
					class:text-white={action.style}
					on:click={action.onClick}
				>
					{action.label}
				</button>
			{/each}
		</div>
	{/if}

	<button
		class="ml-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
		on:click={() => notifications.dismiss(notification.id)}
		aria-label="Dismiss notification"
	>
		✕
	</button>
</div>
