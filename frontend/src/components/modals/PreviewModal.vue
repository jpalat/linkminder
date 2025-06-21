<template>
  <AppModal
    :show="show"
    :title="bookmark?.title || 'Bookmark Preview'"
    size="lg"
    @update:show="$emit('update:show', $event)"
    @close="handleClose"
  >
    <div v-if="bookmark" class="preview-content">
      <!-- Header with metadata -->
      <div class="preview-header">
        <div class="bookmark-info">
          <h3 class="bookmark-title">{{ bookmark.title }}</h3>
          <a
            :href="bookmark.url"
            target="_blank"
            rel="noopener noreferrer"
            class="bookmark-url"
          >
            {{ bookmark.url }}
            <span class="external-icon">↗</span>
          </a>
          <div class="bookmark-meta">
            <span class="meta-item">
              <span class="meta-label">Domain:</span>
              {{ bookmark.domain }}
            </span>
            <span class="meta-item">
              <span class="meta-label">Added:</span>
              {{ formatDate(bookmark.timestamp) }}
            </span>
            <span class="meta-item">
              <span class="meta-label">Age:</span>
              {{ bookmark.age }}
            </span>
            <span v-if="bookmark.action" class="meta-item">
              <span class="meta-label">Status:</span>
              <AppBadge :variant="getActionVariant(bookmark.action)" size="sm">
                {{ getActionLabel(bookmark.action) }}
              </AppBadge>
            </span>
            <span v-if="bookmark.topic" class="meta-item">
              <span class="meta-label">Project:</span>
              <AppBadge variant="info" size="sm">
                {{ bookmark.topic }}
              </AppBadge>
            </span>
          </div>
        </div>
      </div>

      <!-- Description -->
      <div v-if="bookmark.description" class="preview-section">
        <h4 class="section-title">Description</h4>
        <div class="description-content">
          {{ bookmark.description }}
        </div>
      </div>

      <!-- Content -->
      <div v-if="bookmark.content" class="preview-section">
        <h4 class="section-title">Content</h4>
        <div class="content-display">
          <div class="content-text">
            {{ bookmark.content }}
          </div>
        </div>
      </div>

      <!-- Content loading placeholder -->
      <div v-else-if="isLoadingContent" class="preview-section">
        <h4 class="section-title">Content</h4>
        <div class="content-loading">
          <div class="loading-spinner">🔄</div>
          <span>Loading page content...</span>
        </div>
      </div>

      <!-- No content available -->
      <div v-else class="preview-section">
        <h4 class="section-title">Content</h4>
        <div class="no-content">
          <div class="no-content-icon">📄</div>
          <p>No content preview available</p>
          <AppButton size="sm" variant="secondary" @click="loadContent">
            Try to Load Content
          </AppButton>
        </div>
      </div>

      <!-- Actions section -->
      <div class="preview-actions">
        <h4 class="section-title">Actions</h4>
        <div class="action-buttons">
          <AppButton
            size="sm"
            variant="primary"
            @click="openInNewTab"
          >
            🔗 Open Link
          </AppButton>
          <AppButton
            size="sm"
            variant="secondary"
            @click="$emit('edit', bookmark.id)"
          >
            ✏️ Edit
          </AppButton>
          <AppButton
            size="sm"
            variant="secondary"
            @click="copyToClipboard"
          >
            📋 Copy URL
          </AppButton>
          <AppButton
            v-if="bookmark.action !== 'share'"
            size="sm"
            variant="info"
            @click="$emit('move-to-share', bookmark.id)"
          >
            📤 Share
          </AppButton>
          <AppButton
            v-if="bookmark.action !== 'working'"
            size="sm"
            variant="success"
            @click="$emit('move-to-working', bookmark.id)"
          >
            🚀 Move to Working
          </AppButton>
          <AppButton
            v-if="bookmark.action !== 'archived'"
            size="sm"
            variant="info"
            @click="$emit('archive', bookmark.id)"
          >
            📦 Archive
          </AppButton>
        </div>
      </div>

      <!-- Tags/Categories (if any) -->
      <div v-if="bookmark.tags && bookmark.tags.length > 0" class="preview-section">
        <h4 class="section-title">Tags</h4>
        <div class="tags-container">
          <span
            v-for="tag in bookmark.tags"
            :key="tag"
            class="tag-chip"
          >
            #{{ tag }}
          </span>
        </div>
      </div>

      <!-- Share destination (if applicable) -->
      <div v-if="bookmark.action === 'share' && bookmark.shareTo" class="preview-section">
        <h4 class="section-title">Share Destination</h4>
        <div class="share-destination">
          <span class="destination-icon">{{ getDestinationIcon(bookmark.shareTo) }}</span>
          <span class="destination-name">{{ bookmark.shareTo }}</span>
        </div>
      </div>
    </div>

    <div v-else class="no-bookmark">
      <div class="no-bookmark-icon">❌</div>
      <p>No bookmark selected for preview</p>
    </div>

    <template #footer>
      <AppButton variant="secondary" @click="handleClose">
        Close
      </AppButton>
      <AppButton
        v-if="bookmark"
        variant="primary"
        @click="$emit('edit', bookmark.id)"
      >
        Edit Bookmark
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useClipboard } from '@vueuse/core'
import type { Bookmark } from '@/types'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppBadge from '@/components/ui/AppBadge.vue'

interface Props {
  show: boolean
  bookmark?: Bookmark | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'edit': [bookmarkId: string]
  'move-to-share': [bookmarkId: string]
  'move-to-working': [bookmarkId: string]
  'archive': [bookmarkId: string]
}>()

const isLoadingContent = ref(false)
const { copy } = useClipboard()

// Helper functions
const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getActionVariant = (action: string): 'default' | 'primary' | 'success' | 'info' | 'danger' => {
  const variants: Record<string, 'default' | 'primary' | 'success' | 'info' | 'danger'> = {
    'read-later': 'info',
    'working': 'success',
    'share': 'primary',
    'archived': 'default'
  }
  return variants[action] || 'default'
}

const getActionLabel = (action: string): string => {
  const labels: Record<string, string> = {
    'read-later': 'Read Later',
    'working': 'Working',
    'share': 'Ready to Share',
    'archived': 'Archived'
  }
  return labels[action] || action
}

const getDestinationIcon = (destination: string): string => {
  const icons: Record<string, string> = {
    'Team Slack': '💬',
    'Newsletter': '📧',
    'Dev Blog': '📝',
    'Unassigned': '📤'
  }
  return icons[destination] || '📤'
}

// Event handlers
const handleClose = () => {
  emit('update:show', false)
}

const openInNewTab = () => {
  if (props.bookmark?.url) {
    window.open(props.bookmark.url, '_blank', 'noopener,noreferrer')
  }
}

const copyToClipboard = async () => {
  if (props.bookmark?.url) {
    await copy(props.bookmark.url)
    // TODO: Show toast notification
    console.log('URL copied to clipboard')
  }
}

const loadContent = async () => {
  if (!props.bookmark?.url) return
  
  isLoadingContent.value = true
  
  try {
    // Mock content loading - in real implementation, this would fetch page content
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    // This would update the bookmark with loaded content
    console.log('Content loading would happen here for:', props.bookmark.url)
  } catch (error) {
    console.error('Failed to load content:', error)
  } finally {
    isLoadingContent.value = false
  }
}
</script>

<style scoped>
.preview-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

/* Header */
.preview-header {
  border-bottom: 1px solid var(--border-light);
  padding-bottom: var(--spacing-lg);
}

.bookmark-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.bookmark-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  margin: 0;
  line-height: var(--line-height-tight);
}

.bookmark-url {
  color: var(--color-primary);
  text-decoration: none;
  font-size: var(--font-size-base);
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  word-break: break-all;
  transition: var(--transition-fast);
}

.bookmark-url:hover {
  text-decoration: underline;
}

.external-icon {
  font-size: var(--font-size-sm);
  opacity: 0.7;
}

.bookmark-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-lg);
  font-size: var(--font-size-sm);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.meta-label {
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-600);
}

/* Sections */
.preview-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-700);
  margin: 0;
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-light);
}

/* Description */
.description-content {
  background: var(--color-gray-50);
  padding: var(--spacing-lg);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
  line-height: var(--line-height-relaxed);
  color: var(--color-gray-700);
}

/* Content */
.content-display {
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.content-text {
  padding: var(--spacing-lg);
  max-height: 300px;
  overflow-y: auto;
  line-height: var(--line-height-relaxed);
  color: var(--color-gray-700);
  white-space: pre-wrap;
}

.content-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-md);
  padding: var(--spacing-xl);
  color: var(--color-gray-600);
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.no-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-xl);
  color: var(--color-gray-600);
  text-align: center;
}

.no-content-icon {
  font-size: var(--font-size-3xl);
  opacity: 0.7;
}

/* Actions */
.preview-actions {
  border-top: 1px solid var(--border-light);
  padding-top: var(--spacing-lg);
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

/* Tags */
.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.tag-chip {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

/* Share destination */
.share-destination {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.destination-icon {
  font-size: var(--font-size-lg);
}

.destination-name {
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-700);
}

/* No bookmark state */
.no-bookmark {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-xl);
  color: var(--color-gray-600);
  text-align: center;
}

.no-bookmark-icon {
  font-size: var(--font-size-3xl);
  opacity: 0.7;
}

/* Responsive */
@media (max-width: 768px) {
  .bookmark-meta {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .meta-item {
    justify-content: space-between;
  }
  
  .action-buttons {
    justify-content: center;
  }
  
  .bookmark-url {
    font-size: var(--font-size-sm);
  }
}
</style>