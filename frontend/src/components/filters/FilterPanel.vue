<template>
  <div class="filter-panel">
    <!-- Primary Search Row -->
    <div class="primary-search-row">
      <AppInput
        v-model="searchQuery"
        icon="🔍"
        placeholder="Search bookmarks by title, URL, or content..."
      />
      <AppButton 
        size="sm" 
        variant="secondary"
        @click="showAdvanced = !showAdvanced"
        :class="{ 'active': showAdvanced }"
      >
        🎯 {{ showAdvanced ? 'Hide' : 'More' }} Filters
      </AppButton>
    </div>
    
    <!-- Advanced Filters (Progressive Disclosure) -->
    <Transition name="slide">
      <div v-if="showAdvanced" class="advanced-filters">
        <div class="filter-row">
          <div class="filter-group">
            <label class="filter-label">Topic</label>
            <select v-model="localFilters.topic" class="filter-select">
              <option value="">All Topics</option>
              <option value="has-topic">Has Topic</option>
              <option value="no-topic">No Topic</option>
              <option 
                v-for="topic in availableTopics" 
                :key="topic" 
                :value="topic"
              >
                {{ topic }}
              </option>
            </select>
          </div>
          
          <div class="filter-group">
            <label class="filter-label">Domain</label>
            <select v-model="localFilters.domain" class="filter-select">
              <option value="">All Domains</option>
              <option 
                v-for="domain in availableDomains" 
                :key="domain" 
                :value="domain"
              >
                {{ domain }}
              </option>
            </select>
          </div>
          
          <div class="filter-group">
            <label class="filter-label">Age</label>
            <select v-model="localFilters.age" class="filter-select">
              <option value="">Any Time</option>
              <option value="today">Today</option>
              <option value="yesterday">Yesterday</option>
              <option value="week">This Week</option>
              <option value="month">This Month</option>
              <option value="older">Older</option>
            </select>
          </div>
          
          <div class="filter-actions">
            <AppButton size="sm" variant="success" @click="applyFilters">
              Apply
            </AppButton>
            <AppButton size="sm" variant="secondary" @click="clearAllFilters">
              Clear
            </AppButton>
          </div>
        </div>
      </div>
    </Transition>
    
    <!-- Active Filters Display -->
    <div v-if="hasActiveFilters" class="active-filters">
      <div class="active-filters-title">Active Filters:</div>
      <div class="filter-tags">
        <span
          v-for="(value, key) in activeFilterTags"
          :key="key"
          class="filter-tag"
        >
          {{ value }}
          <button class="filter-tag-remove" @click="removeFilter(key)">
            ×
          </button>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'
import type { FilterState } from '@/types'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

const bookmarkStore = useBookmarkStore()
const { filters, availableTopics, availableDomains } = storeToRefs(bookmarkStore)
const { updateFilters, clearFilters } = bookmarkStore

// Local reactive copy of filters
const localFilters = ref<FilterState>({ ...filters.value })
const searchQuery = ref(filters.value.search || '')

// Ensure search is properly initialized in localFilters
if (searchQuery.value) {
  localFilters.value.search = searchQuery.value
}
const showAdvanced = ref(false)

// Watch for external filter changes
watch(filters, (newFilters) => {
  localFilters.value = { ...newFilters }
  searchQuery.value = newFilters.search || ''
}, { deep: true })

// Watch for search query changes and apply filters immediately
watch(searchQuery, (newSearchQuery) => {
  localFilters.value.search = newSearchQuery
  applyFilters()
})

// Computed
const hasActiveFilters = computed(() => {
  return Object.values(localFilters.value).some(value => value && value.trim() !== '') ||
         (searchQuery.value && searchQuery.value.trim() !== '')
})

const activeFilterTags = computed(() => {
  const tags: Record<string, string> = {}
  
  if (searchQuery.value && searchQuery.value.trim() !== '') {
    tags.search = `Search: "${searchQuery.value}"`
  }
  
  if (localFilters.value.topic) {
    if (localFilters.value.topic === 'has-topic') {
      tags.topic = 'Has Topic'
    } else if (localFilters.value.topic === 'no-topic') {
      tags.topic = 'No Topic'
    } else {
      tags.topic = `Topic: ${localFilters.value.topic.replace('-', ' ')}`
    }
  }
  
  if (localFilters.value.domain) {
    tags.domain = `Domain: ${localFilters.value.domain}`
  }
  
  if (localFilters.value.age) {
    tags.age = `Age: ${localFilters.value.age}`
  }
  
  return tags
})

// Methods
const applyFilters = () => {
  updateFilters(localFilters.value)
}

const clearAllFilters = () => {
  localFilters.value = {}
  searchQuery.value = ''
  clearFilters()
}

const removeFilter = (filterKey: string) => {
  localFilters.value = { ...localFilters.value, [filterKey]: '' }
  if (filterKey === 'search') {
    searchQuery.value = ''
  }
  applyFilters()
}

</script>

<style scoped>
.filter-panel {
  background: #f7fafc;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-lg) var(--spacing-xl);
}

.primary-search-row {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.primary-search-row > :first-child {
  flex: 1;
}

.advanced-filters {
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--border-light);
}

.filter-row {
  display: flex;
  gap: var(--spacing-lg);
  align-items: end;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  min-width: 150px;
}

.filter-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-700);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.filter-select {
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  background: white;
  color: var(--color-gray-800);
  transition: var(--transition-fast);
}

.filter-select:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.filter-actions {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

/* Active Filters */
.active-filters {
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--border-light);
}

.active-filters-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}

.filter-tags {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.filter-tag {
  background: var(--color-primary);
  color: white;
  padding: var(--spacing-xs) var(--spacing-md);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.filter-tag-remove {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  opacity: 0.8;
  font-size: var(--font-size-lg);
  line-height: 1;
  padding: 0;
  margin-left: var(--spacing-xs);
}

.filter-tag-remove:hover {
  opacity: 1;
}

/* Button active state */
.btn.active {
  background: var(--color-primary);
  color: white;
}

/* Slide transition */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  margin-bottom: 0;
}

.slide-enter-to,
.slide-leave-from {
  opacity: 1;
  max-height: 200px;
}

/* Responsive */
@media (max-width: 768px) {
  .primary-search-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filter-group {
    min-width: auto;
  }
  
  .filter-actions {
    justify-content: flex-end;
  }
}
</style>
