import { notifications } from '$lib/stores/notification';
import type { DockerError } from '$lib/types/docker';

export async function handleApiError(error: unknown): Promise<DockerError> {
  if (error instanceof Response) {
    try {
      const data = await error.json();
      return {
        message: data.message || 'An error occurred',
        code: `API_${error.status}`,
        timestamp: new Date()
      };
    } catch {
      return {
        message: error.statusText,
        code: `API_${error.status}`,
        timestamp: new Date()
      };
    }
  }
  
  return {
    message: error instanceof Error ? error.message : 'Unknown error',
    code: 'UNKNOWN_ERROR',
    timestamp: new Date()
  };
}

export function createErrorHandler(component: string) {
  return async (error: unknown) => {
    const { message, code } = await handleApiError(error);
    notifications.error(`${component}: ${message}`, {
      title: `Error in ${component}`,
      priority: 'high',
      autoDismiss: false
    });
    
    console.error(`[${component}] ${code}:`, error);
  };
}

export function handleWebSocketError(error: unknown, context: string) {
  const errorData = {
    message: error instanceof Error ? error.message : 'WebSocket connection error',
    code: `WS_${context.toUpperCase()}_ERROR`,
    timestamp: new Date()
  };
  
  notifications.error(errorData.message, {
    title: 'WebSocket Error',
    priority: 'medium'
  });
  
  return errorData;
} 