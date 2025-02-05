import { get } from 'svelte/store';
import type { Container, Image, ComposeProject, SystemInfo, DockerError } from '../types/docker';
import { containersStore, imagesStore, composeStore, systemStore, errorStore } from '../stores/docker';
import { wsManager } from '../websocket/manager';

// Types for update tracking
type UpdateType = 'container' | 'image' | 'compose' | 'system';
type UpdateTracker = Map<string, { timestamp: Date; rollback: () => void }>;

// Update trackers for rollback support
const updateHistory: Record<UpdateType, UpdateTracker> = {
  container: new Map(),
  image: new Map(),
  compose: new Map(),
  system: new Map()
};

// Container status updates
export async function handleContainerUpdate(containerId: string, newState: Partial<Container>) {
  const containers = get(containersStore).data;
  const index = containers.findIndex(c => c.Id === containerId);
  
  if (index === -1) return;

  const oldState = containers[index];
  const updatedContainer = { ...oldState, ...newState };
  
  // Store rollback data
  updateHistory.container.set(containerId, {
    timestamp: new Date(),
    rollback: () => {
      const currentContainers = [...get(containersStore).data];
      currentContainers[index] = oldState;
      containersStore.set(currentContainers);
    }
  });

  // Optimistic update
  const updatedContainers = [...containers];
  updatedContainers[index] = updatedContainer;
  containersStore.set(updatedContainers);

  try {
    // Subscribe to container stats if running
    if (newState.State === 'running') {
      wsManager.connectToContainerStats(containerId);
    }
  } catch (error) {
    handleError('CONTAINER_UPDATE_ERROR', error);
    rollbackUpdate('container', containerId);
  }
}

// Log streaming management
export function setupLogStreaming(containerId: string) {
  return wsManager.subscribe(`/containers/${containerId}/logs`, (data) => {
    // Handle incoming log data
    const logUpdate = new CustomEvent('docker:log', {
      detail: { containerId, log: data }
    });
    window.dispatchEvent(logUpdate);
  });
}

// Stats updates handling
export function handleStatsUpdate(containerId: string, stats: any) {
  const statsUpdate = new CustomEvent('docker:stats', {
    detail: { containerId, stats }
  });
  window.dispatchEvent(statsUpdate);
}

// System-wide updates
export function handleSystemUpdate(update: Partial<SystemInfo>) {
  const currentState = get(systemStore).data;
  const oldState = { ...currentState };
  
  updateHistory.system.set('current', {
    timestamp: new Date(),
    rollback: () => systemStore.set(oldState)
  });

  systemStore.set({ ...currentState, ...update });
}

// Error handling with notifications
export function handleError(code: string, error: unknown) {
  const errorData: DockerError = {
    message: error instanceof Error ? error.message : 'Unknown error',
    code,
    timestamp: new Date()
  };

  errorStore.add(errorData);

  // Dispatch error event for UI notifications
  const errorEvent = new CustomEvent('docker:error', {
    detail: errorData
  });
  window.dispatchEvent(errorEvent);
}

// Rollback support
function rollbackUpdate(type: UpdateType, id: string) {
  const update = updateHistory[type].get(id);
  if (update) {
    update.rollback();
    updateHistory[type].delete(id);
  }
}

// Cleanup old update history
function cleanupUpdateHistory() {
  const maxAge = 5 * 60 * 1000; // 5 minutes
  const now = new Date();

  Object.values(updateHistory).forEach(tracker => {
    for (const [id, update] of tracker.entries()) {
      if (now.getTime() - update.timestamp.getTime() > maxAge) {
        tracker.delete(id);
      }
    }
  });
}

// Set up periodic cleanup
setInterval(cleanupUpdateHistory, 60000);

// Export event names as constants
export const DOCKER_EVENTS = {
  LOG: 'docker:log',
  STATS: 'docker:stats',
  ERROR: 'docker:error',
  UPDATE: 'docker:update'
} as const; 