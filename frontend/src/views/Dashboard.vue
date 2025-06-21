<template>
  <div class="dashboard">
    <!-- Header with Search -->
    <header class="header">
      <div class="header-content">
        <div class="header-left">
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-number">{{ dashboardStats.needsTriage }}</span>
              <span class="stat-label">Triage</span>
            </div>
            <div class="stat-item">
              <span class="stat-number">{{ dashboardStats.readyToShare }}</span>
              <span class="stat-label">Share</span>
            </div>
            <div class="stat-item">
              <span class="stat-number">{{ dashboardStats.activeProjects }}</span>
              <span class="stat-label">Projects</span>
            </div>
            <div class="stat-item">
              <span class="stat-number">{{ dashboardStats.archived }}</span>
              <span class="stat-label">Archived</span>
            </div>
            <div class="stat-item">
              <span class="stat-number">{{ dashboardStats.totalBookmarks }}</span>
              <span class="stat-label">Total</span>
            </div>
          </div>
        </div>
        
        <div class="header-right">
          <div class="search-container">
            <AppInput
              v-model="searchQuery"
              icon="🔍"
              placeholder="Search bookmarks, topics, domains..."
              @input="handleSearch"
            />
          </div>
          <div class="header-actions">
            <button 
              class="header-btn"
              @click="toggleBatchMode"
            >
              {{ batchMode ? 'Cancel' : 'Select' }}
            </button>
            <button 
              class="header-btn header-btn-primary"
              @click="showAddModal = true"
            >
              + Add
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Tab Navigation -->
    <nav class="tab-nav">
      <div class="tab-nav-content">
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
      </div>
    </nav>

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
                :bookmarks="filteredBookmarks"
                :batch-mode="batchMode"
                :total-count="totalBookmarksForCurrentTab"
                :show-results-count="hasActiveFilters"
                :loading="loading"
                @toggle-selection="toggleSelection"
                @preview="handlePreview"
                @edit="handleEdit"
                @move-to-working="(id) => moveBookmarks([id], 'working')"
                @move-to-share="(id) => moveBookmarks([id], 'share')"
                @archive="(id) => moveBookmarks([id], 'archived')"
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
              <ProjectList :projects="projectStats" />
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
              <ShareGroups :groups="shareGroups" />
            </div>
          </div>

          <!-- Archive Tab -->
          <div v-else-if="currentTab === 'archive'" class="section">
            <div class="section-header">
              <div class="section-title">
                📦 Archive
              </div>
              <div class="section-actions">
                <AppButton size="sm" variant="secondary">
                  Export
                </AppButton>
              </div>
            </div>
            <div class="section-content">
              <BookmarkList
                :bookmarks="filteredBookmarks"
                :batch-mode="batchMode"
                @toggle-selection="toggleSelection"
                @preview="handlePreview"
                @edit="handleEdit"
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
          <AppButton size="sm" @click="moveSelectedTo('share')">
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

    <PreviewModal
      v-model:show="showPreviewModal"
      :bookmark="selectedBookmark"
      @edit="handleEdit"
      @move-to-share="(id) => moveBookmarks([id], 'share')"
      @move-to-working="(id) => moveBookmarks([id], 'working')"
      @archive="(id) => moveBookmarks([id], 'archived')"
    />

    <EditBookmarkModal
      v-model:show="showEditModal"
      :bookmark="selectedBookmark"
      :existing-topics="existingTopics"
      @submit="handleUpdateBookmark"
      @delete="handleDeleteBookmark"
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
import { ref, computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'
import type { TabType } from '@/types'

import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import BookmarkList from '@/components/bookmark/BookmarkList.vue'
import FilterPanel from '@/components/filters/FilterPanel.vue'
import SortPanel from '@/components/ui/SortPanel.vue'
import ProjectList from '@/components/project/ProjectList.vue'
import ShareGroups from '@/components/share/ShareGroups.vue'
import AddBookmarkModal from '@/components/modals/AddBookmarkModal.vue'
import PreviewModal from '@/components/modals/PreviewModal.vue'
import EditBookmarkModal from '@/components/modals/EditBookmarkModal.vue'
import ConfirmModal, { type ConfirmationConfig } from '@/components/modals/ConfirmModal.vue'
import type { Bookmark } from '@/types'
import { ServerStatus } from '@/utils/serverStatus'

const bookmarkStore = useBookmarkStore()
const {
  currentTab,
  batchMode,
  selectedItems,
  filteredBookmarks,
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
  loadDashboardStats
} = bookmarkStore

// Local state
const searchQuery = ref('')
const showFilters = ref(false)
const showSort = ref(false)
const showAddModal = ref(false)
const showPreviewModal = ref(false)
const showEditModal = ref(false)
const showConfirmModal = ref(false)
const selectedBookmark = ref<Bookmark | null>(null)
const confirmConfig = ref<ConfirmationConfig>({ type: 'custom' })
const isProcessingConfirm = ref(false)
const pendingConfirmAction = ref<(() => void) | null>(null)

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

// Methods
const handleSearch = (query: string) => {
  updateFilters({ search: query })
}

const handlePreview = (bookmarkId: string) => {
  const bookmark = bookmarks.value.find(b => b.id === bookmarkId)
  if (bookmark) {
    selectedBookmark.value = bookmark
    showPreviewModal.value = true
  }
}

const handleEdit = (bookmarkId: string) => {
  const bookmark = bookmarks.value.find(b => b.id === bookmarkId)
  if (bookmark) {
    selectedBookmark.value = bookmark
    showEditModal.value = true
  }
}

const handleAddBookmark = async (bookmarkData: any) => {
  try {
    // Add the bookmark to the store
    await addBookmark(bookmarkData)
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
  
  pendingConfirmAction.value = () => {
    // Remove from bookmarks array (this would be an API call in real implementation)
    const index = bookmarks.value.findIndex(b => b.id === bookmarkId)
    if (index !== -1) {
      bookmarks.value.splice(index, 1)
    }
    showEditModal.value = false
    console.log('Deleted bookmark:', bookmarkId)
    // TODO: Show success notification
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
  
  pendingConfirmAction.value = () => {
    // Remove from bookmarks array (this would be an API call in real implementation)
    selectedIds.forEach(id => {
      const index = bookmarks.value.findIndex(b => b.id === id)
      if (index !== -1) {
        bookmarks.value.splice(index, 1)
      }
    })
    clearSelection()
    console.log('Deleted bookmarks:', selectedIds)
    // TODO: Show success notification
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
  moveBookmarks(selectedIds, action)
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
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Header */
.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 12px var(--spacing-xl);
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-2xl);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-2xl);
}

.stats-grid {
  display: flex;
  gap: var(--spacing-xl);
}

.stat-item {
  text-align: center;
  min-width: 50px;
}

.stat-number {
  font-size: 1.5rem;
  font-weight: var(--font-weight-bold);
  display: block;
  color: white;
}

.stat-label {
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.9);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.search-container {
  min-width: 300px;
}

.header-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.header-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: none;
  padding: 8px var(--spacing-lg);
  border-radius: 20px;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: var(--transition-fast);
}

.header-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

.header-btn-primary {
  background: rgba(255, 255, 255, 0.9);
  color: var(--color-primary);
}

.header-btn-primary:hover {
  background: white;
}

/* Tab Navigation */
.tab-nav {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: 0 var(--spacing-xl);
}

.tab-nav-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  gap: var(--spacing-lg);
}

.tab-button {
  background: none;
  border: none;
  padding: var(--spacing-lg) 0;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-600);
  cursor: pointer;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  border-bottom: 2px solid transparent;
}

.tab-button:hover {
  color: var(--color-gray-800);
}

.tab-button.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-count {
  background: var(--color-gray-200);
  color: var(--color-gray-700);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  min-width: 20px;
  text-align: center;
}

.tab-button.active .tab-count {
  background: var(--color-primary);
  color: white;
}

/* Main Content */
.main-content {
  flex: 1;
  padding: var(--spacing-xl);
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
  .header-content {
    flex-direction: column;
    gap: var(--spacing-lg);
  }
  
  .header-left {
    order: 2;
    justify-content: center;
  }
  
  .header-right {
    order: 1;
    flex-direction: column;
    gap: var(--spacing-md);
  }
  
  .tab-nav-content {
    overflow-x: auto;
    gap: var(--spacing-md);
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
  .stats-grid {
    gap: var(--spacing-md);
  }
  
  .stat-number {
    font-size: var(--font-size-xl);
  }
  
  .main-content {
    padding: var(--spacing-lg);
  }
  
  .search-container {
    min-width: 200px;
  }
}
</style>
