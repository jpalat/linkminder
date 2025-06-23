<template>
  <div
    :class="[
      'bookmark-card',
      {
        'bookmark-card-selected': selected,
        'bookmark-card-batch-mode': batchMode
      }
    ]"
    @click="handleCardClick"
  >
    <input
      v-if="batchMode"
      type="checkbox"
      class="bookmark-checkbox"
      :checked="selected"
      @click.stop
      @change="$emit('toggle-selection', bookmark.id)"
    />
    
    <div class="bookmark-content">
      <div class="bookmark-header">
        <h3 class="bookmark-title">{{ bookmark.title }}</h3>
        <AppBadge
          v-if="bookmark.topic"
          :variant="getTopicVariant(bookmark.topic)"
          size="sm"
        >
          📁 {{ bookmark.topic }}
        </AppBadge>
      </div>
      
      <a
        :href="bookmark.url"
        class="bookmark-url"
        target="_blank"
        rel="noopener noreferrer"
        @click.stop
      >
        {{ bookmark.url }}
      </a>
      
      <div class="bookmark-meta">
        <span class="bookmark-domain">{{ bookmark.domain }}</span>
        <span class="bookmark-time">{{ bookmark.age }}</span>
        <span v-if="bookmark.action" class="bookmark-action">
          {{ getActionLabel(bookmark.action) }}
        </span>
      </div>
      
      <div v-if="bookmark.description" class="bookmark-description">
        {{ truncateText(bookmark.description, 120) }}
      </div>
    </div>
    
    <div class="bookmark-actions">
      <AppButton
        size="xs"
        variant="info"
        @click.stop="$emit('preview', bookmark.id)"
      >
        👁️
      </AppButton>
      <AppButton
        size="xs"
        variant="secondary"
        @click.stop="$emit('edit', bookmark.id)"
      >
        ✏️
      </AppButton>
      <AppButton
        v-if="bookmark.action !== 'working'"
        size="xs"
        variant="primary"
        @click.stop="$emit('move-to-working', bookmark.id)"
      >
        Work
      </AppButton>
      <AppButton
        v-if="bookmark.action !== 'share'"
        size="xs"
        variant="success"
        @click.stop="$emit('move-to-share', bookmark.id)"
      >
        Share
      </AppButton>
      <AppButton
        v-if="bookmark.action !== 'archived'"
        size="xs"
        variant="secondary"
        @click.stop="$emit('archive', bookmark.id)"
      >
        Archive
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Bookmark } from '@/types'
import AppButton from '@/components/ui/AppButton.vue'
import AppBadge from '@/components/ui/AppBadge.vue'

interface Props {
  bookmark: Bookmark
  selected?: boolean
  batchMode?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  selected: false,
  batchMode: false
})

const emit = defineEmits<{
  'toggle-selection': [id: string]
  'preview': [id: string]
  'edit': [id: string]
  'move-to-working': [id: string]
  'move-to-share': [id: string]
  'archive': [id: string]
  'click': [bookmark: Bookmark]
}>()

const handleCardClick = () => {
  if (props.batchMode) {
    emit('toggle-selection', props.bookmark.id)
  } else {
    emit('click', props.bookmark)
  }
}

const getTopicVariant = (topic: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
  const variants: Record<string, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    'ai-tools': 'info',
    'react-migration': 'primary',
    'css-learning': 'warning',
    'framework-research': 'success'
  }
  return variants[topic] || 'default'
}

const getActionLabel = (action: string) => {
  const labels: Record<string, string> = {
    'read-later': 'Read Later',
    'working': 'Working',
    'share': 'Ready to Share',
    'archived': 'Archived',
    'irrelevant': 'Irrelevant'
  }
  return labels[action] || action
}

const truncateText = (text: string, maxLength: number) => {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}
</script>

<style scoped>
.bookmark-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-sm) var(--spacing-md);
  background: var(--bg-card-hover);
  border-radius: var(--radius-md);
  border-left: 3px solid var(--border-light);
  transition: var(--transition-fast);
  gap: var(--spacing-sm);
  cursor: pointer;
  position: relative;
  margin-bottom: var(--spacing-xs);
}

.bookmark-card:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
  border-left-color: var(--color-primary);
}

.bookmark-card-selected {
  border-color: var(--color-primary);
  background: #ebf8ff;
  box-shadow: 0 0 0 2px rgba(66, 153, 225, 0.2);
}

.bookmark-card-batch-mode .bookmark-content {
  margin-left: 1.5rem;
}

.bookmark-checkbox {
  position: absolute;
  top: var(--spacing-sm);
  left: var(--spacing-sm);
  opacity: 0;
  transition: opacity var(--transition-fast);
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary);
}

.bookmark-card:hover .bookmark-checkbox,
.bookmark-card-batch-mode .bookmark-checkbox {
  opacity: 1;
}

.bookmark-content {
  flex: 1;
  min-width: 0;
}

.bookmark-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-xs);
}

.bookmark-title {
  font-weight: var(--font-weight-semibold);
  margin-bottom: 2px;
  color: var(--color-gray-800);
  line-height: var(--line-height-tight);
  font-size: 0.9rem;
  flex: 1;
}

.bookmark-url {
  color: var(--color-primary);
  font-size: 0.8rem;
  margin-bottom: 4px;
  text-decoration: none;
  display: block;
  word-break: break-all;
}

.bookmark-url:hover {
  text-decoration: underline;
}

.bookmark-meta {
  font-size: var(--font-size-xs);
  color: var(--color-gray-600);
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
  margin-bottom: 0;
}

.bookmark-domain {
  background: var(--color-gray-200);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-xs);
}

.bookmark-description {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  line-height: var(--line-height-normal);
  margin-top: var(--spacing-sm);
}

.bookmark-actions {
  display: flex;
  flex-direction: row;
  gap: 2px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity var(--transition-fast);
  align-items: center;
}

.bookmark-card:hover .bookmark-actions {
  opacity: 1;
}

/* Mobile responsiveness */
@media (max-width: 768px) {
  .bookmark-card {
    flex-direction: column;
    align-items: stretch;
  }
  
  .bookmark-actions {
    flex-direction: row;
    justify-content: flex-end;
    opacity: 1;
    margin-top: var(--spacing-sm);
  }
}
</style>
