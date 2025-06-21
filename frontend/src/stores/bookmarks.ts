import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Bookmark, FilterState, TabType, DashboardStats, Project } from '@/types'

export const useBookmarkStore = defineStore('bookmarks', () => {
  // State
  const bookmarks = ref<Bookmark[]>([])
  const projects = ref<Project[]>([])
  const filters = ref<FilterState>({})
  const selectedItems = ref(new Set<string>())
  const currentTab = ref<TabType>('triage')
  const batchMode = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

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
      // Implementation would depend on your age parsing logic
      // This is a simplified version
      filtered = filtered.filter(b => {
        if (!b.age) return false
        switch (filters.value.age) {
          case 'today':
            return b.age.includes('h')
          case 'yesterday':
            return b.age === '1d'
          case 'week':
            return b.age.includes('h') || b.age.includes('d')
          case 'month':
            return !b.age.includes('w') && !b.age.includes('m')
          case 'older':
            return b.age.includes('w') || b.age.includes('m')
          default:
            return true
        }
      })
    }

    return filtered
  })

  const dashboardStats = computed<DashboardStats>(() => {
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
      lastUpdated: new Date().toISOString(), // This would come from API
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

  const updateBookmark = (bookmarkId: string, updates: Partial<Bookmark>) => {
    const index = bookmarks.value.findIndex(b => b.id === bookmarkId)
    if (index !== -1) {
      bookmarks.value[index] = { ...bookmarks.value[index], ...updates }
    }
  }

  const moveBookmarks = (bookmarkIds: string[], action: string) => {
    bookmarkIds.forEach(id => {
      updateBookmark(id, { action: action as any })
    })
    clearSelection()
  }

  const loadBookmarks = async () => {
    loading.value = true
    error.value = null
    try {
      // This would be an actual API call
      // For now, we'll load mock data
      await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API delay
      bookmarks.value = getMockBookmarks()
    } catch (err) {
      error.value = 'Failed to load bookmarks'
      console.error('Error loading bookmarks:', err)
    } finally {
      loading.value = false
    }
  }

  const addBookmark = async (bookmark: Omit<Bookmark, 'id' | 'timestamp'>) => {
    try {
      // This would be an actual API call
      const newBookmark: Bookmark = {
        ...bookmark,
        id: Date.now().toString(),
        timestamp: new Date().toISOString(),
        age: 'just now',
        domain: extractDomain(bookmark.url)
      }
      bookmarks.value.unshift(newBookmark)
    } catch (err) {
      error.value = 'Failed to add bookmark'
      console.error('Error adding bookmark:', err)
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
    
    // Computed
    filteredBookmarks,
    dashboardStats,
    shareGroups,
    
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
    addBookmark
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

function extractDomain(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return 'unknown'
  }
}

function getMockBookmarks(): Bookmark[] {
  return [
    {
      id: '1',
      url: 'https://react.dev/blog/2022/03/29/react-v18',
      title: 'Building Modern Web Applications with React 18',
      description: 'Learn about the new features in React 18 including concurrent rendering, automatic batching, and more.',
      action: 'read-later',
      topic: 'react-migration',
      timestamp: '2024-01-15T10:30:00Z',
      domain: 'react.dev',
      age: '2h'
    },
    {
      id: '2',
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
      id: '3',
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
      id: '4',
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
      id: '5',
      url: 'https://devblogs.microsoft.com/typescript/announcing-typescript-5-0/',
      title: 'TypeScript 5.0 Release Notes',
      description: 'Discover the new features and improvements in TypeScript 5.0.',
      action: 'archived',
      timestamp: '2024-01-12T09:30:00Z',
      domain: 'microsoft.com',
      age: '3d'
    }
  ]
}
