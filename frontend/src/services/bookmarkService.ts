import { apiClient, type ApiResponse } from './api'
import type { Bookmark } from '@/types'

// API request/response types matching the Go backend
export interface BookmarkCreateRequest {
  url: string
  title: string
  description?: string
  content?: string
  action?: 'read-later' | 'working' | 'share' | 'archived' | 'irrelevant'
  shareTo?: string
  topic?: string        // Legacy support
  projectId?: number    // New field
}

export interface BookmarkUpdateRequest {
  action?: string
  shareTo?: string
  topic?: string        // Legacy support
  projectId?: number    // New field
}

export interface BookmarkFullUpdateRequest {
  title: string
  url: string
  description?: string
  action?: string
  shareTo?: string
  topic?: string
}

export interface TriageBookmark {
  id: number
  url: string
  title: string
  description?: string
  content?: string
  timestamp: string
  domain?: string
  age?: string
  action?: string
}

export interface TriageResponse {
  bookmarks: TriageBookmark[]
  total: number
  limit: number
  offset: number
  hasMore: boolean
}

export interface DashboardStats {
  needsTriage: number
  activeProjects: number
  readyToShare: number
  archived: number
  totalBookmarks: number
  projectStats: Array<{
    topic: string
    count: number
    lastUpdated: string
    status: 'active' | 'stale' | 'inactive'
  }>
}

export interface TopicsResponse {
  topics: string[]
}

class BookmarkService {
  /**
   * Create a new bookmark
   * POST /bookmark
   */
  async createBookmark(bookmark: BookmarkCreateRequest): Promise<Bookmark> {
    const response = await apiClient.post<any>('/bookmark', bookmark)
    
    // Transform the response to match our frontend Bookmark interface
    return this.transformBackendBookmark(response.data)
  }

  /**
   * Update a bookmark (partial update)
   * PATCH /api/bookmarks/{id}
   */
  async updateBookmark(id: string, updates: BookmarkUpdateRequest): Promise<Bookmark> {
    const response = await apiClient.patch<any>(`/api/bookmarks/${id}`, updates)
    return this.transformBackendBookmark(response.data)
  }

  /**
   * Update a bookmark (full update)
   * PUT /api/bookmarks/{id}
   */
  async updateBookmarkFull(id: string, bookmark: BookmarkFullUpdateRequest): Promise<Bookmark> {
    const response = await apiClient.put<any>(`/api/bookmarks/${id}`, bookmark)
    return this.transformBackendBookmark(response.data)
  }

  /**
   * Get bookmarks needing triage
   * GET /api/bookmarks/triage
   */
  async getTriageQueue(limit: number = 50, offset: number = 0): Promise<TriageResponse> {
    const response = await apiClient.get<TriageResponse>('/api/bookmarks/triage', {
      limit,
      offset
    })
    
    return response.data
  }

  /**
   * Get dashboard summary statistics
   * GET /api/stats/summary
   */
  async getDashboardStats(): Promise<DashboardStats> {
    const response = await apiClient.get<DashboardStats>('/api/stats/summary')
    return response.data
  }

  /**
   * Get list of available topics
   * GET /topics
   */
  async getTopics(): Promise<string[]> {
    const response = await apiClient.get<string[]>('/topics')
    return response.data
  }

  /**
   * Transform backend bookmark data to frontend Bookmark interface
   * Handles differences between backend and frontend data structures
   */
  private transformBackendBookmark(backendBookmark: any): Bookmark {
    return {
      id: String(backendBookmark.id || backendBookmark.ID), // Handle both cases
      url: backendBookmark.url,
      title: backendBookmark.title,
      description: backendBookmark.description || undefined,
      content: backendBookmark.content || undefined,
      action: backendBookmark.action || undefined,
      shareTo: backendBookmark.shareTo || backendBookmark.share_to || undefined,
      topic: backendBookmark.topic || undefined,
      project_id: backendBookmark.projectId || backendBookmark.project_id || undefined,
      timestamp: backendBookmark.timestamp || backendBookmark.created_at || new Date().toISOString(),
      domain: backendBookmark.domain || this.extractDomain(backendBookmark.url),
      age: backendBookmark.age || this.calculateAge(backendBookmark.timestamp || backendBookmark.created_at),
      tags: backendBookmark.tags || undefined
    }
  }

  /**
   * Extract domain from URL
   */
  private extractDomain(url: string): string {
    try {
      return new URL(url).hostname
    } catch {
      return 'unknown'
    }
  }

  /**
   * Calculate age string from timestamp
   */
  private calculateAge(timestamp?: string): string {
    if (!timestamp) return 'unknown'
    
    const now = new Date()
    const created = new Date(timestamp)
    const diffMs = now.getTime() - created.getTime()
    
    const diffMinutes = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
    const diffWeeks = Math.floor(diffDays / 7)
    const diffMonths = Math.floor(diffDays / 30)
    
    if (diffMinutes < 1) return 'just now'
    if (diffMinutes < 60) return `${diffMinutes}m`
    if (diffHours < 24) return `${diffHours}h`
    if (diffDays < 7) return `${diffDays}d`
    if (diffWeeks < 4) return `${diffWeeks}w`
    return `${diffMonths}mo`
  }

  /**
   * Convert frontend Bookmark to backend format for creation
   */
  toBackendCreateRequest(bookmark: Partial<Bookmark>): BookmarkCreateRequest {
    if (!bookmark.url || !bookmark.title) {
      throw new Error('URL and title are required')
    }

    return {
      url: bookmark.url,
      title: bookmark.title,
      description: bookmark.description,
      content: bookmark.content,
      action: bookmark.action,
      shareTo: bookmark.shareTo,
      topic: bookmark.topic,
      projectId: bookmark.project_id
    }
  }

  /**
   * Convert frontend Bookmark to backend format for updates
   */
  toBackendUpdateRequest(bookmark: Partial<Bookmark>): BookmarkUpdateRequest {
    return {
      action: bookmark.action,
      shareTo: bookmark.shareTo,
      topic: bookmark.topic,
      projectId: bookmark.project_id
    }
  }

  /**
   * Convert frontend Bookmark to backend format for full updates
   */
  toBackendFullUpdateRequest(bookmark: Bookmark): BookmarkFullUpdateRequest {
    return {
      title: bookmark.title,
      url: bookmark.url,
      description: bookmark.description,
      action: bookmark.action,
      shareTo: bookmark.shareTo,
      topic: bookmark.topic
    }
  }

  /**
   * Batch operations helpers (for future API endpoints)
   * These will be used when the bulk APIs are implemented
   */
  
  /**
   * Prepare bulk update data (for future PATCH /bookmarks/bulk endpoint)
   */
  prepareBulkUpdate(bookmarkIds: string[], updates: BookmarkUpdateRequest) {
    return {
      ids: bookmarkIds.map(id => parseInt(id, 10)),
      updates
    }
  }

  /**
   * Prepare bulk delete data (for future DELETE /bookmarks/bulk endpoint)  
   */
  prepareBulkDelete(bookmarkIds: string[]) {
    return {
      ids: bookmarkIds.map(id => parseInt(id, 10))
    }
  }
}

export const bookmarkService = new BookmarkService()
export default bookmarkService