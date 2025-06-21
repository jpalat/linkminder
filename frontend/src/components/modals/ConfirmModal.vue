<template>
  <AppModal
    :show="show"
    :title="title"
    size="sm"
    @update:show="$emit('update:show', $event)"
    @close="handleCancel"
  >
    <div class="confirm-content">
      <!-- Icon -->
      <div class="confirm-icon" :class="iconClass">
        {{ icon }}
      </div>

      <!-- Message -->
      <div class="confirm-message">
        <h3 class="confirm-title">{{ title }}</h3>
        <p class="confirm-description">{{ message }}</p>
        
        <!-- Details (if provided) -->
        <div v-if="details" class="confirm-details">
          <div class="details-content">
            {{ details }}
          </div>
        </div>

        <!-- Item list (for batch operations) -->
        <div v-if="items && items.length > 0" class="confirm-items">
          <div class="items-header">
            {{ items.length }} item{{ items.length > 1 ? 's' : '' }} will be affected:
          </div>
          <div class="items-list">
            <div
              v-for="item in displayItems"
              :key="item.id"
              class="item-entry"
            >
              <span class="item-title">{{ item.title }}</span>
              <span class="item-domain">{{ item.domain }}</span>
            </div>
            <div v-if="items.length > maxDisplayItems" class="items-more">
              ...and {{ items.length - maxDisplayItems }} more
            </div>
          </div>
        </div>

        <!-- Warning text -->
        <div v-if="isDestructive" class="warning-text">
          ⚠️ This action cannot be undone.
        </div>
      </div>
    </div>

    <template #footer>
      <AppButton
        variant="secondary"
        @click="handleCancel"
      >
        {{ cancelText }}
      </AppButton>
      <AppButton
        :variant="confirmVariant"
        :loading="isProcessing"
        @click="handleConfirm"
      >
        {{ confirmText }}
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Bookmark } from '@/types'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'

export interface ConfirmationConfig {
  type: 'delete' | 'archive' | 'move' | 'share' | 'custom'
  title?: string
  message?: string
  details?: string
  confirmText?: string
  cancelText?: string
  isDestructive?: boolean
  items?: Bookmark[]
}

interface Props {
  show: boolean
  config: ConfirmationConfig
  isProcessing?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isProcessing: false
})

const emit = defineEmits<{
  'update:show': [value: boolean]
  'confirm': []
  'cancel': []
}>()

const maxDisplayItems = 5

// Computed properties based on config type
const title = computed(() => {
  if (props.config.title) return props.config.title
  
  const defaultTitles = {
    delete: 'Delete Bookmark',
    archive: 'Archive Bookmark',
    move: 'Move Bookmark',
    share: 'Share Bookmark',
    custom: 'Confirm Action'
  }
  
  const baseTitle = defaultTitles[props.config.type]
  const count = props.config.items?.length || 0
  
  if (count > 1) {
    return baseTitle.replace('Bookmark', `${count} Bookmarks`)
  }
  
  return baseTitle
})

const message = computed(() => {
  if (props.config.message) return props.config.message
  
  const count = props.config.items?.length || 0
  const isPlural = count > 1
  const bookmarkText = isPlural ? 'bookmarks' : 'bookmark'
  
  const defaultMessages = {
    delete: `Are you sure you want to delete ${isPlural ? 'these' : 'this'} ${bookmarkText}?`,
    archive: `Are you sure you want to archive ${isPlural ? 'these' : 'this'} ${bookmarkText}?`,
    move: `Are you sure you want to move ${isPlural ? 'these' : 'this'} ${bookmarkText}?`,
    share: `Are you sure you want to share ${isPlural ? 'these' : 'this'} ${bookmarkText}?`,
    custom: 'Are you sure you want to proceed?'
  }
  
  return defaultMessages[props.config.type]
})

const confirmText = computed(() => {
  if (props.config.confirmText) return props.config.confirmText
  
  const defaultTexts = {
    delete: 'Delete',
    archive: 'Archive',
    move: 'Move',
    share: 'Share',
    custom: 'Confirm'
  }
  
  return defaultTexts[props.config.type]
})

const cancelText = computed(() => {
  return props.config.cancelText || 'Cancel'
})

const isDestructive = computed(() => {
  return props.config.isDestructive ?? props.config.type === 'delete'
})

const confirmVariant = computed((): 'primary' | 'danger' | 'info' | 'success' => {
  if (isDestructive.value) return 'danger'
  
  const variants = {
    delete: 'danger',
    archive: 'info',
    move: 'primary',
    share: 'success',
    custom: 'primary'
  } as const
  
  return variants[props.config.type] || 'primary'
})

const icon = computed(() => {
  const icons = {
    delete: '🗑️',
    archive: '📦',
    move: '🔄',
    share: '📤',
    custom: '❓'
  }
  
  return icons[props.config.type]
})

const iconClass = computed(() => {
  const classes = {
    delete: 'icon-danger',
    archive: 'icon-warning',
    move: 'icon-info',
    share: 'icon-success',
    custom: 'icon-info'
  }
  
  return classes[props.config.type]
})

const details = computed(() => props.config.details)

const displayItems = computed(() => {
  if (!props.config.items) return []
  return props.config.items.slice(0, maxDisplayItems)
})

const items = computed(() => props.config.items)

// Event handlers
const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
  emit('update:show', false)
}
</script>

<style scoped>
.confirm-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-lg);
  text-align: center;
  padding: var(--spacing-lg) 0;
}

/* Icon */
.confirm-icon {
  font-size: var(--font-size-4xl);
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.icon-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #d97706;
}

.icon-success {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.icon-info {
  background: rgba(59, 130, 246, 0.1);
  color: #2563eb;
}

/* Message */
.confirm-message {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.confirm-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  margin: 0;
}

.confirm-description {
  font-size: var(--font-size-base);
  color: var(--color-gray-600);
  line-height: var(--line-height-relaxed);
  margin: 0;
}

/* Details */
.confirm-details {
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-top: var(--spacing-sm);
}

.details-content {
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
  line-height: var(--line-height-relaxed);
  text-align: left;
}

/* Items list */
.confirm-items {
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-top: var(--spacing-sm);
  text-align: left;
}

.items-header {
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-700);
  margin-bottom: var(--spacing-sm);
  font-size: var(--font-size-sm);
}

.items-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.item-entry {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-xs) var(--spacing-sm);
  background: white;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-light);
  font-size: var(--font-size-sm);
}

.item-title {
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-800);
  flex: 1;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: var(--spacing-md);
}

.item-domain {
  color: var(--color-gray-500);
  font-size: var(--font-size-xs);
  flex-shrink: 0;
}

.items-more {
  font-style: italic;
  color: var(--color-gray-500);
  font-size: var(--font-size-xs);
  text-align: center;
  padding: var(--spacing-xs);
}

/* Warning */
.warning-text {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  border: 1px solid rgba(239, 68, 68, 0.2);
  margin-top: var(--spacing-sm);
}

/* Responsive */
@media (max-width: 768px) {
  .confirm-content {
    padding: var(--spacing-md) 0;
  }
  
  .confirm-icon {
    width: 60px;
    height: 60px;
    font-size: var(--font-size-3xl);
  }
  
  .confirm-title {
    font-size: var(--font-size-lg);
  }
  
  .item-entry {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-xs);
  }
  
  .item-title {
    margin-right: 0;
  }
}
</style>