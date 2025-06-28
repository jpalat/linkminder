import { apiClient } from './api'
import type { Bookmark, BookmarkAction } from '@/types'

// API request/response types matching the Go backend

// Backend bookmark response structure
export interface BackendBookmarkResponse {
  id: number | string
  ID?: number | string
  url: string
  title: string
  description?: string
  content?: string
  action?: BookmarkAction
  shareTo?: string
  share_to?: string
  topic?: string
  projectId?: number
  project_id?: number
  timestamp?: string
  created_at?: string
  domain?: string
  age?: string
  tags?: string[]
  customProperties?: Record<string, string>
}
export interface BookmarkCreateRequest {
  url: string
  title: string
  description?: string
  content?: string
  action?: BookmarkAction
  shareTo?: string
  topic?: string        // Legacy support
  projectId?: number    // New field
  tags?: string[]
  customProperties?: Record<string, string>
}

export interface BookmarkUpdateRequest {
  action?: BookmarkAction
  shareTo?: string
  topic?: string        // Legacy support
  projectId?: number    // New field
  tags?: string[]
  customProperties?: Record<string, string>
}

export interface BookmarkFullUpdateRequest {
  title: string
  url: string
  description?: string
  action?: BookmarkAction
  shareTo?: string
  topic?: string
  tags?: string[]
  customProperties?: Record<string, string>
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
  action?: BookmarkAction
  shareTo?: string
  topic?: string
  tags?: string[]
  customProperties?: Record<string, string>
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
    const response = await apiClient.post<BackendBookmarkResponse>('/bookmark', bookmark)
    
    // Transform the response to match our frontend Bookmark interface
    return this.transformBackendBookmark(response.data)
  }

  /**
   * Update a bookmark (partial update)
   * PATCH /api/bookmarks/{id}
   */
  async updateBookmark(id: string, updates: BookmarkUpdateRequest): Promise<Bookmark> {
    const response = await apiClient.patch<BackendBookmarkResponse>(`/api/bookmarks/${id}`, updates)
    return this.transformBackendBookmark(response.data)
  }

  /**
   * Update a bookmark (full update)
   * PUT /api/bookmarks/{id}
   */
  async updateBookmarkFull(id: string, bookmark: BookmarkFullUpdateRequest): Promise<Bookmark> {
    const response = await apiClient.put<BackendBookmarkResponse>(`/api/bookmarks/${id}`, bookmark)
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
   * Get bookmarks by action type
   * GET /api/bookmarks?action={action}
   */
  async getBookmarksByAction(action: string, limit: number = 50): Promise<Bookmark[]> {
    try {
      // For triage bookmarks (read-later or empty), use the dedicated triage endpoint
      if (action === 'read-later' || action === '') {
        const response = await this.getTriageQueue(limit)
        return response.bookmarks.map(bookmark => this.transformTriageBookmark(bookmark))
      } else {
        // For other actions (share, working, archived), use the new bookmarks endpoint
        const response = await apiClient.get<TriageResponse>('/api/bookmarks', {
          action,
          limit
        })
        return response.data.bookmarks.map(bookmark => this.transformTriageBookmark(bookmark))
      }
    } catch (error) {
      console.error(`Error fetching ${action} bookmarks:`, error)
      // Fallback to mock data if API fails
      return this.getMockBookmarksByAction(action)
    }
  }

  /**
   * Get all bookmarks across all action types
   */
  async getAllBookmarks(): Promise<Bookmark[]> {
    try {
      // Load bookmarks from all action types using real API calls
      const [triageBookmarks, shareBookmarks, workingBookmarks, archivedBookmarks] = await Promise.all([
        this.getBookmarksByAction('read-later', 100),
        this.getBookmarksByAction('share', 100),
        this.getBookmarksByAction('working', 100),
        this.getBookmarksByAction('archived', 100)
      ])

      return [
        ...triageBookmarks,
        ...shareBookmarks,
        ...workingBookmarks,
        ...archivedBookmarks
      ]
    } catch (error) {
      console.error('Error fetching all bookmarks:', error)
      // Fallback to all mock data
      return this.getAllMockBookmarks()
    }
  }

  /**
   * Transform triage bookmark data to frontend format
   */
  private transformTriageBookmark(triageBookmark: BackendBookmarkResponse): Bookmark {
    return {
      id: String(triageBookmark.id),
      url: triageBookmark.url,
      title: triageBookmark.title,
      description: triageBookmark.description,
      content: triageBookmark.content,
      action: triageBookmark.action,
      shareTo: triageBookmark.shareTo,
      topic: triageBookmark.topic,
      timestamp: triageBookmark.timestamp || new Date().toISOString(),
      domain: triageBookmark.domain || this.extractDomain(triageBookmark.url),
      age: triageBookmark.age || this.calculateAge(triageBookmark.timestamp),
      tags: triageBookmark.tags,
      customProperties: triageBookmark.customProperties
    }
  }

  /**
   * Get mock bookmarks filtered by action
   */
  private getMockBookmarksByAction(action: string): Bookmark[] {
    const allMockBookmarks = this.getAllMockBookmarks()
    return allMockBookmarks.filter(bookmark => bookmark.action === action)
  }

  /**
   * Get all mock bookmarks
   */
  private getAllMockBookmarks(): Bookmark[] {
    return [
      // Triage items (read-later or no action)
      {
        id: '1',
        url: 'https://react.dev/blog/2022/03/29/react-v18',
        title: 'Building Modern Web Applications with React 18',
        description: 'Learn about the new features in React 18 including concurrent rendering, automatic batching, and more.',
        action: 'read-later',
        timestamp: '2024-01-15T10:30:00Z',
        domain: 'react.dev',
        age: '2h'
      },
      {
        id: '2',
        url: 'https://vuejs.org/guide/essentials/reactivity.html',
        title: 'Vue.js Reactivity Fundamentals',
        description: 'Understanding how Vue.js reactivity system works under the hood.',
        timestamp: '2024-01-15T08:15:00Z',
        domain: 'vuejs.org',
        age: '4h'
      },
      {
        id: '3',
        url: 'https://developer.mozilla.org/en-US/docs/Web/API/Web_Components',
        title: 'Web Components MDN Guide',
        description: 'Complete guide to creating reusable custom elements with Web Components.',
        action: 'read-later',
        timestamp: '2024-01-14T20:00:00Z',
        domain: 'developer.mozilla.org',
        age: '18h'
      },
      
      // Working items
      {
        id: '4',
        url: 'https://platform.openai.com/docs/guides/gpt',
        title: 'OpenAI GPT-4 API Documentation',
        description: 'Complete guide to using the GPT-4 API for building AI-powered applications.',
        action: 'working',
        topic: 'ai-tools',
        timestamp: '2024-01-15T05:30:00Z',
        domain: 'openai.com',
        age: '5h'
      },
      {
        id: '5',
        url: 'https://blog.logrocket.com/svelte-vs-react/',
        title: 'Svelte vs React: Which Should You Choose?',
        description: 'A detailed comparison of Svelte and React frameworks for modern web development.',
        action: 'working',
        topic: 'framework-research',
        timestamp: '2024-01-13T12:30:00Z',
        domain: 'logrocket.com',
        age: '2d'
      },
      {
        id: '6',
        url: 'https://react.dev/learn/migration-guide',
        title: 'React Migration Guide',
        description: 'Step-by-step guide for migrating from React 17 to React 18.',
        action: 'working',
        topic: 'react-migration',
        timestamp: '2024-01-13T09:00:00Z',
        domain: 'react.dev',
        age: '2d'
      },
      
      // Share items
      {
        id: '7',
        url: 'https://css-tricks.com/snippets/css/complete-guide-grid/',
        title: 'CSS Grid Complete Guide - A comprehensive guide to CSS Grid',
        description: 'Everything you need to know about CSS Grid layout with practical examples.',
        action: 'share',
        shareTo: 'Newsletter',
        topic: 'css-learning',
        timestamp: '2024-01-14T15:30:00Z',
        domain: 'css-tricks.com',
        age: '1d'
      },
      {
        id: '8',
        url: 'https://kentcdodds.com/blog/react-performance-tips',
        title: 'React Performance Tips',
        description: 'Essential tips for optimizing React application performance.',
        action: 'share',
        shareTo: 'Team Slack',
        topic: 'react-migration',
        timestamp: '2024-01-14T14:00:00Z',
        domain: 'kentcdodds.com',
        age: '1d'
      },
      {
        id: '9',
        url: 'https://web.dev/accessibility/',
        title: 'Web Accessibility Guidelines',
        description: 'Best practices for making web applications accessible to all users.',
        action: 'share',
        shareTo: 'Dev Blog',
        timestamp: '2024-01-14T11:00:00Z',
        domain: 'web.dev',
        age: '1d'
      },
      
      // Archived items
      {
        id: '10',
        url: 'https://devblogs.microsoft.com/typescript/announcing-typescript-5-0/',
        title: 'TypeScript 5.0 Release Notes',
        description: 'Discover the new features and improvements in TypeScript 5.0.',
        action: 'archived',
        timestamp: '2024-01-12T09:30:00Z',
        domain: 'microsoft.com',
        age: '3d'
      },
      {
        id: '11',
        url: 'https://nodejs.org/en/blog/announcements/v20-release-announce',
        title: 'Node.js 20 Release Announcement',
        description: 'What\'s new in Node.js version 20 with improved performance and features.',
        action: 'archived',
        timestamp: '2024-01-11T16:00:00Z',
        domain: 'nodejs.org',
        age: '4d'
      },
      {
        id: '12',
        url: 'https://webpack.js.org/guides/getting-started/',
        title: 'Webpack Getting Started Guide',
        description: 'Learn the basics of bundling JavaScript applications with Webpack.',
        action: 'archived',
        timestamp: '2024-01-10T13:20:00Z',
        domain: 'webpack.js.org',
        age: '5d'
      }
    ]
  }


  /**
   * Transform backend bookmark data to frontend Bookmark interface
   * Handles differences between backend and frontend data structures
   */
  private transformBackendBookmark(backendBookmark: BackendBookmarkResponse): Bookmark {
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
      tags: backendBookmark.tags || undefined,
      customProperties: backendBookmark.customProperties || undefined
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
      projectId: bookmark.project_id,
      tags: bookmark.tags,
      customProperties: bookmark.customProperties
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
      projectId: bookmark.project_id,
      tags: bookmark.tags,
      customProperties: bookmark.customProperties
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
      topic: bookmark.topic,
      tags: bookmark.tags,
      customProperties: bookmark.customProperties
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