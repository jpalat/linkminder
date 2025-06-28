export type BookmarkAction = 'read-later' | 'working' | 'share' | 'archived' | 'irrelevant'

export interface Bookmark {
  id: string
  url: string
  title: string
  description?: string
  content?: string
  action?: BookmarkAction
  shareTo?: string
  topic?: string
  project_id?: number
  timestamp: string
  domain?: string
  age?: string
  tags?: string[]
  customProperties?: Record<string, string>
}

export interface Project {
  id: number
  name: string
  description?: string
  status: 'active' | 'stale' | 'inactive'
  created_at: string
  updated_at: string
  linkCount?: number
  lastUpdated?: string
  progress?: number
}

export interface ProjectDetail {
  topic: string
  linkCount: number
  lastUpdated: string
  status: 'active' | 'stale' | 'inactive'
  progress?: number
  bookmarks: Bookmark[]
}

export interface FilterState {
  search?: string
  topic?: string
  domain?: string
  age?: string
  action?: string
}

export interface ShareGroup {
  destination: string
  items: Bookmark[]
  icon: string
  color: string
}

export interface DashboardStats {
  needsTriage: number
  activeProjects: number
  readyToShare: number
  totalBookmarks: number
  archived: number
  projectStats: ProjectStat[]
}

export interface ProjectStat {
  topic: string
  count: number
  lastUpdated: string
  status: 'active' | 'stale' | 'inactive'
  latestURL?: string
  latestTitle?: string
}

export type TabType = 'triage' | 'projects' | 'share' | 'archive'
