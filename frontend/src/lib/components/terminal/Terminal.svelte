<script lang="ts">
	import { onMount, onDestroy, createEventDispatcher } from 'svelte';
	import { Terminal as Xterm } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import { WebglAddon } from '@xterm/addon-webgl';
	import { notifications } from '$lib/stores/notification';
	import { theme } from '$lib/stores/theme';
	import ErrorBoundary from '../notifications/ErrorBoundary.svelte';

	// Props
	let {
		websocketUrl = 'ws://localhost:3000/terminal',
		fontSize = 14,
		connect,
		disconnect,
		command,
		commandError,
		resize,
		exit,
		error
	} = $props<{
		websocketUrl?: string;
		fontSize?: number;
		connect?: () => void;
		disconnect?: () => void;
		command?: (event: { command: string }) => void;
		commandError?: (event: { command: string; error: string }) => void;
		resize?: (event: { cols: number; rows: number }) => void;
		exit?: (event: { code: number }) => void;
		error?: (event: {
			type: 'connection' | 'command' | 'system';
			message: string;
			recoverable: boolean;
		}) => void;
	}>();

	// Component state
	let terminal: Xterm;
	let fitAddon: FitAddon;
	let websocket: WebSocket;
	let terminalElement: HTMLDivElement;
	let connectionStatus = $state<'connecting' | 'connected' | 'disconnected'>('disconnected');
	let reconnectAttempts = $state(0);
	const MAX_RECONNECT_ATTEMPTS = 5;
	let commandHistory: string[] = [];
	let historyIndex = -1;
	let currentCommand = '';

	// Terminal themes
	const themes = {
		light: {
			background: '#ffffff',
			foreground: '#2e3440',
			cursor: '#434c5e'
		},
		dark: {
			background: '#2e3440',
			foreground: '#d8dee9',
			cursor: '#88c0d0'
		}
	};

	const dispatch = createEventDispatcher<{
		connect: void;
		disconnect: void;
		command: { command: string };
		error: { type: string; message: string; recoverable: boolean };
	}>();

	onMount(() => {
		try {
			initializeTerminal();
			connectWebSocket();
			setupResizeHandler();
		} catch (error) {
			handleError('Terminal initialization failed', error);
		}
	});

	onDestroy(() => {
		cleanup();
	});

	function initializeTerminal() {
		terminal = new Xterm({
			fontSize,
			fontFamily: 'Menlo, Monaco, "Courier New", monospace',
			theme: themes[$theme],
			cursorBlink: true,
			scrollback: 1000
		});

		fitAddon = new FitAddon();
		terminal.loadAddon(fitAddon);
		terminal.loadAddon(new WebLinksAddon());

		try {
			terminal.loadAddon(new WebglAddon());
		} catch (error) {
			notifications.warning('WebGL not available, falling back to canvas renderer');
		}

		terminal.onData((data) => {
			if (websocket?.readyState === WebSocket.OPEN) {
				websocket.send(data);
			}
			command?.({ command: data });
			dispatch('command', { command: data });
		});

		terminal.onKey(handleTerminalKey);

		terminal.open(terminalElement);
		fitAddon.fit();
	}

	function connectWebSocket() {
		connectionStatus = 'connecting';
		websocket = new WebSocket(websocketUrl);

		websocket.onopen = () => {
			connectionStatus = 'connected';
			reconnectAttempts = 0;
			notifications.success('Terminal connected');
			connect?.();
			dispatch('connect');
		};

		websocket.onclose = () => {
			connectionStatus = 'disconnected';
			handleReconnect();
			disconnect?.();
			dispatch('disconnect');
		};

		websocket.onerror = (error) => {
			handleError('WebSocket error', error);
		};

		websocket.onmessage = (event) => {
			try {
				terminal.write(event.data);
			} catch (error) {
				handleError('Failed to write to terminal', error);
			}
		};
	}

	function handleReconnect() {
		if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
			notifications.error('Maximum reconnection attempts reached', {
				title: 'Connection Failed',
				autoDismiss: false
			});
			return;
		}

		reconnectAttempts++;
		const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 10000);

		notifications.info(
			`Reconnecting in ${delay / 1000}s... (Attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`
		);

		setTimeout(connectWebSocket, delay);
	}

	function handleError(context: string, error: unknown) {
		notifications.error(`${context}: ${error instanceof Error ? error.message : 'Unknown error'}`, {
			title: 'Terminal Error',
			priority: 'high'
		});
		commandError?.({
			command: currentCommand,
			error: error instanceof Error ? error.message : 'Unknown error'
		});
	}

	function cleanup() {
		websocket?.close();
		terminal?.dispose();
		window.removeEventListener('resize', handleResize);
		exit?.({ code: 0 });
	}

	function handleTerminalData(data: string) {
		if (websocket?.readyState === WebSocket.OPEN) {
			websocket.send(data);
		}
		command?.({ command: data });
	}

	function handleTerminalKey(event: { key: string; domEvent: KeyboardEvent }) {
		const { domEvent } = event;

		if (domEvent.key === 'ArrowUp') {
			if (historyIndex < commandHistory.length - 1) {
				historyIndex++;
				currentCommand = commandHistory[historyIndex];
				terminal.write('\x1b[2K\r$ ' + currentCommand);
			}
		} else if (domEvent.key === 'ArrowDown') {
			if (historyIndex > 0) {
				historyIndex--;
				currentCommand = commandHistory[historyIndex];
				terminal.write('\x1b[2K\r$ ' + currentCommand);
			}
		}
	}

	function handleResize() {
		try {
			fitAddon?.fit();
			resize?.({ cols: terminal.cols, rows: terminal.rows });
		} catch (error) {
			handleError('Resize failed', error);
		}
	}

	function setupResizeHandler() {
		window.addEventListener('resize', handleResize);
		handleResize();
	}

	// Toolbar actions
	function clearScreen() {
		terminal.clear();
	}

	function changeFontSize(delta: number) {
		fontSize = Math.max(8, Math.min(24, fontSize + delta));
		terminal.options.fontSize = fontSize;
		fitAddon.fit();
	}

	$effect(() => {
		if (terminal && $theme) {
			terminal.options.theme = themes[$theme];
		}
	});
</script>

<ErrorBoundary>
	<div class="flex h-full flex-col rounded-lg border dark:border-gray-700">
		<div class="flex items-center justify-between border-b p-2 dark:border-gray-700">
			<div class="flex items-center gap-2">
				<div
					class={`h-2 w-2 rounded-full ${
						connectionStatus === 'connected'
							? 'bg-green-500'
							: connectionStatus === 'connecting'
								? 'bg-yellow-500'
								: 'bg-red-500'
					}`}
				></div>
				<span class="text-sm">{connectionStatus}</span>
			</div>

			<div class="flex items-center gap-2">
				<button
					class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
					onclick={() => {
						cleanup();
						connectWebSocket();
					}}
					title="Reset Connection"
				>
					🔄
				</button>
				<button
					class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
					onclick={() => changeFontSize(-1)}
					title="Decrease Font Size"
				>
					A-
				</button>
				<span class="text-sm">{fontSize}px</span>
				<button
					class="rounded p-1 hover:bg-gray-100 dark:hover:bg-gray-700"
					onclick={() => changeFontSize(1)}
					title="Increase Font Size"
				>
					A+
				</button>
			</div>
		</div>

		<div
			bind:this={terminalElement}
			class="relative flex-1 overflow-hidden"
			onpaste={(e) => {
				e.preventDefault();
				const text = e.clipboardData?.getData('text');
				if (text && websocket?.readyState === WebSocket.OPEN) {
					websocket.send(text);
				}
			}}
		></div>
	</div>
</ErrorBoundary>

<style>
	:global(.xterm) {
		height: 100%;
		padding: 8px;
	}
</style>
