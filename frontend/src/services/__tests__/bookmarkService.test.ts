import { describe, it, expect, beforeEach, vi } from 'vitest'
import { bookmarkService } from '../bookmarkService'
import { apiClient } from '../api'
import type { Bookmark } from '@/types'

// Mock the API client
vi.mock('../api', () => ({
  apiClient: {
    patch: vi.fn(),
    put: vi.fn(),
    get: vi.fn(),
    post: vi.fn()
  }
}))

describe('BookmarkService - Update Methods', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockBookmark: Bookmark = {
    id: '123',
    url: 'https://example.com',
    title: 'Test Bookmark',
    description: 'Test description',
    action: 'read-later',
    topic: 'testing',
    shareTo: 'Newsletter',
    timestamp: '2024-01-15T10:30:00Z',
    domain: 'example.com',
    age: '1d'
  }

  const mockBackendBookmark = {
    id: 123,
    url: 'https://example.com',
    title: 'Test Bookmark',
    description: 'Test description',
    action: 'read-later',
    topic: 'testing',
    shareTo: 'Newsletter',
    timestamp: '2024-01-15T10:30:00Z'
  }

  describe('updateBookmark (PATCH)', () => {
    it('should call PATCH endpoint with correct data', async () => {
      const mockResponse = { data: mockBackendBookmark }
      vi.mocked(apiClient.patch).mockResolvedValue(mockResponse)

      const updates = { action: 'working', topic: 'new-topic' }
      
      await bookmarkService.updateBookmark('123', updates)

      expect(apiClient.patch).toHaveBeenCalledWith('/api/bookmarks/123', updates)
    })

    it('should transform backend response to frontend format', async () => {
      const mockResponse = { data: mockBackendBookmark }
      vi.mocked(apiClient.patch).mockResolvedValue(mockResponse)

      const result = await bookmarkService.updateBookmark('123', { action: 'working' })

      expect(result).toMatchObject({
        id: '123',
        url: 'https://example.com',
        title: 'Test Bookmark',
        description: 'Test description',
        action: 'read-later',
        topic: 'testing',
        domain: 'example.com'
      })
    })
  })

  describe('updateBookmarkFull (PUT)', () => {
    it('should call PUT endpoint with correct data', async () => {
      const mockResponse = { data: mockBackendBookmark }
      vi.mocked(apiClient.put).mockResolvedValue(mockResponse)

      const fullUpdate = {
        title: 'Updated Title',
        url: 'https://newexample.com',
        description: 'Updated description',
        action: 'working',
        shareTo: 'Team Slack',
        topic: 'new-topic'
      }
      
      await bookmarkService.updateBookmarkFull('123', fullUpdate)

      expect(apiClient.put).toHaveBeenCalledWith('/api/bookmarks/123', fullUpdate)
    })

    it('should transform backend response to frontend format', async () => {
      const updatedBackendBookmark = {
        ...mockBackendBookmark,
        title: 'Updated Title',
        url: 'https://newexample.com',
        description: 'Updated description'
      }
      const mockResponse = { data: updatedBackendBookmark }
      vi.mocked(apiClient.put).mockResolvedValue(mockResponse)

      const fullUpdate = {
        title: 'Updated Title',
        url: 'https://newexample.com',
        description: 'Updated description',
        action: 'working',
        shareTo: 'Team Slack',
        topic: 'new-topic'
      }

      const result = await bookmarkService.updateBookmarkFull('123', fullUpdate)

      expect(result).toMatchObject({
        id: '123',
        title: 'Updated Title',
        url: 'https://newexample.com',
        description: 'Updated description'
      })
    })
  })

  describe('toBackendUpdateRequest', () => {
    it('should convert frontend partial bookmark to backend update format', () => {
      const frontendUpdate: Partial<Bookmark> = {
        action: 'working',
        shareTo: 'Newsletter',
        topic: 'test-topic',
        project_id: 42
      }

      const result = bookmarkService.toBackendUpdateRequest(frontendUpdate)

      expect(result).toEqual({
        action: 'working',
        shareTo: 'Newsletter',
        topic: 'test-topic',
        projectId: 42
      })
    })

    it('should handle undefined values', () => {
      const frontendUpdate: Partial<Bookmark> = {
        action: undefined,
        shareTo: undefined,
        topic: undefined,
        project_id: undefined
      }

      const result = bookmarkService.toBackendUpdateRequest(frontendUpdate)

      expect(result).toEqual({
        action: undefined,
        shareTo: undefined,
        topic: undefined,
        projectId: undefined
      })
    })

    it('should ignore non-update fields', () => {
      const frontendUpdate: Partial<Bookmark> = {
        id: '123',
        title: 'Should be ignored',
        url: 'Should be ignored',
        description: 'Should be ignored',
        action: 'working',
        topic: 'test-topic'
      }

      const result = bookmarkService.toBackendUpdateRequest(frontendUpdate)

      expect(result).toEqual({
        action: 'working',
        shareTo: undefined,
        topic: 'test-topic',
        projectId: undefined
      })
      expect(result).not.toHaveProperty('id')
      expect(result).not.toHaveProperty('title')
      expect(result).not.toHaveProperty('url')
      expect(result).not.toHaveProperty('description')
    })
  })

  describe('toBackendFullUpdateRequest', () => {
    it('should convert frontend bookmark to backend full update format', () => {
      const result = bookmarkService.toBackendFullUpdateRequest(mockBookmark)

      expect(result).toEqual({
        title: 'Test Bookmark',
        url: 'https://example.com',
        description: 'Test description',
        action: 'read-later',
        shareTo: 'Newsletter',
        topic: 'testing'
      })
    })

    it('should handle bookmark without optional fields', () => {
      const minimalBookmark: Bookmark = {
        id: '123',
        url: 'https://example.com',
        title: 'Minimal Bookmark',
        timestamp: '2024-01-15T10:30:00Z',
        domain: 'example.com',
        age: '1d'
      }

      const result = bookmarkService.toBackendFullUpdateRequest(minimalBookmark)

      expect(result).toEqual({
        title: 'Minimal Bookmark',
        url: 'https://example.com',
        description: undefined,
        action: undefined,
        shareTo: undefined,
        topic: undefined
      })
    })

    it('should not include frontend-only fields', () => {
      const result = bookmarkService.toBackendFullUpdateRequest(mockBookmark)

      expect(result).not.toHaveProperty('id')
      expect(result).not.toHaveProperty('timestamp')
      expect(result).not.toHaveProperty('domain')
      expect(result).not.toHaveProperty('age')
      expect(result).not.toHaveProperty('project_id')
    })
  })

  describe('transformBackendBookmark', () => {
    it('should transform backend bookmark with all fields', () => {
      const backendBookmark = {
        id: 123,
        url: 'https://example.com',
        title: 'Backend Title',
        description: 'Backend description',
        content: 'Full content',
        action: 'working',
        shareTo: 'Team Slack',
        topic: 'backend-topic',
        project_id: 42,
        timestamp: '2024-01-15T10:30:00Z',
        domain: 'example.com',
        age: '2h'
      }

      // Access the private method through the service instance
      const result = (bookmarkService as any).transformBackendBookmark(backendBookmark)

      expect(result).toMatchObject({
        id: '123',
        url: 'https://example.com',
        title: 'Backend Title',
        description: 'Backend description',
        content: 'Full content',
        action: 'working',
        shareTo: 'Team Slack',
        topic: 'backend-topic',
        project_id: 42,
        timestamp: '2024-01-15T10:30:00Z',
        domain: 'example.com',
        age: '2h'
      })
    })

    it('should handle backend bookmark with ID field instead of id', () => {
      const backendBookmark = {
        ID: 456,
        url: 'https://example.com',
        title: 'Backend Title',
        timestamp: '2024-01-15T10:30:00Z'
      }

      const result = (bookmarkService as any).transformBackendBookmark(backendBookmark)

      expect(result.id).toBe('456')
    })

    it('should handle backend bookmark with snake_case fields', () => {
      const backendBookmark = {
        id: 123,
        url: 'https://example.com',
        title: 'Backend Title',
        share_to: 'Newsletter',
        created_at: '2024-01-15T10:30:00Z'
      }

      const result = (bookmarkService as any).transformBackendBookmark(backendBookmark)

      expect(result.shareTo).toBe('Newsletter')
      expect(result.timestamp).toBe('2024-01-15T10:30:00Z')
    })

    it('should extract domain from URL when not provided', () => {
      const backendBookmark = {
        id: 123,
        url: 'https://test.example.com/path',
        title: 'Test',
        timestamp: '2024-01-15T10:30:00Z'
      }

      const result = (bookmarkService as any).transformBackendBookmark(backendBookmark)

      expect(result.domain).toBe('test.example.com')
    })

    it('should calculate age when not provided', () => {
      const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000).toISOString()
      const backendBookmark = {
        id: 123,
        url: 'https://example.com',
        title: 'Test',
        timestamp: oneHourAgo
      }

      const result = (bookmarkService as any).transformBackendBookmark(backendBookmark)

      expect(result.age).toBe('1h')
    })
  })

  describe('Error Handling', () => {
    it('should handle network errors in updateBookmark', async () => {
      const networkError = new Error('Network error')
      vi.mocked(apiClient.patch).mockRejectedValue(networkError)

      await expect(
        bookmarkService.updateBookmark('123', { action: 'working' })
      ).rejects.toThrow('Network error')
    })

    it('should handle validation errors in updateBookmarkFull', async () => {
      const validationError = new Error('Validation failed')
      vi.mocked(apiClient.put).mockRejectedValue(validationError)

      const fullUpdate = {
        title: '',
        url: 'invalid-url',
        description: 'Test',
        action: 'working'
      }

      await expect(
        bookmarkService.updateBookmarkFull('123', fullUpdate)
      ).rejects.toThrow('Validation failed')
    })
  })
})