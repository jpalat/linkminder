import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBookmarkStore } from '../bookmarks'
import { bookmarkService } from '@/services/bookmarkService'
import type { Bookmark } from '@/types'

// Mock the bookmark service
vi.mock('@/services/bookmarkService', () => ({
  bookmarkService: {
    updateBookmark: vi.fn(),
    updateBookmarkFull: vi.fn(),
    toBackendUpdateRequest: vi.fn(),
    toBackendFullUpdateRequest: vi.fn(),
    getAllBookmarks: vi.fn(),
    getDashboardStats: vi.fn(),
    createBookmark: vi.fn(),
    toBackendCreateRequest: vi.fn()
  }
}))

// Mock the project service
vi.mock('@/services/projectService', () => ({
  projectService: {
    getProjects: vi.fn().mockResolvedValue({ projects: [] })
  }
}))

// Mock the error utilities
vi.mock('@/services/api', () => ({
  getErrorMessage: vi.fn(),
  isApiError: vi.fn()
}))

vi.mock('@/composables/useApiError', () => ({
  getErrorDisplayMessage: vi.fn(),
  isNetworkError: vi.fn()
}))

describe('Bookmark Store - Update Functionality', () => {
  let store: ReturnType<typeof useBookmarkStore>
  
  const mockBookmark: Bookmark = {
    id: '1',
    url: 'https://example.com',
    title: 'Original Title',
    description: 'Original description',
    action: 'read-later',
    topic: 'testing',
    timestamp: '2024-01-15T10:30:00Z',
    domain: 'example.com',
    age: '1d'
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useBookmarkStore()
    
    // Set up initial bookmark in store
    store.bookmarks = [mockBookmark]
    
    // Clear all mocks
    vi.clearAllMocks()
  })

  describe('updateBookmark - Smart Routing Logic', () => {
    it('should use PUT endpoint when updating title', async () => {
      const updatedBookmark = { ...mockBookmark, title: 'New Title' }
      const mockBackendRequest = {
        title: 'New Title',
        url: 'https://example.com',
        description: 'Original description',
        action: 'read-later',
        topic: 'testing'
      }

      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { title: 'New Title' })

      expect(bookmarkService.toBackendFullUpdateRequest).toHaveBeenCalledWith({
        ...mockBookmark,
        title: 'New Title'
      })
      expect(bookmarkService.updateBookmarkFull).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmark).not.toHaveBeenCalled()
    })

    it('should use PUT endpoint when updating URL', async () => {
      const updatedBookmark = { ...mockBookmark, url: 'https://newexample.com' }
      const mockBackendRequest = {
        title: 'Original Title',
        url: 'https://newexample.com',
        description: 'Original description',
        action: 'read-later',
        topic: 'testing'
      }

      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { url: 'https://newexample.com' })

      expect(bookmarkService.updateBookmarkFull).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmark).not.toHaveBeenCalled()
    })

    it('should use PUT endpoint when updating description', async () => {
      const updatedBookmark = { ...mockBookmark, description: 'New description' }
      const mockBackendRequest = {
        title: 'Original Title',
        url: 'https://example.com',
        description: 'New description',
        action: 'read-later',
        topic: 'testing'
      }

      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { description: 'New description' })

      expect(bookmarkService.updateBookmarkFull).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmark).not.toHaveBeenCalled()
    })

    it('should use PUT endpoint when clearing description', async () => {
      const updatedBookmark = { ...mockBookmark, description: undefined }
      const mockBackendRequest = {
        title: 'Original Title',
        url: 'https://example.com',
        description: undefined,
        action: 'read-later',
        topic: 'testing'
      }

      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { description: undefined })

      expect(bookmarkService.updateBookmarkFull).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmark).not.toHaveBeenCalled()
    })

    it('should use PATCH endpoint when only updating action', async () => {
      const updatedBookmark = { ...mockBookmark, action: 'working' }
      const mockBackendRequest = { action: 'working' }

      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { action: 'working' })

      expect(bookmarkService.toBackendUpdateRequest).toHaveBeenCalledWith({ action: 'working' })
      expect(bookmarkService.updateBookmark).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmarkFull).not.toHaveBeenCalled()
    })

    it('should use PATCH endpoint when only updating topic', async () => {
      const updatedBookmark = { ...mockBookmark, topic: 'new-topic' }
      const mockBackendRequest = { topic: 'new-topic' }

      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { topic: 'new-topic' })

      expect(bookmarkService.updateBookmark).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmarkFull).not.toHaveBeenCalled()
    })

    it('should use PATCH endpoint when only updating shareTo', async () => {
      const updatedBookmark = { ...mockBookmark, shareTo: 'Newsletter' }
      const mockBackendRequest = { shareTo: 'Newsletter' }

      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { shareTo: 'Newsletter' })

      expect(bookmarkService.updateBookmark).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmarkFull).not.toHaveBeenCalled()
    })

    it('should use PUT endpoint when updating both content and metadata fields', async () => {
      const updatedBookmark = { ...mockBookmark, title: 'New Title', action: 'working' }
      const mockBackendRequest = {
        title: 'New Title',
        url: 'https://example.com',
        description: 'Original description',
        action: 'working',
        topic: 'testing'
      }

      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { title: 'New Title', action: 'working' })

      expect(bookmarkService.updateBookmarkFull).toHaveBeenCalledWith('1', mockBackendRequest)
      expect(bookmarkService.updateBookmark).not.toHaveBeenCalled()
    })
  })

  describe('Local State Updates', () => {
    it('should update local state after successful PATCH', async () => {
      const updatedBookmark = { ...mockBookmark, action: 'working' }
      
      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue({ action: 'working' })
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { action: 'working' })

      expect(store.bookmarks[0].action).toBe('working')
    })

    it('should update local state after successful PUT', async () => {
      const updatedBookmark = { ...mockBookmark, title: 'Updated Title' }
      
      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue({
        title: 'Updated Title',
        url: 'https://example.com',
        description: 'Original description',
        action: 'read-later',
        topic: 'testing'
      })
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { title: 'Updated Title' })

      expect(store.bookmarks[0].title).toBe('Updated Title')
    })

    it('should preserve other fields when updating specific fields', async () => {
      const updatedBookmark = { ...mockBookmark, action: 'working' }
      
      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue({ action: 'working' })
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { action: 'working' })

      const bookmark = store.bookmarks[0]
      expect(bookmark.title).toBe('Original Title')
      expect(bookmark.url).toBe('https://example.com')
      expect(bookmark.description).toBe('Original description')
      expect(bookmark.action).toBe('working')
      expect(bookmark.topic).toBe('testing')
    })
  })

  describe('Error Handling', () => {
    it('should throw error when bookmark not found for full update', async () => {
      await expect(
        store.updateBookmark('999', { title: 'New Title' })
      ).rejects.toThrow('Bookmark with ID 999 not found')
    })

    it('should handle PATCH endpoint errors', async () => {
      const mockError = new Error('Network error')
      
      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue({ action: 'working' })
      ;(bookmarkService.updateBookmark as Mock).mockRejectedValue(mockError)

      await expect(
        store.updateBookmark('1', { action: 'working' })
      ).rejects.toThrow('Network error')
    })

    it('should handle PUT endpoint errors', async () => {
      const mockError = new Error('Validation error')
      
      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue({
        title: 'New Title',
        url: 'https://example.com',
        description: 'Original description',
        action: 'read-later',
        topic: 'testing'
      })
      ;(bookmarkService.updateBookmarkFull as Mock).mockRejectedValue(mockError)

      await expect(
        store.updateBookmark('1', { title: 'New Title' })
      ).rejects.toThrow('Validation error')
    })

    it('should not update local state when API call fails', async () => {
      const originalTitle = store.bookmarks[0].title
      
      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue({
        title: 'New Title',
        url: 'https://example.com',
        description: 'Original description',
        action: 'read-later',
        topic: 'testing'
      })
      ;(bookmarkService.updateBookmarkFull as Mock).mockRejectedValue(new Error('API Error'))

      try {
        await store.updateBookmark('1', { title: 'New Title' })
      } catch (error) {
        // Expected to throw
      }

      expect(store.bookmarks[0].title).toBe(originalTitle)
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty updates object', async () => {
      const mockBackendRequest = {}
      
      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(mockBookmark)

      await store.updateBookmark('1', {})

      expect(bookmarkService.updateBookmark).toHaveBeenCalledWith('1', mockBackendRequest)
    })

    it('should handle updates with null/undefined values', async () => {
      const updatedBookmark = { ...mockBookmark, shareTo: undefined }
      const mockBackendRequest = { shareTo: undefined }

      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      await store.updateBookmark('1', { shareTo: undefined })

      expect(bookmarkService.updateBookmark).toHaveBeenCalledWith('1', mockBackendRequest)
    })

    it('should handle updates to bookmark without topic field', async () => {
      const bookmarkWithoutTopic = { ...mockBookmark, topic: undefined }
      store.bookmarks = [bookmarkWithoutTopic]

      const mockBackendRequest = {
        title: 'New Title',
        url: 'https://example.com',
        description: 'Original description',
        action: 'read-later',
        topic: undefined
      }

      ;(bookmarkService.toBackendFullUpdateRequest as Mock).mockReturnValue(mockBackendRequest)
      ;(bookmarkService.updateBookmarkFull as Mock).mockResolvedValue({
        ...bookmarkWithoutTopic,
        title: 'New Title'
      })

      await store.updateBookmark('1', { title: 'New Title' })

      expect(bookmarkService.updateBookmarkFull).toHaveBeenCalledWith('1', mockBackendRequest)
    })
  })

  describe('Integration with moveBookmarks', () => {
    it('should use PATCH endpoint when moving bookmarks via moveBookmarks', async () => {
      const updatedBookmark = { ...mockBookmark, action: 'working' }
      
      ;(bookmarkService.toBackendUpdateRequest as Mock).mockReturnValue({ action: 'working' })
      ;(bookmarkService.updateBookmark as Mock).mockResolvedValue(updatedBookmark)

      store.moveBookmarks(['1'], 'working')

      // Wait for async operation
      await new Promise(resolve => setTimeout(resolve, 0))

      expect(bookmarkService.updateBookmark).toHaveBeenCalledWith('1', { action: 'working' })
      expect(bookmarkService.updateBookmarkFull).not.toHaveBeenCalled()
    })
  })
})