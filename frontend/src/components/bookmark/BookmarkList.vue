<template>
  <div class="bookmark-list">
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
        @preview="$emit('preview', $event)"
        @edit="$emit('edit', $event)"
        @move-to-working="$emit('move-to-working', $event)"
        @move-to-share="$emit('move-to-share', $event)"
        @archive="$emit('archive', $event)"
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
}

const props = withDefaults(defineProps<Props>(), {
  batchMode: false,
  loading: false,
  emptyMessage: 'Try adding some bookmarks or adjusting your filters.'
})

const emit = defineEmits<{
  'toggle-selection': [id: string]
  'preview': [id: string]
  'edit': [id: string]
  'move-to-working': [id: string]
  'move-to-share': [id: string]
  'archive': [id: string]
  'bookmark-click': [bookmark: Bookmark]
}>()

const bookmarkStore = useBookmarkStore()
const { selectedItems } = storeToRefs(bookmarkStore)

const handleBookmarkClick = (bookmark: Bookmark) => {
  emit('bookmark-click', bookmark)
}
</script>

<style scoped>
.bookmark-list {
  width: 100%;
}

.bookmark-grid {
  display: grid;
  gap: var(--spacing-sm);
  max-height: 600px;
  overflow-y: auto;
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

/* Scrollbar styling */
.bookmark-grid::-webkit-scrollbar {
  width: 6px;
}

.bookmark-grid::-webkit-scrollbar-track {
  background: var(--color-gray-100);
  border-radius: var(--radius-sm);
}

.bookmark-grid::-webkit-scrollbar-thumb {
  background: var(--color-gray-400);
  border-radius: var(--radius-sm);
}

.bookmark-grid::-webkit-scrollbar-thumb:hover {
  background: var(--color-gray-500);
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .bookmark-grid {
    max-height: 500px;
  }
}
</style>
