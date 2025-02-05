<script lang="ts">
	import { activeNotifications } from '$lib/stores/notification';
	import Toast from './Toast.svelte';
</script>

<div aria-live="polite" class="fixed right-4 top-4 z-50 flex w-full max-w-md flex-col gap-2">
	{#each $activeNotifications as notification (notification.id)}
		<Toast {notification} />
	{/each}
</div>
