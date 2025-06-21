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
            <AppButton 
              variant="secondary" 
              @click="toggleBatchMode"
            >
              {{ batchMode ? 'Cancel' : 'Select' }}
            </AppButton>
            <AppButton 
              variant="primary"
              @click="showAddModal = true"
            >
              + Add
            </AppButton>
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
                <AppButton size="sm" variant="secondary">
                  Sort
                </AppButton>
                <AppButton size="sm" variant="primary" @click="loadBookmarks" :loading="loading">
                  Refresh
                </AppButton>
              </div>
            </div>
            
            <!-- Filter Panel -->
            <FilterPanel v-if="showFilters" />
            
            <div class="section-content">
              <BookmarkList
                :bookmarks="filteredBookmarks"
                :batch-mode="batchMode"
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
          <AppButton size="sm" variant="secondary" @click="clearSelection">
            Cancel
          </AppButton>
        </div>
      </div>
    </Transition>
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
import ProjectList from '@/components/project/ProjectList.vue'
import ShareGroups from '@/components/share/ShareGroups.vue'

const bookmarkStore = useBookmarkStore()
const {
  currentTab,
  batchMode,
  selectedItems,
  filteredBookmarks,
  dashboardStats,
  shareGroups,
  loading
} = storeToRefs(bookmarkStore)

const {
  setCurrentTab,
  toggleBatchMode,
  toggleSelection,
  clearSelection,
  moveBookmarks,
  loadBookmarks,
  updateFilters
} = bookmarkStore

// Local state
const searchQuery = ref('')
const showFilters = ref(false)
const showAddModal = ref(false)

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

// Methods
const handleSearch = (query: string) => {
  updateFilters({ search: query })
}

const handlePreview = (bookmarkId: string) => {
  console.log('Preview bookmark:', bookmarkId)
  // Implementation for preview modal
}

const handleEdit = (bookmarkId: string) => {
  console.log('Edit bookmark:', bookmarkId)
  // Implementation for edit modal
}

const moveSelectedTo = (action: string) => {
  const selectedIds = Array.from(selectedItems.value)
  moveBookmarks(selectedIds, action)
}

// Lifecycle
onMounted(() => {
  loadBookmarks()
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
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-lg) var(--spacing-xl);
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  box-shadow: var(--shadow-sm);
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
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  display: block;
  color: var(--color-gray-800);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-gray-600);
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
