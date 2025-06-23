import { apiClient, type ApiResponse } from './api'
import type { Project, Bookmark, ProjectDetail } from '@/types'

// API response types matching the Go backend
export interface ProjectDetailResponse {
  topic: string
  linkCount: number
  lastUpdated: string
  status: 'active' | 'stale' | 'inactive'
  progress?: number
  bookmarks: BackendBookmark[]
}

export interface BackendBookmark {
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

export interface ProjectsResponse {
  projects: BackendProject[]
  referenceCollections?: BackendProject[]
}

export interface BackendProject {
  id: number
  name: string
  description?: string
  status: string
  linkCount: number
  lastUpdated: string
  createdAt: string
}

class ProjectService {
  /**
   * Get active projects and reference collections
   * GET /api/projects
   */
  async getProjects(): Promise<ProjectsResponse> {
    const response = await apiClient.get<ProjectsResponse>('/api/projects')
    return response.data
  }

  /**
   * Get detailed view of a specific project by topic name
   * GET /api/projects/{topic}
   */
  async getProjectByTopic(topic: string): Promise<ProjectDetailResponse> {
    const encodedTopic = encodeURIComponent(topic)
    const response = await apiClient.get<ProjectDetailResponse>(`/api/projects/${encodedTopic}`)
    return response.data
  }

  /**
   * Get project detail view - frontend-friendly version
   */
  async getProjectDetail(topicOrId: string): Promise<ProjectDetail> {
    // Check if it's a numeric ID or topic name
    const isNumericId = /^\d+$/.test(topicOrId)
    
    const response = isNumericId 
      ? await this.getProjectById(parseInt(topicOrId))
      : await this.getProjectByTopic(topicOrId)

    return {
      topic: response.topic,
      linkCount: response.linkCount,
      lastUpdated: response.lastUpdated,
      status: response.status,
      progress: response.progress,
      bookmarks: response.bookmarks.map(bookmark => this.transformBackendBookmark(bookmark))
    }
  }

  /**
   * Get detailed view of a project by ID
   * GET /api/projects/id/{id}
   */
  async getProjectById(id: number): Promise<ProjectDetailResponse> {
    const response = await apiClient.get<ProjectDetailResponse>(`/api/projects/id/${id}`)
    return response.data
  }

  /**
   * Transform backend project data to frontend Project interface
   */
  transformBackendProject(backendProject: BackendProject): Project {
    return {
      id: backendProject.id,
      name: backendProject.name,
      description: backendProject.description,
      status: this.mapProjectStatus(backendProject.status),
      created_at: backendProject.createdAt,
      updated_at: backendProject.lastUpdated,
      linkCount: backendProject.linkCount,
      lastUpdated: backendProject.lastUpdated,
      progress: this.calculateProgress(backendProject)
    }
  }

  /**
   * Transform backend bookmark data to frontend Bookmark interface
   */
  transformBackendBookmark(backendBookmark: BackendBookmark): Bookmark {
    return {
      id: String(backendBookmark.id),
      url: backendBookmark.url,
      title: backendBookmark.title,
      description: backendBookmark.description,
      content: backendBookmark.content,
      action: backendBookmark.action as any,
      timestamp: backendBookmark.timestamp,
      domain: backendBookmark.domain || this.extractDomain(backendBookmark.url),
      age: backendBookmark.age || this.calculateAge(backendBookmark.timestamp)
    }
  }

  /**
   * Transform project detail response for frontend consumption
   */
  transformProjectDetail(response: ProjectDetailResponse): {
    project: Partial<Project>
    bookmarks: Bookmark[]
  } {
    const project: Partial<Project> = {
      name: response.topic,
      status: response.status,
      linkCount: response.linkCount,
      lastUpdated: response.lastUpdated,
      progress: response.progress
    }

    const bookmarks = response.bookmarks.map(bookmark => 
      this.transformBackendBookmark(bookmark)
    )

    return { project, bookmarks }
  }

  /**
   * Map backend project status to frontend status
   */
  private mapProjectStatus(backendStatus: string): 'active' | 'stale' | 'inactive' {
    switch (backendStatus.toLowerCase()) {
      case 'active':
        return 'active'
      case 'stale':
        return 'stale'
      case 'inactive':
        return 'inactive'
      default:
        return 'active'
    }
  }

  /**
   * Calculate project progress based on bookmark actions
   */
  private calculateProgress(project: BackendProject): number {
    // This is a simple progress calculation
    // In a real implementation, this might be calculated on the backend
    if (project.linkCount === 0) return 0
    
    // Simple heuristic: assume some progress based on age and activity
    const daysSinceUpdate = this.getDaysSince(project.lastUpdated)
    
    if (daysSinceUpdate < 7) return 75 // Active project
    if (daysSinceUpdate < 30) return 50 // Stale project  
    return 25 // Inactive project
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
  private calculateAge(timestamp: string): string {
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
   * Get days since a given date
   */
  private getDaysSince(dateString: string): number {
    const now = new Date()
    const date = new Date(dateString)
    const diffMs = now.getTime() - date.getTime()
    return Math.floor(diffMs / (1000 * 60 * 60 * 24))
  }

  /**
   * Get project statistics for dashboard
   */
  async getProjectStats(): Promise<Array<{
    topic: string
    count: number
    lastUpdated: string
    status: 'active' | 'stale' | 'inactive'
  }>> {
    try {
      const response = await this.getProjects()
      
      return response.projects.map(project => ({
        topic: project.name,
        count: project.linkCount,
        lastUpdated: project.lastUpdated,
        status: this.mapProjectStatus(project.status)
      }))
    } catch (error) {
      console.error('Failed to fetch project stats:', error)
      return []
    }
  }

  /**
   * Search projects by name or description
   */
  async searchProjects(query: string): Promise<Project[]> {
    try {
      const response = await this.getProjects()
      const searchTerm = query.toLowerCase()
      
      const filteredProjects = response.projects.filter(project => 
        project.name.toLowerCase().includes(searchTerm) ||
        (project.description && project.description.toLowerCase().includes(searchTerm))
      )
      
      return filteredProjects.map(project => this.transformBackendProject(project))
    } catch (error) {
      console.error('Failed to search projects:', error)
      return []
    }
  }

  /**
   * Get bookmarks for a specific project topic with filtering
   */
  async getProjectBookmarks(
    topic: string, 
    filters?: {
      action?: string
      domain?: string
      search?: string
    }
  ): Promise<Bookmark[]> {
    try {
      const response = await this.getProjectByTopic(topic)
      let bookmarks = response.bookmarks.map(bookmark => 
        this.transformBackendBookmark(bookmark)
      )

      // Apply client-side filtering if filters are provided
      if (filters) {
        if (filters.action) {
          bookmarks = bookmarks.filter(bookmark => bookmark.action === filters.action)
        }
        
        if (filters.domain) {
          bookmarks = bookmarks.filter(bookmark => bookmark.domain === filters.domain)
        }
        
        if (filters.search) {
          const searchTerm = filters.search.toLowerCase()
          bookmarks = bookmarks.filter(bookmark =>
            bookmark.title.toLowerCase().includes(searchTerm) ||
            bookmark.url.toLowerCase().includes(searchTerm) ||
            (bookmark.description && bookmark.description.toLowerCase().includes(searchTerm))
          )
        }
      }

      return bookmarks
    } catch (error) {
      console.error('Failed to fetch project bookmarks:', error)
      return []
    }
  }

  /**
   * Get project metadata without bookmarks (lighter request)
   */
  async getProjectMetadata(topic: string): Promise<Partial<Project> | null> {
    try {
      const response = await this.getProjectByTopic(topic)
      
      return {
        name: response.topic,
        status: response.status,
        linkCount: response.linkCount,
        lastUpdated: response.lastUpdated,
        progress: response.progress
      }
    } catch (error) {
      console.error('Failed to fetch project metadata:', error)
      return null
    }
  }
}

export const projectService = new ProjectService()
export default projectService