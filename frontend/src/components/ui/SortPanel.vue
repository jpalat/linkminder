<template>
  <div class="sort-panel">
    <div class="sort-options">
      <button
        v-for="option in sortOptions"
        :key="option.key"
        :class="['sort-option', { active: currentSort === option.key }]"
        @click="handleSortChange(option.key)"
      >
        {{ option.icon }} {{ option.label }}
        <span class="sort-direction">
          {{ getSortDirection(option.key) }}
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface SortOption {
  key: string
  label: string
  icon: string
}

interface Props {
  currentSort?: string
}

const props = withDefaults(defineProps<Props>(), {
  currentSort: 'date-desc'
})

const emit = defineEmits<{
  'sort-change': [sortKey: string]
}>()

const sortOptions: SortOption[] = [
  { key: 'date-desc', label: 'Newest First', icon: '📅' },
  { key: 'date-asc', label: 'Oldest First', icon: '📅' },
  { key: 'title-asc', label: 'Title A-Z', icon: '📝' },
  { key: 'title-desc', label: 'Title Z-A', icon: '📝' },
  { key: 'domain-asc', label: 'Domain A-Z', icon: '🌐' },
  { key: 'topic-asc', label: 'Topic A-Z', icon: '📁' }
]

const getSortDirection = (sortKey: string): string => {
  if (sortKey.endsWith('-desc')) return '↓'
  if (sortKey.endsWith('-asc')) return '↑'
  return ''
}

const handleSortChange = (sortKey: string) => {
  emit('sort-change', sortKey)
}
</script>

<style scoped>
.sort-panel {
  background: #f0fff4;
  border-bottom: 1px solid #9ae6b4;
  padding: var(--spacing-lg) var(--spacing-xl);
}

.sort-options {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.sort-option {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-lg);
  background: white;
  border: 1px solid var(--border-light);
  border-radius: 20px;
  cursor: pointer;
  transition: var(--transition-fast);
  font-size: var(--font-size-base);
  color: var(--color-gray-700);
}

.sort-option:hover {
  border-color: #48bb78;
  background: #f0fff4;
}

.sort-option.active {
  background: #48bb78;
  color: white;
  border-color: #48bb78;
}

.sort-direction {
  font-size: var(--font-size-sm);
  margin-left: var(--spacing-xs);
}

/* Responsive */
@media (max-width: 768px) {
  .sort-options {
    flex-direction: column;
  }
}
</style>
