import { writable, derived } from 'svelte/store';
import type { Container, Image, ComposeProject, SystemInfo, DockerError } from '../types/docker';

// Utility function for WebSocket connection
const createWebSocketConnection = () => {
  const ws = new WebSocket('ws://localhost:3000/docker');
  
  ws.onclose = () => {
    setTimeout(createWebSocketConnection, 5000); // Reconnect after 5s
  };
  
  return ws;
};

// Create base stores with loading states
const createLoadingStore = <T>() => {
  const { subscribe, set: baseSet, update } = writable<{
    data: T;
    loading: boolean;
    lastUpdated: Date | null;
  }>({
    data: [] as unknown as T,
    loading: false,
    lastUpdated: null
  });

  return {
    subscribe,
    set: (data: T) => baseSet({
      data,
      loading: false,
      lastUpdated: new Date()
    }),
    setLoading: (loading: boolean) => update(store => ({ ...store, loading })),
    refresh: async (fetchFn: () => Promise<T>) => {
      update(store => ({ ...store, loading: true }));
      try {
        const data = await fetchFn();
        baseSet({
          data,
          loading: false,
          lastUpdated: new Date()
        });
      } catch (error: unknown) {
        errorStore.add({
          message: error instanceof Error ? error.message : 'Unknown error',
          code: 'FETCH_ERROR',
          timestamp: new Date()
        });
      }
    }
  };
};

// Create stores
export const containersStore = createLoadingStore<Container[]>();
export const imagesStore = createLoadingStore<Image[]>();
export const composeStore = createLoadingStore<ComposeProject[]>();
export const systemStore = createLoadingStore<SystemInfo>();

// Error store with history
export const errorStore = (() => {
  const { subscribe, update } = writable<DockerError[]>([]);

  return {
    subscribe,
    add: (error: DockerError) => update(errors => [error, ...errors].slice(0, 10)),
    clear: () => update(() => [])
  };
})();

// Auto-refresh functionality
const setupAutoRefresh = () => {
  const refreshInterval = 30000; // 30 seconds

  const refresh = async () => {
    await Promise.all([
      containersStore.refresh(fetchContainers),
      imagesStore.refresh(fetchImages),
      composeStore.refresh(fetchComposeProjects),
      systemStore.refresh(fetchSystemInfo)
    ]);
  };

  // Initial load
  refresh();

  // Set up interval
  const interval = setInterval(refresh, refreshInterval);

  // Set up WebSocket listeners
  const ws = createWebSocketConnection();
  
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    switch (data.type) {
      case 'container':
        containersStore.refresh(fetchContainers);
        break;
      case 'image':
        imagesStore.refresh(fetchImages);
        break;
      case 'compose':
        composeStore.refresh(fetchComposeProjects);
        break;
      case 'system':
        systemStore.refresh(fetchSystemInfo);
        break;
    }
  };

  // Cleanup function
  return () => {
    clearInterval(interval);
    ws.close();
  };
};

// Example fetch functions (implement these based on your API)
const fetchContainers = async (): Promise<Container[]> => {
  const response = await fetch('/api/containers');
  return response.json();
};

const fetchImages = async (): Promise<Image[]> => {
  const response = await fetch('/api/images');
  return response.json();
};

const fetchComposeProjects = async (): Promise<ComposeProject[]> => {
  const response = await fetch('/api/compose');
  return response.json();
};

const fetchSystemInfo = async (): Promise<SystemInfo> => {
  const response = await fetch('/api/system');
  return response.json();
};

// Initialize auto-refresh
if (typeof window !== 'undefined') {
  setupAutoRefresh();
} 