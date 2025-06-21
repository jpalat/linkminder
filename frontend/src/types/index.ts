export interface Bookmark {
  id: string
  url: string
  title: string
  description?: string
  content?: string
  action?: 'read-later' | 'working' | 'share' | 'archived' | 'irrelevant'
  shareTo?: string
  topic?: string
  project_id?: number
  timestamp: string
  domain?: string
  age?: string
  tags?: string[]
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
}

export type TabType = 'triage' | 'projects' | 'share' | 'archive'
