import { ref, reactive } from 'vue'
import type { ToastProps } from '@/components/ui/AppToast.vue'

export interface Notification extends ToastProps {
  id: string
  timestamp: number
}

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface NotificationOptions {
  title?: string
  duration?: number
  dismissible?: boolean
  persistent?: boolean
}

// Global notification state
const notificationList = ref<Notification[]>([])
let notificationIdCounter = 0

export function useNotifications() {
  
  // Generate unique ID for notifications
  const generateId = (): string => {
    return `notification_${++notificationIdCounter}_${Date.now()}`
  }

  // Add a notification
  const addNotification = (
    type: NotificationType,
    message: string,
    options: NotificationOptions = {}
  ): string => {
    const id = generateId()
    
    const notification: Notification = {
      id,
      type,
      message,
      title: options.title,
      duration: options.duration ?? (type === 'error' ? 7000 : 5000),
      dismissible: options.dismissible ?? true,
      persistent: options.persistent ?? false,
      timestamp: Date.now()
    }

    notificationList.value.push(notification)
    
    // Auto-remove after max time if not persistent
    if (!notification.persistent) {
      const maxDuration = Math.max(notification.duration || 5000, 10000)
      setTimeout(() => {
        removeNotification(id)
      }, maxDuration)
    }

    return id
  }

  // Remove a notification
  const removeNotification = (id: string) => {
    const index = notificationList.value.findIndex(n => n.id === id)
    if (index > -1) {
      notificationList.value.splice(index, 1)
    }
  }

  // Clear all notifications
  const clearAll = () => {
    notificationList.value = []
  }

  // Convenience methods for different notification types
  const success = (message: string, options?: NotificationOptions) => {
    return addNotification('success', message, options)
  }

  const error = (message: string, options?: NotificationOptions) => {
    return addNotification('error', message, {
      duration: 7000,
      ...options
    })
  }

  const warning = (message: string, options?: NotificationOptions) => {
    return addNotification('warning', message, options)
  }

  const info = (message: string, options?: NotificationOptions) => {
    return addNotification('info', message, options)
  }

  // Convenience methods for common bookmark operations
  const bookmarkCreated = (title?: string) => {
    return success(
      title ? `"${title}" has been saved` : 'Bookmark saved successfully',
      { title: 'Bookmark Created' }
    )
  }

  const bookmarkUpdated = (title?: string) => {
    return success(
      title ? `"${title}" has been updated` : 'Bookmark updated successfully',
      { title: 'Bookmark Updated' }
    )
  }

  const bookmarkDeleted = (title?: string) => {
    return success(
      title ? `"${title}" has been deleted` : 'Bookmark deleted successfully',
      { title: 'Bookmark Deleted' }
    )
  }

  const bookmarkMoved = (action: string, title?: string) => {
    const actionLabels: Record<string, string> = {
      'working': 'moved to working',
      'share': 'marked for sharing',
      'archived': 'archived',
      'read-later': 'moved to triage'
    }
    
    const actionLabel = actionLabels[action] || `action changed to ${action}`
    
    return success(
      title ? `"${title}" has been ${actionLabel}` : `Bookmark ${actionLabel}`,
      { title: 'Bookmark Updated' }
    )
  }

  const bulkOperation = (count: number, operation: string) => {
    return success(
      `${count} bookmark${count === 1 ? '' : 's'} ${operation}`,
      { title: 'Bulk Operation Complete' }
    )
  }

  // Error handling for API failures
  const apiError = (operation: string, error?: Error | string) => {
    const errorMessage = typeof error === 'string' ? error : error?.message || 'Unknown error occurred'
    
    return addNotification('error', `Failed to ${operation}: ${errorMessage}`, {
      title: 'Operation Failed',
      duration: 8000,
      dismissible: true
    })
  }

  const networkError = () => {
    return error(
      'Unable to connect to the server. Please check your internet connection and try again.',
      { 
        title: 'Network Error',
        duration: 10000 
      }
    )
  }

  const validationError = (message: string) => {
    return warning(message, {
      title: 'Validation Error',
      duration: 6000
    })
  }

  // Project-specific notifications
  const projectCreated = (name: string) => {
    return success(`Project "${name}" has been created`, {
      title: 'Project Created'
    })
  }

  const projectUpdated = (name: string) => {
    return success(`Project "${name}" has been updated`, {
      title: 'Project Updated'
    })
  }

  const projectDeleted = (name: string) => {
    return success(`Project "${name}" has been deleted`, {
      title: 'Project Deleted'
    })
  }

  return {
    // State
    notifications: notificationList,
    
    // Core methods
    addNotification,
    removeNotification,
    clearAll,
    
    // Type-specific methods
    success,
    error,
    warning,
    info,
    
    // Bookmark operations
    bookmarkCreated,
    bookmarkUpdated,
    bookmarkDeleted,
    bookmarkMoved,
    bulkOperation,
    
    // Error handling
    apiError,
    networkError,
    validationError,
    
    // Project operations
    projectCreated,
    projectUpdated,
    projectDeleted
  }
}

// Export a global instance for use across the app
export const notifications = useNotifications()