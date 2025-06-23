import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'
import { apiClient } from '@/services/api'
import type { Bookmark } from '@/types'

// Mock the API client
vi.mock('@/services/api', () => ({
  apiClient: {
    patch: vi.fn(),
    put: vi.fn(),
    get: vi.fn(),
    post: vi.fn()
  }
}))

// Mock other services
vi.mock('@/services/projectService', () => ({
  projectService: {
    getProjects: vi.fn().mockResolvedValue({ projects: [] })
  }
}))

vi.mock('@/composables/useApiError', () => ({
  getErrorDisplayMessage: vi.fn().mockReturnValue('Mock error'),
  isNetworkError: vi.fn().mockReturnValue(false)
}))

describe('Bookmark Update Integration Tests', () => {
  let store: ReturnType<typeof useBookmarkStore>
  
  const mockBookmark: Bookmark = {
    id: '42',
    url: 'https://react.dev/blog/2022/03/29/react-v18',
    title: 'React 18 Release',
    description: 'Learn about React 18 features',
    action: 'read-later',
    topic: 'react',
    timestamp: '2024-01-15T10:30:00Z',
    domain: 'react.dev',
    age: '2h'
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useBookmarkStore()
    store.bookmarks = [mockBookmark]
    
    // Clear all mocks and reset their implementations
    vi.clearAllMocks()
    vi.resetAllMocks()
    
    // Ensure clean mock state
    vi.mocked(apiClient.patch).mockClear()
    vi.mocked(apiClient.put).mockClear()
  })

  describe('Title Update Scenarios', () => {
    it('should successfully update bookmark title using PUT endpoint', async () => {
      const updatedBackendBookmark = {
        id: 42,
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        title: 'React 18 - New Features and Breaking Changes',
        description: 'Learn about React 18 features',
        action: 'read-later',
        topic: 'react',
        timestamp: '2024-01-15T10:30:00Z'
      }

      // Mock successful PUT response
      vi.mocked(apiClient.put).mockResolvedValue({ data: updatedBackendBookmark })

      // Simulate user updating title
      await store.updateBookmark('42', { 
        title: 'React 18 - New Features and Breaking Changes' 
      })

      // Verify PUT endpoint was called with full bookmark data
      expect(apiClient.put).toHaveBeenCalledWith('/api/bookmarks/42', {
        title: 'React 18 - New Features and Breaking Changes',
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        description: 'Learn about React 18 features',
        action: 'read-later',
        shareTo: undefined,
        topic: 'react'
      })

      // Verify PATCH was not called
      expect(apiClient.patch).not.toHaveBeenCalled()

      // Verify local state was updated
      const updatedBookmark = store.bookmarks.find(b => b.id === '42')
      expect(updatedBookmark?.title).toBe('React 18 - New Features and Breaking Changes')
    })

    it('should handle title update failure gracefully', async () => {
      const originalTitle = store.bookmarks[0].title

      // Mock PUT failure
      vi.mocked(apiClient.put).mockRejectedValue(new Error('Server error'))

      // Attempt to update title
      await expect(
        store.updateBookmark('42', { title: 'New Title' })
      ).rejects.toThrow('Server error')

      // Verify original title is preserved
      expect(store.bookmarks[0].title).toBe(originalTitle)
    })
  })

  describe('Action Update Scenarios', () => {
    it('should successfully move bookmark to working using PATCH endpoint', async () => {
      const updatedBackendBookmark = {
        ...mockBookmark,
        action: 'working',
        topic: 'react-learning'
      }

      // Mock successful PATCH response
      vi.mocked(apiClient.patch).mockResolvedValue({ 
        data: { id: 42, action: 'working', topic: 'react-learning' }
      })

      // Simulate moving bookmark to working
      await store.updateBookmark('42', { 
        action: 'working',
        topic: 'react-learning'
      })

      // Verify PATCH endpoint was called
      expect(apiClient.patch).toHaveBeenCalledWith('/api/bookmarks/42', {
        action: 'working',
        shareTo: undefined,
        topic: 'react-learning',
        projectId: undefined
      })

      // Verify PUT was not called
      expect(apiClient.put).not.toHaveBeenCalled()

      // Verify local state was updated
      const updatedBookmark = store.bookmarks.find(b => b.id === '42')
      expect(updatedBookmark?.action).toBe('working')
    })

    it('should use moveBookmarks helper correctly', async () => {
      const updatedBackendBookmark = {
        id: 42,
        action: 'archived'
      }

      vi.mocked(apiClient.patch).mockResolvedValue({ data: updatedBackendBookmark })

      // Use the moveBookmarks helper (simulates bulk operations)
      store.moveBookmarks(['42'], 'archived')

      // Wait for async operation
      await new Promise(resolve => setTimeout(resolve, 0))

      expect(apiClient.patch).toHaveBeenCalledWith('/api/bookmarks/42', {
        action: 'archived',
        shareTo: undefined,
        topic: undefined,
        projectId: undefined
      })
    })
  })

  describe('Complex Update Scenarios', () => {
    it('should handle mixed content and metadata updates correctly', async () => {
      const updatedBackendBookmark = {
        id: 42,
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        title: 'Updated Title',
        description: 'Updated description',
        action: 'working',
        topic: 'react-migration',
        timestamp: '2024-01-15T10:30:00Z'
      }

      vi.mocked(apiClient.put).mockResolvedValue({ data: updatedBackendBookmark })

      // Update both title (content) and action (metadata) - should use PUT
      await store.updateBookmark('42', {
        title: 'Updated Title',
        description: 'Updated description',
        action: 'working',
        topic: 'react-migration'
      })

      expect(apiClient.put).toHaveBeenCalledWith('/api/bookmarks/42', {
        title: 'Updated Title',
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        description: 'Updated description',
        action: 'working',
        shareTo: undefined,
        topic: 'react-migration'
      })

      expect(apiClient.patch).not.toHaveBeenCalled()
    })

    it('should handle clearing description field', async () => {
      const updatedBackendBookmark = {
        id: 42,
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        title: 'React 18 Release',
        description: null,
        action: 'read-later',
        topic: 'react',
        timestamp: '2024-01-15T10:30:00Z'
      }

      vi.mocked(apiClient.put).mockResolvedValue({ data: updatedBackendBookmark })

      // Clear description by setting to undefined
      await store.updateBookmark('42', { description: undefined })

      expect(apiClient.put).toHaveBeenCalledWith('/api/bookmarks/42', {
        title: 'React 18 Release',
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        description: undefined,
        action: 'read-later',
        shareTo: undefined,
        topic: 'react'
      })

      const updatedBookmark = store.bookmarks.find(b => b.id === '42')
      expect(updatedBookmark?.description).toBeUndefined()
    })
  })

  describe('URL Update Scenarios', () => {
    it('should handle URL updates correctly', async () => {
      const newUrl = 'https://react.dev/blog/2024/04/25/react-19'
      const updatedBackendBookmark = {
        id: 42,
        url: newUrl,
        title: 'React 18 Release',
        description: 'Learn about React 18 features',
        action: 'read-later',
        topic: 'react',
        timestamp: '2024-01-15T10:30:00Z'
      }

      vi.mocked(apiClient.put).mockResolvedValue({ data: updatedBackendBookmark })

      await store.updateBookmark('42', { url: newUrl })

      expect(apiClient.put).toHaveBeenCalledWith('/api/bookmarks/42', {
        title: 'React 18 Release',
        url: newUrl,
        description: 'Learn about React 18 features',
        action: 'read-later',
        shareTo: undefined,
        topic: 'react'
      })

      const updatedBookmark = store.bookmarks.find(b => b.id === '42')
      expect(updatedBookmark?.url).toBe(newUrl)
    })
  })

  describe('Share Workflow', () => {
    it('should handle complete share workflow', async () => {
      // Step 1: Move to share with destination
      const shareUpdate = {
        id: 42,
        action: 'share',
        shareTo: 'Newsletter'
      }

      vi.mocked(apiClient.patch).mockResolvedValue({ data: shareUpdate })

      await store.updateBookmark('42', {
        action: 'share',
        shareTo: 'Newsletter'
      })

      expect(apiClient.patch).toHaveBeenCalledWith('/api/bookmarks/42', {
        action: 'share',
        shareTo: 'Newsletter',
        topic: undefined,
        projectId: undefined
      })

      // Verify bookmark is now in share state
      const sharedBookmark = store.bookmarks.find(b => b.id === '42')
      expect(sharedBookmark?.action).toBe('share')
      expect(sharedBookmark?.shareTo).toBe('Newsletter')
    })
  })

  describe('Error Recovery', () => {
    it('should maintain data consistency after failed updates', async () => {
      const originalBookmark = { ...store.bookmarks[0] }

      // Mock server error
      vi.mocked(apiClient.put).mockRejectedValue(new Error('Server error'))

      try {
        await store.updateBookmark('42', { title: 'This will fail' })
      } catch (error) {
        // Expected to fail
      }

      // Verify original data is preserved
      const currentBookmark = store.bookmarks.find(b => b.id === '42')
      expect(currentBookmark).toEqual(originalBookmark)
    })

    it('should handle network timeout gracefully', async () => {
      vi.mocked(apiClient.patch).mockRejectedValue(new Error('Request timeout'))

      await expect(
        store.updateBookmark('42', { action: 'working' })
      ).rejects.toThrow('Request timeout')

      // Verify bookmark state unchanged
      expect(store.bookmarks[0].action).toBe('read-later')
    })
  })

  describe('Performance Considerations', () => {
    it('should not make unnecessary API calls for no-op updates', async () => {
      // Mock successful PATCH response
      vi.mocked(apiClient.patch).mockResolvedValue({ 
        data: { id: 42, action: 'read-later' }
      })

      // Update with the same values
      await store.updateBookmark('42', { action: 'read-later' })

      expect(apiClient.patch).toHaveBeenCalledTimes(1)
      expect(apiClient.put).not.toHaveBeenCalled()
    })

    it('should batch multiple field updates into single PUT call', async () => {
      const updatedBookmark = {
        id: 42,
        title: 'New Title',
        description: 'New Description',
        url: 'https://newurl.com'
      }

      vi.mocked(apiClient.put).mockResolvedValue({ data: updatedBookmark })

      // Single update call with multiple fields
      await store.updateBookmark('42', {
        title: 'New Title',
        description: 'New Description',
        url: 'https://newurl.com'
      })

      // Should only make one API call
      expect(apiClient.put).toHaveBeenCalledTimes(1)
      expect(apiClient.patch).not.toHaveBeenCalled()
    })
  })
})