<template>
  <div class="bookmark-list">
    <!-- Results Info -->
    <div v-if="showResultsCount && !loading" class="results-info">
      {{ resultsText }}
    </div>
    
    <div v-if="loading" class="loading">
      Loading bookmarks...
    </div>
    
    <div v-else-if="bookmarks.length === 0" class="empty-state">
      <h3>No bookmarks found</h3>
      <p>{{ emptyMessage }}</p>
    </div>
    
    <TransitionGroup
      v-else
      name="bookmark"
      tag="div"
      class="bookmark-grid"
    >
      <BookmarkCard
        v-for="bookmark in bookmarks"
        :key="bookmark.id"
        :bookmark="bookmark"
        :selected="selectedItems.has(bookmark.id)"
        :batch-mode="batchMode"
        @toggle-selection="$emit('toggle-selection', $event)"
        @edit="$emit('edit', $event)"
        @move-to-working="$emit('move-to-working', $event)"
        @move-to-share="$emit('move-to-share', $event)"
        @archive="$emit('archive', $event)"
        @delete="$emit('delete', $event)"
        @click="handleBookmarkClick"
      />
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'
import type { Bookmark } from '@/types'
import BookmarkCard from './BookmarkCard.vue'

interface Props {
  bookmarks: Bookmark[]
  batchMode?: boolean
  loading?: boolean
  emptyMessage?: string
  totalCount?: number
  showResultsCount?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  batchMode: false,
  loading: false,
  emptyMessage: 'Try adding some bookmarks or adjusting your filters.',
  showResultsCount: false
})

const emit = defineEmits<{
  'toggle-selection': [id: string]
  'edit': [id: string]
  'move-to-working': [id: string]
  'move-to-share': [id: string]
  'archive': [id: string]
  'delete': [id: string]
  'bookmark-click': [bookmark: Bookmark]
}>()

const bookmarkStore = useBookmarkStore()
const { selectedItems } = storeToRefs(bookmarkStore)

const resultsText = computed(() => {
  const count = props.bookmarks.length
  const total = props.totalCount || count
  
  if (props.totalCount && count < total) {
    return `Showing ${count} of ${total} items`
  }
  return `${count} item${count !== 1 ? 's' : ''}`
})

const handleBookmarkClick = (bookmark: Bookmark) => {
  emit('bookmark-click', bookmark)
}
</script>

<style scoped>
.bookmark-list {
  width: 100%;
}

.results-info {
  padding: var(--spacing-sm) var(--spacing-md);
  background: #fffbeb;
  border-bottom: 1px solid #fed7aa;
  font-size: var(--font-size-sm);
  color: #92400e;
  border-radius: var(--radius-md) var(--radius-md) 0 0;
  margin-bottom: var(--spacing-xs);
}

.bookmark-grid {
  display: grid;
  gap: 2px;
}

/* Transition animations */
.bookmark-move,
.bookmark-enter-active,
.bookmark-leave-active {
  transition: var(--transition-normal);
}

.bookmark-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.bookmark-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.bookmark-leave-active {
  position: absolute;
  right: 0;
  left: 0;
}


/* Responsive adjustments */
@media (max-width: 768px) {
  .bookmark-grid {
    gap: 1px;
  }
}
</style>
