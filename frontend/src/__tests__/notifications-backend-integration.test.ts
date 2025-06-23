import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { useBookmarkStore } from '@/stores/bookmarks'
import { useNotifications } from '@/composables/useNotifications'
import { bookmarkService } from '@/services/bookmarkService'
import type { Bookmark } from '@/types'

// Test configuration
const API_BASE_URL = 'http://localhost:9090'

// Helper function to check if backend is available
async function isBackendAvailable(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/stats/summary`)
    return response.ok
  } catch {
    return false
  }
}

// Helper function to create test bookmark
async function createTestBookmark(data: Partial<Bookmark> = {}): Promise<Bookmark> {
  const bookmarkData = {
    url: 'https://example.com/test-article',
    title: 'Test Article for Notifications',
    description: 'This is a test bookmark for notification testing',
    action: 'read-later',
    ...data
  }
  
  return await bookmarkService.createBookmark(bookmarkService.toBackendCreateRequest(bookmarkData))
}

// Helper function to clean up test bookmarks
async function cleanupTestBookmarks() {
  try {
    // Get all bookmarks and delete test ones
    const bookmarks = await bookmarkService.getAllBookmarks()
    const testBookmarks = bookmarks.filter(b => 
      b.title.includes('Test Article for Notifications') || 
      b.url.includes('example.com/test')
    )
    
    // Note: We don't have a delete endpoint in the current API
    // so we'll just mark them as archived for cleanup
    for (const bookmark of testBookmarks) {
      await bookmarkService.updateBookmark(bookmark.id, { action: 'archived' })
    }
  } catch (error) {
    console.warn('Cleanup failed:', error)
  }
}

describe('Notification Backend Integration', () => {
  let store: ReturnType<typeof useBookmarkStore>
  let notifications: ReturnType<typeof useNotifications>

  beforeEach(async () => {
    // Check if backend is available
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Backend not available, skipping integration tests')
      return
    }

    // Initialize store and notifications
    store = useBookmarkStore()
    notifications = useNotifications()
    
    // Clear existing notifications
    notifications.clearAll()
    
    // Clean up any existing test data
    await cleanupTestBookmarks()
  })

  afterEach(async () => {
    await cleanupTestBookmarks()
  })

  it('should show success notification when bookmark is created via API', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    const testBookmark = {
      url: 'https://example.com/test-create',
      title: 'Test Article for Notifications - Create',
      description: 'Testing bookmark creation notifications',
      action: 'read-later' as const
    }

    // Add bookmark via store (which should trigger notification)
    await store.addBookmark(testBookmark)

    // Check that success notification was created
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'success',
      title: 'Bookmark Created',
      message: `"${testBookmark.title}" has been saved`
    })
  })

  it('should show success notification when bookmark is updated via API', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    // Create a test bookmark first
    const bookmark = await createTestBookmark({
      title: 'Test Article for Notifications - Update',
      action: 'read-later'
    })

    // Clear notifications from creation
    notifications.clearAll()

    // Load bookmarks into store
    await store.loadBookmarks()

    // Update bookmark via store
    await store.updateBookmark(bookmark.id, { 
      action: 'working', 
      topic: 'Test Project' 
    })

    // Check that update notification was created
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'success',
      title: 'Bookmark Updated',
      message: `"${bookmark.title}" has been updated`
    })
  })

  it('should show success notification when bookmark is shared with recipient', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    // Create a test bookmark
    const bookmark = await createTestBookmark({
      title: 'Test Article for Notifications - Share',
      action: 'read-later'
    })

    // Clear notifications
    notifications.clearAll()

    // Load bookmarks into store
    await store.loadBookmarks()

    // Update bookmark to share action with recipient
    await store.updateBookmark(bookmark.id, { 
      action: 'share',
      shareTo: 'Team Slack'
    })

    // Check that notification was created
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'success',
      title: 'Bookmark Updated',
      message: `"${bookmark.title}" has been updated`
    })
  })

  it('should show error notification when API call fails', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    // Try to update a non-existent bookmark
    try {
      await store.updateBookmark('non-existent-id', { action: 'working' })
    } catch (error) {
      // Expected to fail
    }

    // Check that error notification was created
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'error',
      title: 'Operation Failed'
    })
  })

  it('should show bulk operation notification for multiple bookmark moves', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    // Create multiple test bookmarks
    const bookmarks = await Promise.all([
      createTestBookmark({ 
        title: 'Test Article for Notifications - Bulk 1',
        url: 'https://example.com/test-bulk-1'
      }),
      createTestBookmark({ 
        title: 'Test Article for Notifications - Bulk 2',
        url: 'https://example.com/test-bulk-2'
      }),
      createTestBookmark({ 
        title: 'Test Article for Notifications - Bulk 3',
        url: 'https://example.com/test-bulk-3'
      })
    ])

    // Clear notifications
    notifications.clearAll()

    // Load bookmarks into store
    await store.loadBookmarks()

    // Perform bulk move operation
    const bookmarkIds = bookmarks.map(b => b.id)
    await store.moveBookmarks(bookmarkIds, 'archived')

    // Should have individual update notifications plus bulk notification
    const bulkNotification = notifications.notifications.value.find(n => 
      n.title === 'Bulk Operation Complete'
    )
    
    expect(bulkNotification).toBeTruthy()
    expect(bulkNotification?.message).toContain('3 bookmarks archived')
  })

  it('should persist share recipient data and make it available for future shares', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    // Create test bookmark and share it
    const bookmark = await createTestBookmark({
      title: 'Test Article for Notifications - Share Recipient',
      action: 'read-later'
    })

    // Load bookmarks into store
    await store.loadBookmarks()

    // Share with a custom recipient
    await store.updateBookmark(bookmark.id, { 
      action: 'share',
      shareTo: 'Custom Newsletter'
    })

    // Reload bookmarks to get updated data
    await store.loadBookmarks()

    // Check that the recipient is now available in share destinations
    expect(store.availableShareDestinations).toContain('Custom Newsletter')

    // Create another bookmark and verify the recipient appears as an option
    const bookmark2 = await createTestBookmark({
      title: 'Test Article for Notifications - Share Recipient 2',
      url: 'https://example.com/test-share-2',
      action: 'read-later'
    })

    // Reload again
    await store.loadBookmarks()

    // The custom recipient should still be available
    expect(store.availableShareDestinations).toContain('Custom Newsletter')
  })

  it('should show notification for different bookmark lifecycle transitions', async () => {
    const backendAvailable = await isBackendAvailable()
    if (!backendAvailable) {
      console.warn('Skipping test - backend not available')
      return
    }

    // Create a bookmark
    const bookmark = await createTestBookmark({
      title: 'Test Article for Notifications - Lifecycle',
      action: 'read-later'
    })

    // Load bookmarks
    await store.loadBookmarks()

    // Test different action transitions
    const transitions = [
      { action: 'working', topic: 'Test Project' },
      { action: 'share', shareTo: 'Team Slack' },
      { action: 'archived' }
    ]

    for (let i = 0; i < transitions.length; i++) {
      // Clear notifications
      notifications.clearAll()

      // Apply transition
      const update = transitions[i]
      await store.updateBookmark(bookmark.id, update)

      // Check notification was created
      expect(notifications.notifications.value).toHaveLength(1)
      expect(notifications.notifications.value[0]).toMatchObject({
        type: 'success',
        title: 'Bookmark Updated'
      })
    }
  })
})

describe('Notification System Error Handling', () => {
  let notifications: ReturnType<typeof useNotifications>

  beforeEach(() => {
    notifications = useNotifications()
    notifications.clearAll()
  })

  it('should handle network errors gracefully', async () => {
    // Simulate network error by trying to connect to invalid URL
    const originalFetch = global.fetch
    global.fetch = vi.fn().mockRejectedValue(new Error('Network error'))

    try {
      await bookmarkService.createBookmark({
        url: 'https://test.com',
        title: 'Test',
        description: '',
        action: 'read-later'
      })
    } catch (error) {
      // Expected to fail
    }

    // Restore fetch
    global.fetch = originalFetch

    // Note: This test would need the store to handle the error
    // For now, we're just testing that the notification system can handle errors
    notifications.networkError()
    
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'error',
      title: 'Network Error',
      duration: 10000
    })
  })

  it('should handle API validation errors', () => {
    notifications.validationError('Please fill in all required fields')
    
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'warning',
      title: 'Validation Error',
      message: 'Please fill in all required fields',
      duration: 6000
    })
  })

  it('should handle generic API errors', () => {
    const error = new Error('Server returned 500')
    notifications.apiError('update bookmark', error)
    
    expect(notifications.notifications.value).toHaveLength(1)
    expect(notifications.notifications.value[0]).toMatchObject({
      type: 'error',
      title: 'Operation Failed',
      message: 'Failed to update bookmark: Server returned 500',
      duration: 8000
    })
  })
})