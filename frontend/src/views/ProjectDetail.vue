<template>
  <div class="project-detail">
    <!-- Header with Navigation and Actions -->
    <header class="header">
      <div class="header-content">
        <div class="breadcrumb">
          <router-link to="/" class="breadcrumb-link">Dashboard</router-link>
          <span class="breadcrumb-separator">></span>
          <span class="breadcrumb-current">{{ projectData?.topic || 'Loading...' }}</span>
        </div>
        
        <div class="header-actions">
          <AppButton variant="secondary" @click="$router.back()">
            ← Back
          </AppButton>
          <AppButton variant="secondary" @click="showSettingsModal = true">
            ⚙️ Settings
          </AppButton>
          <AppButton variant="primary" @click="exportProject">
            📤 Export
          </AppButton>
        </div>
      </div>
    </header>
    
    <!-- Project Header with Stats -->
    <div class="project-header" v-if="projectData">
      <div class="container">
        <div class="project-info">
          <div class="project-title">
            <h1>📁 {{ projectData.topic }}</h1>
            <AppBadge 
              :variant="getStatusVariant(projectData.status)"
              class="status-badge"
            >
              {{ projectData.status }}
            </AppBadge>
          </div>
          
          <div class="project-stats">
            <div class="stat-item">
              <span class="stat-value">{{ projectData.linkCount }}</span>
              <span class="stat-label">Bookmarks</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ formatDate(projectData.lastUpdated) }}</span>
              <span class="stat-label">Last Updated</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ filteredBookmarks.length }}</span>
              <span class="stat-label">Visible</span>
            </div>
          </div>
        </div>
        
        <div class="project-actions">
          <AppButton 
            variant="primary" 
            @click="showAddBookmarkModal = true"
            class="add-bookmark-btn"
          >
            ➕ Add Bookmark
          </AppButton>
        </div>
      </div>
    </div>
    
    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Loading project data...</p>
    </div>
    
    <!-- Error State -->
    <div v-else-if="error" class="error-container">
      <div class="error-message">
        <h3>⚠️ Error Loading Project</h3>
        <p>{{ error }}</p>
        <AppButton @click="loadProjectData" variant="primary">
          🔄 Retry
        </AppButton>
      </div>
    </div>
    
    <!-- Main Content -->
    <main class="main-content" v-else-if="projectData">
      <div class="container">
        <!-- Filter Controls -->
        <div class="filters-section">
          <div class="filter-row">
            <!-- Search Input -->
            <div class="search-filter">
              <AppInput
                v-model="filters.search"
                type="search"
                placeholder="Search titles, descriptions, URLs..."
                class="search-input"
              >
                <template #prepend>
                  <span class="search-icon">🔍</span>
                </template>
              </AppInput>
            </div>
            
            <!-- Action Filter -->
            <div class="action-filter">
              <select v-model="filters.action" class="filter-select">
                <option value="">All Actions</option>
                <option value="working">Working</option>
                <option value="share">Share</option>
                <option value="read-later">Read Later</option>
                <option value="archived">Archived</option>
                <option value="irrelevant">Irrelevant</option>
              </select>
            </div>
            
            <!-- Domain Filter -->
            <div class="domain-filter">
              <select v-model="filters.domain" class="filter-select">
                <option value="">All Domains</option>
                <option v-for="domain in availableDomains" :key="domain" :value="domain">
                  {{ domain }}
                </option>
              </select>
            </div>
            
            <!-- Sort Control -->
            <div class="sort-control">
              <select v-model="sortBy" class="filter-select">
                <option value="timestamp-desc">Newest First</option>
                <option value="timestamp-asc">Oldest First</option>
                <option value="title-asc">Title A-Z</option>
                <option value="title-desc">Title Z-A</option>
                <option value="domain-asc">Domain A-Z</option>
                <option value="action-asc">Action A-Z</option>
              </select>
            </div>
          </div>
          
          <!-- Date Range Filters -->
          <div class="date-filters">
            <div class="date-filter">
              <label for="date-from">From:</label>
              <input 
                id="date-from"
                v-model="filters.dateFrom" 
                type="date" 
                class="date-input"
              >
            </div>
            <div class="date-filter">
              <label for="date-to">To:</label>
              <input 
                id="date-to"
                v-model="filters.dateTo" 
                type="date" 
                class="date-input"
              >
            </div>
            
            <!-- Clear Filters -->
            <AppButton 
              v-if="hasActiveFilters"
              variant="secondary" 
              @click="clearAllFilters"
              class="clear-filters-btn"
            >
              🗑️ Clear Filters
            </AppButton>
          </div>
          
          <!-- Active Filters Summary -->
          <div v-if="hasActiveFilters" class="active-filters">
            <span class="filter-label">Active filters:</span>
            <div class="filter-tags">
              <span v-if="filters.search" class="filter-tag">
                Search: "{{ filters.search }}"
                <button @click="filters.search = ''" class="filter-tag-remove">×</button>
              </span>
              <span v-if="filters.action" class="filter-tag">
                Action: {{ filters.action }}
                <button @click="filters.action = ''" class="filter-tag-remove">×</button>
              </span>
              <span v-if="filters.domain" class="filter-tag">
                Domain: {{ filters.domain }}
                <button @click="filters.domain = ''" class="filter-tag-remove">×</button>
              </span>
              <span v-if="filters.dateFrom" class="filter-tag">
                From: {{ filters.dateFrom }}
                <button @click="filters.dateFrom = ''" class="filter-tag-remove">×</button>
              </span>
              <span v-if="filters.dateTo" class="filter-tag">
                To: {{ filters.dateTo }}
                <button @click="filters.dateTo = ''" class="filter-tag-remove">×</button>
              </span>
            </div>
          </div>
        </div>
        
        <!-- Results Summary -->
        <div class="results-summary">
          <div class="results-info">
            <span class="results-count">
              Showing {{ filteredBookmarks.length }} of {{ projectData.linkCount }} bookmarks
            </span>
            <span v-if="hasActiveFilters" class="filtered-notice">
              (filtered)
            </span>
          </div>
          
          <!-- Bulk Actions -->
          <div class="bulk-actions" v-if="selectedBookmarks.size > 0">
            <span class="selection-count">{{ selectedBookmarks.size }} selected</span>
            <AppButton @click="bulkAction('working')" variant="secondary" size="sm">
              📝 Working
            </AppButton>
            <AppButton @click="bulkAction('share')" variant="secondary" size="sm">
              📤 Share
            </AppButton>
            <AppButton @click="bulkAction('archived')" variant="secondary" size="sm">
              📁 Archive
            </AppButton>
            <AppButton @click="bulkDelete" variant="danger" size="sm">
              🗑️ Delete
            </AppButton>
            <AppButton @click="selectedBookmarks.clear()" variant="secondary" size="sm">
              Clear
            </AppButton>
          </div>
        </div>
        
        <!-- Bookmark List -->
        <div class="bookmarks-section">
          <div v-if="filteredBookmarks.length === 0" class="empty-state">
            <div class="empty-icon">📚</div>
            <h3>No bookmarks found</h3>
            <p v-if="hasActiveFilters">
              Try adjusting your filters or 
              <button @click="clearAllFilters" class="link-button">clear all filters</button>
            </p>
            <p v-else>
              This project doesn't have any bookmarks yet.
            </p>
            <AppButton 
              @click="showAddBookmarkModal = true" 
              variant="primary"
              class="add-first-bookmark"
            >
              ➕ Add First Bookmark
            </AppButton>
          </div>
          
          <div v-else class="bookmark-list">
            <div 
              v-for="bookmark in filteredBookmarks" 
              :key="bookmark.id"
              class="bookmark-item"
              :class="{ 
                'bookmark-selected': selectedBookmarks.has(bookmark.id)
              }"
            >
              <!-- Selection Checkbox -->
              <div class="bookmark-checkbox">
                <input 
                  type="checkbox" 
                  :checked="selectedBookmarks.has(bookmark.id)"
                  @change="toggleBookmarkSelection(bookmark.id)"
                  class="checkbox-input"
                >
              </div>
              
              <!-- Bookmark Content -->
              <div class="bookmark-content">
                <div class="bookmark-header">
                  <h3 class="bookmark-title">{{ bookmark.title }}</h3>
                  <div class="bookmark-actions">
                    <AppBadge 
                      :variant="getActionVariant(bookmark.action)"
                      class="action-badge"
                    >
                      {{ bookmark.action || 'read-later' }}
                    </AppBadge>
                    <button 
                      @click.stop="editBookmark(bookmark)"
                      class="action-btn edit-btn"
                      title="Edit bookmark"
                    >
                      ✏️
                    </button>
                    <button 
                      @click.stop="deleteBookmark(bookmark)"
                      class="action-btn delete-btn"
                      title="Delete bookmark"
                    >
                      🗑️
                    </button>
                  </div>
                </div>
                
                <div class="bookmark-url">
                  <a :href="bookmark.url" target="_blank" @click.stop>
                    {{ bookmark.url }}
                  </a>
                </div>
                
                <div v-if="bookmark.description" class="bookmark-description">
                  {{ bookmark.description }}
                </div>
                
                <div class="bookmark-meta">
                  <span class="meta-item">
                    🌐 {{ bookmark.domain }}
                  </span>
                  <span class="meta-item">
                    📅 {{ formatDate(bookmark.timestamp) }}
                  </span>
                  <span class="meta-item">
                    ⏰ {{ bookmark.age }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
    
    <!-- Modals -->
    <AddBookmarkModal 
      v-if="showAddBookmarkModal"
      :show="showAddBookmarkModal"
      @close="showAddBookmarkModal = false"
      @submit="handleAddBookmark"
      :existing-topics="[projectData?.topic || ''].filter(Boolean)"
    />
    
    <EditBookmarkModal 
      v-if="editingBookmark"
      :show="!!editingBookmark"
      :bookmark="editingBookmark"
      @close="editingBookmark = null"
      @submit="handleEditBookmark"
      :existing-topics="[projectData?.topic || ''].filter(Boolean)"
    />
    
    
    <ProjectSettingsModal
      v-if="showSettingsModal && projectData"
      :show="showSettingsModal"
      :project="projectData"
      @close="showSettingsModal = false"
      @save="handleProjectSave"
      @export="handleProjectExport"
      @archive="handleProjectArchive"
      @delete="handleProjectDelete"
    />
    
    <ConfirmModal 
      v-if="confirmAction"
      :show="!!confirmAction"
      :config="{
        type: 'custom',
        title: confirmAction.title,
        message: confirmAction.message,
        confirmText: confirmAction.confirmText,
        isDestructive: confirmAction.variant === 'danger'
      }"
      @close="confirmAction = null"
      @confirm="handleConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBookmarkStore } from '@/stores/bookmarks'
import { projectService } from '@/services/projectService'
import type { Bookmark, ProjectDetail } from '@/types'
import AppButton from '@/components/ui/AppButton.vue'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppInput from '@/components/ui/AppInput.vue'
import AddBookmarkModal from '@/components/modals/AddBookmarkModal.vue'
import EditBookmarkModal from '@/components/modals/EditBookmarkModal.vue'
import ConfirmModal from '@/components/modals/ConfirmModal.vue'
import ProjectSettingsModal from '@/components/modals/ProjectSettingsModal.vue'

const route = useRoute()
const router = useRouter()
const bookmarkStore = useBookmarkStore()

// State
const projectData = ref<ProjectDetail | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

// Modal states
const showAddBookmarkModal = ref(false)
const showSettingsModal = ref(false)
const editingBookmark = ref<Bookmark | null>(null)
const confirmAction = ref<{
  title: string
  message: string
  confirmText: string
  variant: 'danger' | 'primary'
  action: () => void
} | null>(null)

// Filters
const filters = ref({
  search: '',
  action: '',
  domain: '',
  dateFrom: '',
  dateTo: ''
})

const sortBy = ref('timestamp-desc')
const selectedBookmarks = ref(new Set<string>())

// Computed
const projectId = computed(() => route.params.id as string)

const availableDomains = computed(() => {
  if (!projectData.value?.bookmarks) return []
  const domains = new Set(projectData.value.bookmarks.map(b => b.domain).filter(Boolean))
  return Array.from(domains).sort()
})

const filteredBookmarks = computed(() => {
  if (!projectData.value?.bookmarks) return []
  
  let filtered = [...projectData.value.bookmarks]
  
  // Apply filters
  if (filters.value.search) {
    const searchTerm = filters.value.search.toLowerCase()
    filtered = filtered.filter(bookmark => 
      bookmark.title.toLowerCase().includes(searchTerm) ||
      bookmark.url.toLowerCase().includes(searchTerm) ||
      (bookmark.description && bookmark.description.toLowerCase().includes(searchTerm))
    )
  }
  
  if (filters.value.action) {
    filtered = filtered.filter(b => b.action === filters.value.action)
  }
  
  if (filters.value.domain) {
    filtered = filtered.filter(b => b.domain === filters.value.domain)
  }
  
  if (filters.value.dateFrom) {
    const fromDate = new Date(filters.value.dateFrom)
    filtered = filtered.filter(b => new Date(b.timestamp) >= fromDate)
  }
  
  if (filters.value.dateTo) {
    const toDate = new Date(filters.value.dateTo + 'T23:59:59')
    filtered = filtered.filter(b => new Date(b.timestamp) <= toDate)
  }
  
  // Apply sorting
  filtered = applySorting(filtered, sortBy.value)
  
  return filtered
})

const hasActiveFilters = computed(() => {
  return filters.value.search ||
         filters.value.action ||
         filters.value.domain ||
         filters.value.dateFrom ||
         filters.value.dateTo
})

// Functions
const loadProjectData = async () => {
  loading.value = true
  error.value = null
  
  try {
    projectData.value = await projectService.getProjectDetail(projectId.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load project data'
    console.error('Error loading project data:', err)
  } finally {
    loading.value = false
  }
}

const applySorting = (bookmarks: Bookmark[], sortKey: string): Bookmark[] => {
  const sorted = [...bookmarks]
  
  switch (sortKey) {
    case 'timestamp-desc':
      return sorted.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    case 'timestamp-asc':
      return sorted.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
    case 'title-asc':
      return sorted.sort((a, b) => a.title.localeCompare(b.title))
    case 'title-desc':
      return sorted.sort((a, b) => b.title.localeCompare(a.title))
    case 'domain-asc':
      return sorted.sort((a, b) => (a.domain || '').localeCompare(b.domain || ''))
    case 'action-asc':
      return sorted.sort((a, b) => (a.action || 'read-later').localeCompare(b.action || 'read-later'))
    default:
      return sorted
  }
}

const clearAllFilters = () => {
  filters.value = {
    search: '',
    action: '',
    domain: '',
    dateFrom: '',
    dateTo: ''
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const getStatusVariant = (status: string): 'primary' | 'success' | 'danger' | 'info' | 'default' | 'warning' => {
  switch (status) {
    case 'active': return 'success'
    case 'stale': return 'warning'
    case 'inactive': return 'default'
    default: return 'default'
  }
}

const getActionVariant = (action: string | undefined): 'primary' | 'success' | 'danger' | 'info' | 'default' | 'warning' => {
  switch (action) {
    case 'working': return 'primary'
    case 'share': return 'success'
    case 'archived': return 'default'
    case 'irrelevant': return 'warning'
    default: return 'default'
  }
}

// Bookmark selection
const toggleBookmarkSelection = (bookmarkId: string) => {
  if (selectedBookmarks.value.has(bookmarkId)) {
    selectedBookmarks.value.delete(bookmarkId)
  } else {
    selectedBookmarks.value.add(bookmarkId)
  }
}

// Modal handlers
const editBookmark = (bookmark: Bookmark) => {
  editingBookmark.value = bookmark
}

const deleteBookmark = (bookmark: Bookmark) => {
  confirmAction.value = {
    title: 'Delete Bookmark',
    message: `Are you sure you want to delete "${bookmark.title}"? This action cannot be undone.`,
    confirmText: 'Delete',
    variant: 'danger',
    action: () => performDeleteBookmark(bookmark.id)
  }
}

const bulkAction = async (action: string) => {
  const bookmarkIds = Array.from(selectedBookmarks.value)
  try {
    await Promise.all(bookmarkIds.map(id => 
      bookmarkStore.updateBookmark(id, { action })
    ))
    selectedBookmarks.value.clear()
    await loadProjectData()
  } catch (err) {
    error.value = 'Failed to update bookmarks'
    console.error('Bulk action error:', err)
  }
}

const bulkDelete = () => {
  const count = selectedBookmarks.value.size
  confirmAction.value = {
    title: 'Delete Bookmarks',
    message: `Are you sure you want to delete ${count} bookmark${count > 1 ? 's' : ''}? This action cannot be undone.`,
    confirmText: 'Delete All',
    variant: 'danger',
    action: () => performBulkDelete()
  }
}

const performBulkDelete = async () => {
  const bookmarkIds = Array.from(selectedBookmarks.value)
  try {
    await Promise.all(bookmarkIds.map(id => 
      console.log('Would delete bookmark:', id)
    ))
    selectedBookmarks.value.clear()
    await loadProjectData()
  } catch (err) {
    error.value = 'Failed to delete bookmarks'
    console.error('Bulk delete error:', err)
  }
}

const performDeleteBookmark = async (bookmarkId: string) => {
  try {
    console.log('Would delete bookmark:', bookmarkId)
    await loadProjectData()
  } catch (err) {
    error.value = 'Failed to delete bookmark'
    console.error('Delete error:', err)
  }
}

const handleAddBookmark = async (bookmark: Omit<Bookmark, 'id' | 'timestamp'>) => {
  try {
    await bookmarkStore.addBookmark({
      ...bookmark,
      topic: projectData.value?.topic
    })
    showAddBookmarkModal.value = false
    await loadProjectData()
  } catch (err) {
    error.value = 'Failed to add bookmark'
    console.error('Add bookmark error:', err)
  }
}

const handleEditBookmark = async (updatedBookmark: Bookmark) => {
  try {
    await bookmarkStore.updateBookmark(updatedBookmark.id, updatedBookmark)
    editingBookmark.value = null
    await loadProjectData()
  } catch (err) {
    error.value = 'Failed to update bookmark'
    console.error('Edit bookmark error:', err)
  }
}

const handleConfirm = () => {
  if (confirmAction.value) {
    confirmAction.value.action()
    confirmAction.value = null
  }
}

// Project management handlers
const handleProjectSave = async (updates: Partial<ProjectDetail>) => {
  try {
    console.log('Would save project updates:', updates)
    showSettingsModal.value = false
    await loadProjectData()
  } catch (err) {
    error.value = 'Failed to update project'
    console.error('Project save error:', err)
  }
}

const handleProjectExport = (project: ProjectDetail) => {
  console.log('Would export project:', project)
  showSettingsModal.value = false
}

const handleProjectArchive = (project: ProjectDetail) => {
  confirmAction.value = {
    title: 'Archive Project',
    message: `Are you sure you want to archive "${project.topic}"? Archived projects can be restored later.`,
    confirmText: 'Archive',
    variant: 'primary',
    action: () => performProjectArchive(project)
  }
  showSettingsModal.value = false
}

const handleProjectDelete = (project: ProjectDetail) => {
  confirmAction.value = {
    title: 'Delete Project',
    message: `Are you sure you want to delete "${project.topic}"? This will permanently delete the project and all its bookmarks. This action cannot be undone.`,
    confirmText: 'Delete Forever',
    variant: 'danger',
    action: () => performProjectDelete(project)
  }
  showSettingsModal.value = false
}

const performProjectArchive = async (project: ProjectDetail) => {
  try {
    console.log('Would archive project:', project)
    router.push('/')
  } catch (err) {
    error.value = 'Failed to archive project'
    console.error('Project archive error:', err)
  }
}

const performProjectDelete = async (project: ProjectDetail) => {
  try {
    console.log('Would delete project:', project)
    router.push('/')
  } catch (err) {
    error.value = 'Failed to delete project'
    console.error('Project delete error:', err)
  }
}

const exportProject = () => {
  console.log('Export project:', projectData.value)
}

// Watch for route changes
watch(() => route.params.id, () => {
  if (route.params.id) {
    loadProjectData()
  }
})

// Load data on mount
onMounted(() => {
  if (projectId.value) {
    loadProjectData()
  }
})
</script>

<style scoped>
.project-detail {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-gray-50);
}

.header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-lg) var(--spacing-xl);
  box-shadow: var(--shadow-sm);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: var(--color-gray-600);
  font-size: var(--font-size-base);
}

.breadcrumb-link {
  color: var(--color-primary);
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.breadcrumb-separator {
  color: var(--color-gray-400);
}

.breadcrumb-current {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
}

.header-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* Project Header */
.project-header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-xl);
}

.project-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.project-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.project-title h1 {
  margin: 0;
  font-size: var(--font-size-2xl);
  color: var(--color-gray-900);
}

.status-badge {
  font-size: var(--font-size-sm);
}

.project-stats {
  display: flex;
  gap: var(--spacing-xl);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.stat-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-gray-900);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  margin-top: var(--spacing-xs);
}

.project-actions {
  display: flex;
  align-items: center;
  margin-top: var(--spacing-lg);
}

/* Loading and Error States */
.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-4xl);
  text-align: center;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--color-gray-200);
  border-top: 4px solid var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: var(--spacing-lg);
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-message {
  background: white;
  padding: var(--spacing-xl);
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--color-red-200);
  max-width: 500px;
}

/* Main Content */
.main-content {
  flex: 1;
  padding: var(--spacing-xl);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}

/* Filters Section */
.filters-section {
  background: white;
  padding: var(--spacing-lg);
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-light);
  margin-bottom: var(--spacing-lg);
}

.filter-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: var(--spacing-md);
  align-items: end;
}

.search-filter {
  position: relative;
}

.search-input {
  width: 100%;
}

.search-icon {
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-sm);
  color: var(--color-gray-500);
}

.filter-select {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--border-radius);
  background: white;
  font-size: var(--font-size-base);
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-light);
}

.date-filters {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--border-light);
}

.date-filter {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.date-filter label {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  white-space: nowrap;
}

.date-input {
  padding: var(--spacing-sm);
  border: 1px solid var(--border-light);
  border-radius: var(--border-radius);
  font-size: var(--font-size-sm);
}

.clear-filters-btn {
  margin-left: auto;
}

/* Active Filters */
.active-filters {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--border-light);
}

.filter-label {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  margin-right: var(--spacing-md);
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  align-items: center;
}

.filter-tag {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  background: var(--color-primary-light);
  color: var(--color-primary-dark);
  border-radius: var(--border-radius);
  font-size: var(--font-size-sm);
}

.filter-tag-remove {
  background: none;
  border: none;
  color: var(--color-primary-dark);
  cursor: pointer;
  padding: 0;
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-weight: bold;
}

.filter-tag-remove:hover {
  background: var(--color-primary);
  color: white;
}

/* Results Summary */
.results-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  padding: var(--spacing-md) 0;
}

.results-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.results-count {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
}

.filtered-notice {
  color: var(--color-gray-600);
  font-size: var(--font-size-sm);
}

.bulk-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.selection-count {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  margin-right: var(--spacing-md);
}

/* Bookmarks Section */
.bookmarks-section {
  background: white;
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-light);
  overflow: hidden;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-4xl);
  text-align: center;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: var(--spacing-lg);
}

.empty-state h3 {
  margin: 0 0 var(--spacing-md) 0;
  color: var(--color-gray-900);
}

.empty-state p {
  color: var(--color-gray-600);
  margin-bottom: var(--spacing-lg);
}

.link-button {
  background: none;
  border: none;
  color: var(--color-primary);
  text-decoration: underline;
  cursor: pointer;
  font-size: inherit;
}

.link-button:hover {
  color: var(--color-primary-dark);
}

.add-first-bookmark {
  margin-top: var(--spacing-lg);
}

/* Bookmark List */
.bookmark-list {
  divide-y: 1px solid var(--border-light);
}

.bookmark-item {
  display: flex;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  transition: all 0.15s ease;
  border-bottom: 1px solid var(--border-light);
}

.bookmark-item:last-child {
  border-bottom: none;
}

.bookmark-item:hover {
  background: var(--color-gray-50);
}

.bookmark-selected {
  background: var(--color-primary-light) !important;
  border-left: 4px solid var(--color-primary);
}

.bookmark-preview {
  background: var(--color-blue-50) !important;
  border-left: 4px solid var(--color-blue-500);
}

.bookmark-checkbox {
  display: flex;
  align-items: flex-start;
  padding-top: var(--spacing-xs);
}

.checkbox-input {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.bookmark-content {
  flex: 1;
  cursor: pointer;
}

.bookmark-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-sm);
}

.bookmark-title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1.4;
}

.bookmark-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.action-badge {
  font-size: var(--font-size-xs);
  white-space: nowrap;
}

.action-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: var(--spacing-xs);
  border-radius: var(--border-radius);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  transition: background-color 0.15s ease;
}

.action-btn:hover {
  background: var(--color-gray-200);
}

.delete-btn:hover {
  background: var(--color-red-100);
}

.bookmark-url {
  margin-bottom: var(--spacing-sm);
}

.bookmark-url a {
  color: var(--color-primary);
  text-decoration: none;
  font-size: var(--font-size-sm);
  word-break: break-all;
}

.bookmark-url a:hover {
  text-decoration: underline;
}

.bookmark-description {
  color: var(--color-gray-700);
  font-size: var(--font-size-base);
  line-height: 1.5;
  margin-bottom: var(--spacing-sm);
}

.bookmark-meta {
  display: flex;
  gap: var(--spacing-lg);
  color: var(--color-gray-500);
  font-size: var(--font-size-sm);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

/* Responsive Design */
@media (max-width: 768px) {
  .header {
    padding: var(--spacing-md);
  }
  
  .header-content {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .header-actions {
    justify-content: center;
  }
  
  .project-header {
    padding: var(--spacing-lg) var(--spacing-md);
  }
  
  .project-info {
    align-items: center;
    text-align: center;
  }
  
  .project-title {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .project-stats {
    justify-content: center;
    gap: var(--spacing-lg);
  }
  
  .main-content {
    padding: var(--spacing-md);
  }
  
  .filter-row {
    grid-template-columns: 1fr;
    gap: var(--spacing-sm);
  }
  
  .date-filters {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-sm);
  }
  
  .date-filter {
    justify-content: space-between;
  }
  
  .results-summary {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .bulk-actions {
    justify-content: center;
    flex-wrap: wrap;
  }
  
  .bookmark-item {
    padding: var(--spacing-md);
  }
  
  .bookmark-header {
    flex-direction: column;
    gap: var(--spacing-sm);
    align-items: flex-start;
  }
  
  .bookmark-actions {
    align-self: flex-end;
  }
  
  .bookmark-meta {
    flex-direction: column;
    gap: var(--spacing-xs);
  }
}
</style>