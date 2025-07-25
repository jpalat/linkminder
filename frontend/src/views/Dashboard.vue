<template>
  <div class="dashboard">
    <!-- Header with Tabs -->
    <header class="header">
      <div class="header-content">
        <div class="header-left">
          <!-- Tab Navigation -->
          <nav class="tab-nav">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              :class="['tab-button', { active: currentTab === tab.key }]"
              @click="setCurrentTab(tab.key)"
            >
              {{ tab.icon }} {{ tab.label }}
              <span v-if="tab.count !== undefined" class="tab-count">
                {{ tab.count }}
              </span>
            </button>
          </nav>
        </div>
        
        <div class="header-right">
          <div class="search-container">
            <AppInput
              v-model="searchInputValue"
              icon="🔍"
              placeholder="Search bookmarks..."
              class="compact-search"
            />
          </div>
          <div class="header-actions">
            <button 
              class="header-btn"
              @click="toggleBatchMode"
              :class="{ active: batchMode }"
            >
              {{ batchMode ? '✕' : '☑' }}
            </button>
            <button 
              class="header-btn header-btn-primary"
              @click="showAddModal = true"
            >
              +
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <div class="container">
        <!-- Tab Content -->
        <div class="tab-content">
          <!-- Triage Tab -->
          <div v-if="currentTab === 'triage'" class="section">
            <div class="section-header">
              <div class="section-title">
                🔍 Triage Queue
              </div>
              <div class="section-actions">
                <AppButton size="sm" variant="secondary" @click="showFilters = !showFilters">
                  Filter
                </AppButton>
                <AppButton size="sm" variant="secondary" @click="showSort = !showSort">
                  Sort
                </AppButton>
                <AppButton size="sm" variant="primary" @click="loadBookmarks" :loading="loading">
                  Refresh
                </AppButton>
              </div>
            </div>
            
            <!-- Filter Panel -->
            <FilterPanel v-if="showFilters" />
            
            <!-- Sort Panel -->
            <SortPanel 
              v-if="showSort" 
              :current-sort="currentSort"
              @sort-change="setSortOrder"
            />
            
            <div class="section-content">
              <BookmarkList
                :bookmarks="displayedBookmarks"
                :batch-mode="batchMode"
                :total-count="isGlobalSearchActive ? bookmarks.length : totalBookmarksForCurrentTab"
                :show-results-count="!!(hasActiveFilters || isGlobalSearchActive)"
                :loading="loading"
                @toggle-selection="toggleSelection"
                @edit="handleEdit"
                @move-to-working="handleMoveToWorking"
                @move-to-share="handleShareSingle"
                @archive="(id) => moveBookmarks([id], 'archived')"
                @delete="handleDeleteBookmark"
              />
            </div>
          </div>

          <!-- Projects Tab -->
          <div v-else-if="currentTab === 'projects'" class="section">
            <div class="section-header">
              <div class="section-title">
                🚀 Active Projects
              </div>
              <div class="section-actions">
                <AppButton size="sm" variant="primary">
                  New Project
                </AppButton>
              </div>
            </div>
            <div class="section-content">
              <ProjectList 
                :projects="projectStats" 
                @export-project="handleProjectExport"
              />
            </div>
          </div>

          <!-- Share Tab -->
          <div v-else-if="currentTab === 'share'" class="section">
            <div class="section-header">
              <div class="section-title">
                📤 Ready to Share
              </div>
            </div>
            <div class="section-content">
              <ShareGroups 
                :groups="shareGroups" 
                @edit-item="(item) => handleEdit(item.id)"
                @archive-item="(item) => moveBookmarks([item.id], 'archived')"
                @preview-item="handlePreviewItem"
                @share-item="(item) => handleShareSingle(item.id)"
              />
            </div>
          </div>

          <!-- Archive Tab -->
          <div v-else-if="currentTab === 'archive'" class="section">
            <div class="section-header">
              <div class="section-title">
                📦 Archive
              </div>
              <div class="section-actions">
                <AppButton size="sm" variant="secondary" @click="showFilters = !showFilters">
                  Filter
                </AppButton>
                <AppButton size="sm" variant="secondary" @click="showSort = !showSort">
                  Sort
                </AppButton>
                <AppButton size="sm" variant="secondary">
                  Export
                </AppButton>
              </div>
            </div>
            
            <!-- Filter Panel -->
            <FilterPanel v-if="showFilters" />
            
            <!-- Sort Panel -->
            <SortPanel 
              v-if="showSort" 
              :current-sort="currentSort"
              @sort-change="setSortOrder"
            />
            
            <div class="section-content">
              <BookmarkList
                :bookmarks="displayedBookmarks"
                :batch-mode="batchMode"
                :total-count="isGlobalSearchActive ? bookmarks.length : totalBookmarksForCurrentTab"
                :show-results-count="!!(hasActiveFilters || isGlobalSearchActive)"
                :loading="loading"
                @toggle-selection="toggleSelection"
                @edit="handleEdit"
                @move-to-working="handleMoveToWorking"
                @move-to-share="handleShareSingle"
                @archive="(id) => moveBookmarks([id], 'archived')"
                @delete="handleDeleteBookmark"
              />
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Batch Operations Bar -->
    <Transition name="slide">
      <div v-if="batchMode && selectedItems.size > 0" class="batch-bar">
        <div class="batch-count">
          {{ selectedItems.size }} item{{ selectedItems.size > 1 ? 's' : '' }} selected
        </div>
        <div class="batch-actions">
          <AppButton size="sm" @click="moveSelectedTo('working')">
            Move to Working
          </AppButton>
          <AppButton size="sm" @click="handleShareSelected">
            Share
          </AppButton>
          <AppButton size="sm" @click="moveSelectedTo('archived')">
            Archive
          </AppButton>
          <AppButton size="sm" variant="danger" @click="handleBatchDelete">
            Delete
          </AppButton>
          <AppButton size="sm" variant="secondary" @click="clearSelection">
            Cancel
          </AppButton>
        </div>
      </div>
    </Transition>

    <!-- Modals -->
    <AddBookmarkModal
      v-model:show="showAddModal"
      :existing-topics="existingTopics"
      @submit="handleAddBookmark"
    />


    <EditBookmarkModal
      v-model:show="showEditModal"
      :bookmark="selectedBookmark"
      :existing-topics="existingTopics"
      @submit="handleUpdateBookmark"
      @delete="handleDeleteBookmark"
    />

    <MoveToProjectModal
      v-model:show="showMoveToProjectModal"
      :bookmark="selectedBookmark"
      :existing-topics="existingTopics"
      @submit="handleMoveToProject"
    />

    <ShareBookmarkModal
      v-model:show="showShareModal"
      :bookmarks="bookmarksToShare"
      @submit="handleShareBookmarks"
    />

    <ConfirmModal
      v-model:show="showConfirmModal"
      :config="confirmConfig"
      :is-processing="isProcessingConfirm"
      @confirm="handleConfirmAction"
      @cancel="handleCancelConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { defineOptions } from 'vue'
import { ref, computed, onMounted } from 'vue'

defineOptions({
  name: 'DashboardView'
})
import { storeToRefs } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'
import type { TabType, ProjectStat, Bookmark, BookmarkAction } from '@/types'
import { useNotifications } from '@/composables/useNotifications'

interface BookmarkFormData {
  url: string
  title: string
  description: string
  action: string
  topic: string
  shareTo: string
  content: string
}

import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import BookmarkList from '@/components/bookmark/BookmarkList.vue'
import FilterPanel from '@/components/filters/FilterPanel.vue'
import SortPanel from '@/components/ui/SortPanel.vue'
import ProjectList from '@/components/project/ProjectList.vue'
import ShareGroups from '@/components/share/ShareGroups.vue'
import AddBookmarkModal from '@/components/modals/AddBookmarkModal.vue'
import EditBookmarkModal from '@/components/modals/EditBookmarkModal.vue'
import MoveToProjectModal from '@/components/modals/MoveToProjectModal.vue'
import ShareBookmarkModal from '@/components/modals/ShareBookmarkModal.vue'
import ConfirmModal, { type ConfirmationConfig } from '@/components/modals/ConfirmModal.vue'
import { ServerStatus } from '@/utils/serverStatus'

const bookmarkStore = useBookmarkStore()
const {
  currentTab,
  batchMode,
  selectedItems,
  filteredBookmarks,
  globalSearchResults,
  globalSearchQuery,
  dashboardStats,
  shareGroups,
  loading,
  bookmarks,
  filters,
  currentSort
} = storeToRefs(bookmarkStore)

const {
  setCurrentTab,
  toggleBatchMode,
  toggleSelection,
  clearSelection,
  moveBookmarks,
  loadBookmarks,
  updateFilters,
  setSortOrder,
  addBookmark,
  updateBookmark,
  loadDashboardStats,
  deleteBookmark,
  deleteBookmarks
} = bookmarkStore

// Local state
const searchInputValue = computed({
  get: () => globalSearchQuery.value,
  set: (value) => {
    globalSearchQuery.value = value
    // Clear local filters when doing global search
    if (value && value.trim() !== '') {
      updateFilters({ search: '' })
    }
  }
})

const showFilters = ref(false)
const showSort = ref(false)
const showAddModal = ref(false)
const showEditModal = ref(false)
const showMoveToProjectModal = ref(false)
const showShareModal = ref(false)
const showConfirmModal = ref(false)
const selectedBookmark = ref<Bookmark | null>(null)
const bookmarksToShare = ref<Bookmark[]>([])
const confirmConfig = ref<ConfirmationConfig>({ type: 'custom' })
const isProcessingConfirm = ref(false)
const pendingConfirmAction = ref<(() => void) | null>(null)

// Add notifications
const { bookmarkDeleted, bulkOperation, apiError } = useNotifications()

// Computed
const tabs = computed(() => [
  {
    key: 'triage' as TabType,
    label: 'Triage',
    icon: '🔍',
    count: dashboardStats.value.needsTriage
  },
  {
    key: 'projects' as TabType,
    label: 'Projects',
    icon: '🚀',
    count: dashboardStats.value.activeProjects
  },
  {
    key: 'share' as TabType,
    label: 'Ready to Share',
    icon: '📤',
    count: dashboardStats.value.readyToShare
  },
  {
    key: 'archive' as TabType,
    label: 'Archive',
    icon: '📦',
    count: dashboardStats.value.archived
  }
])

const projectStats = computed(() => dashboardStats.value.projectStats)

// Get unique topics for the topic selector
const existingTopics = computed(() => {
  const topics = new Set<string>()
  bookmarks.value.forEach(bookmark => {
    if (bookmark.topic) {
      topics.add(bookmark.topic)
    }
  })
  return Array.from(topics).sort()
})

const hasActiveFilters = computed(() => {
  return Object.values(filters.value).some(value => value && value.trim() !== '')
})

const totalBookmarksForCurrentTab = computed(() => {
  switch (currentTab.value) {
    case 'triage':
      return bookmarks.value.filter(b => !b.action || b.action === 'read-later').length
    case 'projects':
      return bookmarks.value.filter(b => b.action === 'working').length
    case 'share':
      return bookmarks.value.filter(b => b.action === 'share').length
    case 'archive':
      return bookmarks.value.filter(b => b.action === 'archived').length
    default:
      return 0
  }
})

// Computed for search mode
const isGlobalSearchActive = computed(() => {
  return globalSearchQuery.value && globalSearchQuery.value.trim() !== ''
})

const displayedBookmarks = computed(() => {
  return isGlobalSearchActive.value ? globalSearchResults.value : filteredBookmarks.value
})

// Methods
const handleEdit = (bookmarkId: string) => {
  const bookmark = bookmarks.value.find(b => b.id === bookmarkId)
  if (bookmark) {
    selectedBookmark.value = bookmark
    showEditModal.value = true
  }
}

const handleMoveToWorking = (bookmarkId: string) => {
  const bookmark = bookmarks.value.find(b => b.id === bookmarkId)
  if (bookmark) {
    selectedBookmark.value = bookmark
    showMoveToProjectModal.value = true
  }
}

const handleMoveToProject = async (bookmarkId: string, topic: string) => {
  try {
    await updateBookmark(bookmarkId, { action: 'working', topic })
    showMoveToProjectModal.value = false
    selectedBookmark.value = null
  } catch (error) {
    console.error('Failed to move bookmark to project:', error)
  }
}

const handleProjectExport = (project: ProjectStat) => {
  // Simple CSV export functionality
  const csvContent = `Project: ${project.topic}\nLinks: ${project.count}\nStatus: ${project.status}\nLast Updated: ${project.lastUpdated}\n\nExported at: ${new Date().toISOString()}`
  
  const blob = new Blob([csvContent], { type: 'text/plain' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `project-${project.topic.toLowerCase().replace(/\s+/g, '-')}.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

const handleShareSingle = (bookmarkId: string) => {
  const bookmark = filteredBookmarks.value.find(b => b.id === bookmarkId)
  if (bookmark) {
    bookmarksToShare.value = [bookmark]
    showShareModal.value = true
  }
}

const handleShareSelected = () => {
  const selectedIds = Array.from(selectedItems.value)
  const selectedBookmarks = filteredBookmarks.value.filter(b => selectedIds.includes(b.id))
  if (selectedBookmarks.length > 0) {
    bookmarksToShare.value = selectedBookmarks
    showShareModal.value = true
  }
}

const handleShareBookmarks = async (shareData: { destination: string; notes?: string }) => {
  try {
    const updatePromises = bookmarksToShare.value.map(bookmark =>
      updateBookmark(bookmark.id, { 
        action: 'share', 
        shareTo: shareData.destination 
      })
    )
    
    await Promise.all(updatePromises)
    
    // Clear selections if batch operation
    if (bookmarksToShare.value.length > 1) {
      clearSelection()
    }
    
    showShareModal.value = false
    bookmarksToShare.value = []
  } catch (error) {
    console.error('Failed to share bookmarks:', error)
  }
}

const handleAddBookmark = async (bookmarkData: BookmarkFormData) => {
  try {
    // Add the bookmark to the store
    await addBookmark({
      ...bookmarkData,
      action: bookmarkData.action as BookmarkAction
    })
    console.log('Added bookmark:', bookmarkData)
    // TODO: Show success notification
  } catch (error) {
    console.error('Failed to add bookmark:', error)
    // TODO: Show error notification
  }
}

const handleUpdateBookmark = async (bookmarkData: Bookmark) => {
  try {
    // Update the bookmark in the store
    await updateBookmark(bookmarkData.id, bookmarkData)
    console.log('Updated bookmark:', bookmarkData)
    // TODO: Show success notification
  } catch (error) {
    console.error('Failed to update bookmark:', error)
    // TODO: Show error notification
  }
}

const handleDeleteBookmark = (bookmarkId: string) => {
  const bookmark = bookmarks.value.find(b => b.id === bookmarkId)
  if (!bookmark) return

  confirmConfig.value = {
    type: 'delete',
    title: 'Delete Bookmark',
    message: 'Are you sure you want to delete this bookmark?',
    details: 'This action cannot be undone.',
    items: [bookmark],
    isDestructive: true
  }
  
  pendingConfirmAction.value = async () => {
    try {
      await deleteBookmark(bookmarkId)
      showEditModal.value = false
      bookmarkDeleted(bookmark.title)
    } catch (error) {
      console.error('Failed to delete bookmark:', error)
      apiError('delete bookmark', error as Error)
    }
  }
  
  showConfirmModal.value = true
}

const handleBatchDelete = () => {
  const selectedIds = Array.from(selectedItems.value)
  const selectedBookmarks = bookmarks.value.filter(b => selectedIds.includes(b.id))
  
  if (selectedBookmarks.length === 0) return

  confirmConfig.value = {
    type: 'delete',
    title: `Delete ${selectedBookmarks.length} Bookmark${selectedBookmarks.length > 1 ? 's' : ''}`,
    message: `Are you sure you want to delete ${selectedBookmarks.length} bookmark${selectedBookmarks.length > 1 ? 's' : ''}?`,
    details: 'This action cannot be undone.',
    items: selectedBookmarks,
    isDestructive: true
  }
  
  pendingConfirmAction.value = async () => {
    try {
      await deleteBookmarks(selectedIds)
      clearSelection()
      bulkOperation(selectedBookmarks.length, 'deleted')
    } catch (error) {
      console.error('Failed to delete bookmarks:', error)
      apiError('delete bookmarks', error as Error)
    }
  }
  
  showConfirmModal.value = true
}

const handleConfirmAction = async () => {
  if (!pendingConfirmAction.value) return
  
  isProcessingConfirm.value = true
  
  try {
    await pendingConfirmAction.value()
  } catch (error) {
    console.error('Error executing confirm action:', error)
    // TODO: Show error notification
  } finally {
    isProcessingConfirm.value = false
    showConfirmModal.value = false
    pendingConfirmAction.value = null
  }
}

const handleCancelConfirm = () => {
  pendingConfirmAction.value = null
  showConfirmModal.value = false
}

const moveSelectedTo = (action: string) => {
  const selectedIds = Array.from(selectedItems.value)
  moveBookmarks(selectedIds, action as BookmarkAction)
}

const handlePreviewItem = (item: Bookmark) => {
  // For now, just open the URL in a new tab
  // This could be enhanced with a modal preview in the future
  window.open(item.url, '_blank')
}

// Lifecycle
onMounted(async () => {
  // Debug: Log current server configuration
  console.log('🌐 API Server URL:', ServerStatus.getCurrentServerURL())
  
  // Check server health
  const healthCheck = await ServerStatus.checkServerHealth()
  console.log('🏥 Server Health:', healthCheck)
  
  if (!healthCheck.isConnected) {
    console.warn('⚠️  Server not reachable, will use mock data fallback')
  }
  
  await loadBookmarks()
  await loadDashboardStats()
})
</script>

<style scoped>
.dashboard {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Minimal Header */
.header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-md) var(--spacing-xl);
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-xl);
}

.header-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.search-container {
  min-width: 260px;
}

.compact-search {
  font-size: var(--font-size-sm);
}

.header-actions {
  display: flex;
  gap: var(--spacing-xs);
}

.header-btn {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  border: 1px solid var(--border-light);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: var(--transition-fast);
  min-width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-btn:hover {
  background: var(--color-gray-200);
  border-color: var(--color-gray-300);
}

.header-btn.active {
  background: var(--color-primary-light);
  color: var(--color-primary-dark);
  border-color: var(--color-primary);
}

.header-btn-primary {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.header-btn-primary:hover {
  background: var(--color-primary-dark);
  border-color: var(--color-primary-dark);
}

/* Tab Navigation in Header */
.tab-nav {
  display: flex;
  gap: var(--spacing-lg);
}

.tab-button {
  background: none;
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-600);
  cursor: pointer;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  border-radius: var(--border-radius);
  white-space: nowrap;
}

.tab-button:hover {
  color: var(--color-gray-800);
  background: var(--color-gray-50);
}

.tab-button.active {
  color: var(--color-primary);
  background: var(--color-primary-light);
}

.tab-count {
  background: var(--color-gray-200);
  color: var(--color-gray-700);
  padding: 2px var(--spacing-xs);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  min-width: 18px;
  text-align: center;
  line-height: 1;
}

.tab-button.active .tab-count {
  background: var(--color-primary);
  color: white;
}

/* Main Content */
.main-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-xl);
  scroll-behavior: smooth;
}

/* Scrollbar styling for main content */
.main-content::-webkit-scrollbar {
  width: 8px;
}

.main-content::-webkit-scrollbar-track {
  background: var(--color-gray-50);
}

.main-content::-webkit-scrollbar-thumb {
  background: var(--color-gray-300);
  border-radius: 4px;
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: var(--color-gray-400);
}

.container {
  max-width: 1400px;
  margin: 0 auto;
}

.section-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* Batch Operations Bar */
.batch-bar {
  position: fixed;
  bottom: var(--spacing-2xl);
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-gray-800);
  color: white;
  padding: var(--spacing-lg) var(--spacing-2xl);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  z-index: var(--z-modal);
}

.batch-count {
  font-weight: var(--font-weight-semibold);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.batch-count::before {
  content: '✓';
  background: rgba(255, 255, 255, 0.2);
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
}

.batch-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* Responsive Design */
@media (max-width: 1200px) {
  .search-container {
    min-width: 250px;
  }
}

@media (max-width: 968px) {
  .header {
    padding: var(--spacing-sm) var(--spacing-lg);
  }
  
  .header-content {
    flex-direction: column;
    gap: var(--spacing-md);
  }
  
  .header-left {
    order: 2;
    justify-content: flex-start;
    overflow-x: auto;
  }
  
  .header-right {
    order: 1;
    width: 100%;
  }
  
  .search-container {
    min-width: auto;
    flex: 1;
  }
  
  .tab-nav {
    overflow-x: auto;
    gap: var(--spacing-md);
    padding-bottom: var(--spacing-xs);
  }
  
  .tab-nav::-webkit-scrollbar {
    height: 2px;
  }
  
  .tab-nav::-webkit-scrollbar-track {
    background: transparent;
  }
  
  .tab-nav::-webkit-scrollbar-thumb {
    background: var(--color-gray-300);
    border-radius: 1px;
  }
  
  .batch-bar {
    left: var(--spacing-lg);
    right: var(--spacing-lg);
    transform: none;
    flex-direction: column;
    gap: var(--spacing-md);
  }
}

@media (max-width: 640px) {
  .header {
    padding: var(--spacing-sm) var(--spacing-md);
  }
  
  .header-right {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .main-content {
    padding: var(--spacing-md);
    max-width: 100vw;
    overflow-x: hidden;
  }
  
  .search-container {
    min-width: 200px;
  }
  
  .tab-nav {
    gap: var(--spacing-sm);
  }
  
  .tab-button {
    font-size: var(--font-size-xs);
    padding: var(--spacing-xs) var(--spacing-sm);
  }
}

@media (max-width: 480px) {
  .main-content {
    padding: var(--spacing-sm);
  }
}
</style>
