import { writable, derived } from 'svelte/store';

export type NotificationType = 'success' | 'error' | 'warning' | 'info';
export type NotificationPriority = 'low' | 'medium' | 'high';

export interface NotificationAction {
  label: string;
  onClick: () => void;
  style?: 'primary' | 'secondary' | 'danger';
}

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  priority: NotificationPriority;
  actions?: NotificationAction[];
  autoDismiss?: boolean;
  dismissAfter?: number; // in milliseconds
  timestamp: Date;
}

interface NotificationState {
  notifications: Notification[];
  maxNotifications: number;
}

const DEFAULT_DISMISS_AFTER = 5000; // 5 seconds
const MAX_NOTIFICATIONS = 5;

function createNotificationStore() {
  const { subscribe, update } = writable<NotificationState>({
    notifications: [],
    maxNotifications: MAX_NOTIFICATIONS
  });

  const dismissTimers = new Map<string, NodeJS.Timeout>();

  function addNotification(
    notification: Omit<Notification, 'id' | 'timestamp'> & { id?: string; timestamp?: Date }
  ) {
    const id = notification.id || crypto.randomUUID();
    const timestamp = notification.timestamp || new Date();

    update(state => {
      // Sort notifications by priority and timestamp
      const newNotification = { ...notification, id, timestamp } as Notification;
      let notifications = [...state.notifications, newNotification].sort((a, b) => {
        const priorityOrder = { high: 0, medium: 1, low: 2 };
        const priorityDiff = priorityOrder[a.priority] - priorityOrder[b.priority];
        return priorityDiff === 0 ? b.timestamp.getTime() - a.timestamp.getTime() : priorityDiff;
      });

      // Limit the number of notifications
      if (notifications.length > state.maxNotifications) {
        // Remove oldest low priority notifications first
        notifications = notifications.slice(0, state.maxNotifications);
      }

      return { ...state, notifications };
    });

    // Set up auto-dismiss
    if (notification.autoDismiss !== false) {
      const timer = setTimeout(
        () => dismiss(id),
        notification.dismissAfter || DEFAULT_DISMISS_AFTER
      );
      dismissTimers.set(id, timer);
    }

    return id;
  }

  function dismiss(id: string) {
    // Clear any existing dismiss timer
    const timer = dismissTimers.get(id);
    if (timer) {
      clearTimeout(timer);
      dismissTimers.delete(id);
    }

    update(state => ({
      ...state,
      notifications: state.notifications.filter(n => n.id !== id)
    }));
  }

  function dismissAll() {
    // Clear all dismiss timers
    dismissTimers.forEach(timer => clearTimeout(timer));
    dismissTimers.clear();

    update(state => ({ ...state, notifications: [] }));
  }

  return {
    subscribe,
    add: addNotification,
    dismiss,
    dismissAll,
    success: (message: string, options?: Partial<Omit<Notification, 'type' | 'message' | 'title'>> & { title?: string }) =>
      addNotification({ type: 'success', message, title: message, priority: 'medium', ...options }),
    error: (message: string, options?: Partial<Omit<Notification, 'type' | 'message' | 'title'>> & { title?: string }) =>
      addNotification({ type: 'error', message, title: message, priority: 'high', ...options }),
    warning: (message: string, options?: Partial<Omit<Notification, 'type' | 'message' | 'title'>> & { title?: string }) =>
      addNotification({ type: 'warning', message, title: message, priority: 'medium', ...options }),
    info: (message: string, options?: Partial<Omit<Notification, 'type' | 'message' | 'title'>> & { title?: string }) =>
      addNotification({ type: 'info', message, title: message, priority: 'low', ...options })
  };
}

// Create and export the notification store
export const notifications = createNotificationStore();

// Derived store for active notifications
export const activeNotifications = derived(
  notifications,
  $notifications => $notifications.notifications
); 