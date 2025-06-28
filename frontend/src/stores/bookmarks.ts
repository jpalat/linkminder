import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Bookmark, FilterState, TabType, DashboardStats, Project } from '@/types'
import { bookmarkService } from '@/services/bookmarkService'
import { projectService } from '@/services/projectService'
// Remove unused imports
import { getErrorDisplayMessage, isNetworkError } from '@/composables/useApiError'
import { useNotifications } from '@/composables/useNotifications'

export const useBookmarkStore = defineStore('bookmarks', () => {
  // Initialize notifications
  const { bookmarkCreated, bookmarkUpdated, apiError, networkError, bulkOperation } = useNotifications()
  
  // State
  const bookmarks = ref<Bookmark[]>([])
  const projects = ref<Project[]>([])
  const filters = ref<FilterState>({})
  const selectedItems = ref(new Set<string>())
  const currentTab = ref<TabType>('triage')
  const batchMode = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const isConnected = ref(true) // Track API connectivity
  const currentSort = ref('date-desc')

  // Computed
  const filteredBookmarks = computed(() => {
    let filtered = bookmarks.value

    // Filter by current tab
    switch (currentTab.value) {
      case 'triage':
        filtered = filtered.filter(b => !b.action || b.action === 'read-later')
        break
      case 'projects':
        filtered = filtered.filter(b => b.action === 'working')
        break
      case 'share':
        filtered = filtered.filter(b => b.action === 'share')
        break
      case 'archive':
        filtered = filtered.filter(b => b.action === 'archived')
        break
    }

    // Apply search filter
    if (filters.value.search) {
      const searchTerm = filters.value.search.toLowerCase()
      filtered = filtered.filter(bookmark => 
        bookmark.title.toLowerCase().includes(searchTerm) ||
        bookmark.url.toLowerCase().includes(searchTerm) ||
        (bookmark.description && bookmark.description.toLowerCase().includes(searchTerm))
      )
    }

    // Apply topic filter
    if (filters.value.topic) {
      if (filters.value.topic === 'has-topic') {
        filtered = filtered.filter(b => b.topic)
      } else if (filters.value.topic === 'no-topic') {
        filtered = filtered.filter(b => !b.topic)
      } else {
        filtered = filtered.filter(b => b.topic === filters.value.topic)
      }
    }

    // Apply domain filter
    if (filters.value.domain) {
      filtered = filtered.filter(b => b.domain === filters.value.domain)
    }

    // Apply age filter
    if (filters.value.age) {
      const now = new Date()
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      const yesterday = new Date(today.getTime() - 24 * 60 * 60 * 1000)
      const weekAgo = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000)
      const monthAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000)
      
      filtered = filtered.filter(b => {
        const bookmarkDate = new Date(b.timestamp)
        switch (filters.value.age) {
          case 'today':
            return bookmarkDate >= today
          case 'yesterday':
            return bookmarkDate >= yesterday && bookmarkDate < today
          case 'week':
            return bookmarkDate >= weekAgo
          case 'month':
            return bookmarkDate >= monthAgo
          case 'older':
            return bookmarkDate < monthAgo
          default:
            return true
        }
      })
    }

    // Apply sorting
    filtered = applySorting(filtered, currentSort.value)
    
    return filtered
  })
  
  // Sorting function
  const applySorting = (bookmarks: Bookmark[], sortKey: string): Bookmark[] => {
    const sorted = [...bookmarks]
    
    switch (sortKey) {
      case 'date-desc':
        return sorted.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      case 'date-asc':
        return sorted.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
      case 'title-asc':
        return sorted.sort((a, b) => a.title.localeCompare(b.title))
      case 'title-desc':
        return sorted.sort((a, b) => b.title.localeCompare(a.title))
      case 'domain-asc':
        return sorted.sort((a, b) => (a.domain || '').localeCompare(b.domain || ''))
      case 'topic-asc':
        return sorted.sort((a, b) => (a.topic || 'zzz').localeCompare(b.topic || 'zzz'))
      default:
        return sorted
    }
  }

  // Store API-fetched dashboard stats
  const apiDashboardStats = ref<DashboardStats | null>(null)
  
  const dashboardStats = computed<DashboardStats>(() => {
    // If we have API stats, use them; otherwise fall back to computed stats
    if (apiDashboardStats.value) {
      return apiDashboardStats.value
    }
    
    // Fallback to local computation (for when API is unavailable)
    const needsTriage = bookmarks.value.filter(b => !b.action || b.action === 'read-later').length
    const activeProjects = new Set(bookmarks.value.filter(b => b.action === 'working').map(b => b.topic)).size
    const readyToShare = bookmarks.value.filter(b => b.action === 'share').length
    const archived = bookmarks.value.filter(b => b.action === 'archived').length
    const totalBookmarks = bookmarks.value.length

    // Generate project stats
    const projectCounts = new Map<string, number>()
    bookmarks.value.filter(b => b.action === 'working' && b.topic).forEach(b => {
      const count = projectCounts.get(b.topic!) || 0
      projectCounts.set(b.topic!, count + 1)
    })

    const projectStats = Array.from(projectCounts.entries()).map(([topic, count]) => ({
      topic,
      count,
      lastUpdated: new Date().toISOString(),
      status: 'active' as const
    }))

    return {
      needsTriage,
      activeProjects,
      readyToShare,
      totalBookmarks,
      archived,
      projectStats
    }
  })

  const shareGroups = computed(() => {
    const shareBookmarks = bookmarks.value.filter(b => b.action === 'share')
    const groups = new Map<string, Bookmark[]>()

    shareBookmarks.forEach(bookmark => {
      const destination = bookmark.shareTo || 'Unassigned'
      if (!groups.has(destination)) {
        groups.set(destination, [])
      }
      groups.get(destination)!.push(bookmark)
    })

    return Array.from(groups.entries()).map(([destination, items]) => ({
      destination,
      items,
      icon: getDestinationIcon(destination),
      color: getDestinationColor(destination)
    }))
  })

  const availableTopics = computed(() => {
    const topics = new Set<string>()
    bookmarks.value.forEach(b => {
      if (b.topic) {
        topics.add(b.topic)
      }
    })
    return Array.from(topics).sort()
  })

  const availableDomains = computed(() => {
    const domains = new Set<string>()
    bookmarks.value.forEach(b => {
      if (b.domain) {
        domains.add(b.domain)
      }
    })
    return Array.from(domains).sort()
  })

  const availableShareDestinations = computed(() => {
    const destinations = new Set<string>()
    bookmarks.value.forEach(b => {
      if (b.shareTo && b.shareTo.trim() !== '') {
        destinations.add(b.shareTo)
      }
    })
    return Array.from(destinations).sort()
  })

  // Actions
  const updateFilters = (newFilters: Partial<FilterState>) => {
    filters.value = { ...filters.value, ...newFilters }
  }

  const clearFilters = () => {
    filters.value = {}
  }

  const setCurrentTab = (tab: TabType) => {
    currentTab.value = tab
    clearFilters() // Clear filters when switching tabs
  }

  const toggleSelection = (bookmarkId: string) => {
    if (selectedItems.value.has(bookmarkId)) {
      selectedItems.value.delete(bookmarkId)
    } else {
      selectedItems.value.add(bookmarkId)
    }
  }

  const clearSelection = () => {
    selectedItems.value.clear()
  }

  const toggleBatchMode = () => {
    batchMode.value = !batchMode.value
    if (!batchMode.value) {
      clearSelection()
    }
  }

  const updateBookmark = async (bookmarkId: string, updates: Partial<Bookmark>) => {
    try {
      // Determine if we need full update (PUT) or partial update (PATCH)
      const needsFullUpdate = updates.title !== undefined || updates.url !== undefined || 'description' in updates
      
      let updatedBookmark: Bookmark
      
      if (needsFullUpdate) {
        // Get the current bookmark to merge with updates
        const currentBookmark = bookmarks.value.find(b => b.id === bookmarkId)
        if (!currentBookmark) {
          throw new Error(`Bookmark with ID ${bookmarkId} not found`)
        }
        
        // Merge current bookmark with updates for full update
        const fullBookmark: Bookmark = {
          ...currentBookmark,
          ...updates
        }
        
        const backendRequest = bookmarkService.toBackendFullUpdateRequest(fullBookmark)
        updatedBookmark = await bookmarkService.updateBookmarkFull(bookmarkId, backendRequest)
      } else {
        // Use partial update for action/topic changes only
        const backendUpdates = bookmarkService.toBackendUpdateRequest(updates)
        updatedBookmark = await bookmarkService.updateBookmark(bookmarkId, backendUpdates)
      }
      
      // Update local state
      const index = bookmarks.value.findIndex(b => b.id === bookmarkId)
      if (index !== -1) {
        bookmarks.value[index] = { ...bookmarks.value[index], ...updatedBookmark }
      }
      
      // Show success notification
      bookmarkUpdated(updatedBookmark.title)
      console.log('Successfully updated bookmark:', updatedBookmark)
    } catch (err) {
      error.value = getErrorDisplayMessage(err)
      console.error('Error updating bookmark:', err)
      
      // Show error notification
      apiError('update bookmark', err as Error)
      
      if (isNetworkError(err)) {
        isConnected.value = false
      }
      
      throw err // Re-throw for component error handling
    }
  }

  const moveBookmarks = async (bookmarkIds: string[], action: string) => {
    try {
      // Update each bookmark
      const promises = bookmarkIds.map(id => updateBookmark(id, { action }))
      await Promise.all(promises)
      
      // Show bulk operation notification
      const count = bookmarkIds.length
      if (count > 1) {
        const actionLabels: Record<string, string> = {
          'working': 'moved to working',
          'share': 'marked for sharing', 
          'archived': 'archived',
          'read-later': 'moved to triage'
        }
        const operation = actionLabels[action] || `updated to ${action}`
        bulkOperation(count, operation)
      }
      
      clearSelection()
    } catch (err) {
      console.error('Error in bulk move operation:', err)
      apiError('move bookmarks', err as Error)
    }
  }

  const loadBookmarks = async () => {
    loading.value = true
    error.value = null
    try {
      // Load ALL bookmarks (triage + share + working + archived)
      const allBookmarks = await bookmarkService.getAllBookmarks()
      
      // Load project data
      const projectsResponse = await projectService.getProjects()
      
      // Use the complete bookmark dataset
      bookmarks.value = allBookmarks
      
      // Update projects
      projects.value = projectsResponse.projects?.map(project => 
        projectService.transformBackendProject(project)
      ) || []
      
    } catch (err) {
      error.value = getErrorDisplayMessage(err)
      console.error('Error loading bookmarks:', err)
      
      // Check if it's a network error
      if (isNetworkError(err)) {
        isConnected.value = false
        error.value = 'Unable to connect to server. Using offline mode.'
      }
      
      // Fallback to mock data if API fails
      console.log('Falling back to mock data')
      bookmarks.value = getMockBookmarks()
    } finally {
      loading.value = false
    }
  }

  const addBookmark = async (bookmark: Omit<Bookmark, 'id' | 'timestamp'>) => {
    try {
      // Convert to backend format and create via API
      const backendRequest = bookmarkService.toBackendCreateRequest(bookmark)
      const newBookmark = await bookmarkService.createBookmark(backendRequest)
      
      // Add to local state
      bookmarks.value.unshift(newBookmark)
      
      // Show success notification
      bookmarkCreated(newBookmark.title)
      console.log('Successfully added bookmark:', newBookmark)
    } catch (err) {
      error.value = getErrorDisplayMessage(err)
      console.error('Error adding bookmark:', err)
      
      // Show error notification
      apiError('create bookmark', err as Error)
      
      if (isNetworkError(err)) {
        isConnected.value = false
      }
      
      throw err // Re-throw for component error handling
    }
  }
  
  const setSortOrder = (sortKey: string) => {
    currentSort.value = sortKey
  }

  // Load dashboard stats from API
  const loadDashboardStats = async () => {
    try {
      apiDashboardStats.value = await bookmarkService.getDashboardStats()
    } catch (err) {
      console.error('Failed to load dashboard stats:', err)
      // Only show network errors, not API errors (stats are optional)
      if (isNetworkError(err)) {
        networkError()
      }
      // Keep using computed stats as fallback
    }
  }
  

  return {
    // State
    bookmarks,
    projects,
    filters,
    selectedItems,
    currentTab,
    batchMode,
    loading,
    error,
    isConnected,
    currentSort,
    
    // Computed
    filteredBookmarks,
    dashboardStats,
    shareGroups,
    availableTopics,
    availableDomains,
    availableShareDestinations,
    
    // Actions
    updateFilters,
    clearFilters,
    setCurrentTab,
    toggleSelection,
    clearSelection,
    toggleBatchMode,
    updateBookmark,
    moveBookmarks,
    loadBookmarks,
    addBookmark,
    setSortOrder,
    loadDashboardStats
  }
})

// Helper functions
function getDestinationIcon(destination: string): string {
  const icons: Record<string, string> = {
    'Team Slack': '💬',
    'Newsletter': '📧',
    'Dev Blog': '📝',
    'Unassigned': '📤'
  }
  return icons[destination] || '📤'
}

function getDestinationColor(destination: string): string {
  const colors: Record<string, string> = {
    'Team Slack': '#0f9f0f',
    'Newsletter': '#1e40af',
    'Dev Blog': '#d97706',
    'Unassigned': '#6b7280'
  }
  return colors[destination] || '#6b7280'
}


function getMockBookmarks(): Bookmark[] {
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
