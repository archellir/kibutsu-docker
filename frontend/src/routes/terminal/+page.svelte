<script lang="ts">
	import { notifications } from '$lib/stores/notification';
	import Terminal from '$lib/components/terminal/Terminal.svelte';
	import ErrorBoundary from '$lib/components/notifications/ErrorBoundary.svelte';
	import { theme } from '$lib/stores/theme';

	interface TerminalError {
		type: 'connection' | 'command' | 'system';
		message: string;
		recoverable: boolean;
	}

	interface TerminalState {
		connected: boolean;
		fontSize: number;
		theme: 'light' | 'dark';
		commandHistory: string[];
	}

	let terminalState: TerminalState = {
		connected: false,
		fontSize: 14,
		theme: $theme,
		commandHistory: []
	};

	let lastError: TerminalError | null = null;

	function handleConnect() {
		terminalState.connected = true;
		notifications.success('Terminal connected', {
			title: 'Connection Established',
			autoDismiss: true
		});
	}

	function handleDisconnect() {
		terminalState.connected = false;
		notifications.warning('Terminal disconnected', {
			title: 'Connection Lost',
			autoDismiss: true
		});
	}

	function handleCommand(event: CustomEvent<{ command: string }>) {
		terminalState.commandHistory = [event.detail.command, ...terminalState.commandHistory];
	}

	function handleError(
		event: CustomEvent<{ type: string; message: string; recoverable: boolean }>
	) {
		lastError = {
			type: event.detail.type as 'connection' | 'command' | 'system',
			message: event.detail.message,
			recoverable: event.detail.recoverable
		};
		notifications.error(event.detail.message, {
			title: `Terminal ${event.detail.type} Error`,
			priority: 'high',
			autoDismiss: !event.detail.recoverable
		});
	}

	function resetConnection() {
		lastError = null;
		notifications.info('Resetting terminal connection...', {
			autoDismiss: true
		});
	}

	function copyToClipboard() {
		const selection = window.getSelection()?.toString();
		if (selection) {
			navigator.clipboard.writeText(selection);
			notifications.success('Copied to clipboard', { autoDismiss: true });
		}
	}
</script>

<div class="flex h-full flex-col space-y-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">Terminal</h1>
		<div class="flex items-center gap-2">
			<button
				class="rounded bg-blue-500 px-3 py-1 text-white hover:bg-blue-600"
				onclick={copyToClipboard}
			>
				Copy
			</button>
			<button
				class="rounded bg-gray-500 px-3 py-1 text-white hover:bg-gray-600"
				onclick={resetConnection}
			>
				Reset
			</button>
		</div>
	</div>

	{#if lastError && lastError.recoverable}
		<div class="rounded-lg border border-red-400 bg-red-100 p-4 dark:bg-red-900/20">
			<h3 class="mb-2 font-semibold">{lastError.message}</h3>
			<button
				class="rounded bg-red-500 px-3 py-1 text-white hover:bg-red-600"
				onclick={resetConnection}
			>
				Try Again
			</button>
		</div>
	{/if}

	<ErrorBoundary>
		<Terminal
			on:connect={handleConnect}
			on:disconnect={handleDisconnect}
			on:command={handleCommand}
			on:error={handleError}
		/>
	</ErrorBoundary>

	{#if terminalState.commandHistory.length > 0}
		<div class="mt-4">
			<h3 class="mb-2 text-sm font-semibold">Command History</h3>
			<div class="max-h-32 overflow-y-auto rounded border p-2 dark:border-gray-700">
				{#each terminalState.commandHistory as command}
					<div class="font-mono text-sm">{command}</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
