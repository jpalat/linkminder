<template>
  <div class="share-groups">
    <div v-if="groups.length === 0" class="empty-state">
      <h3>No items ready to share</h3>
      <p>Mark bookmarks as "share" to see them organized by destination here.</p>
    </div>
    
    <div v-else class="groups-list">
      <div
        v-for="group in groups"
        :key="group.destination"
        class="share-group"
      >
        <div class="group-header">
          <div class="group-title">
            {{ group.icon }} {{ group.destination }} ({{ group.items.length }} items)
          </div>
          <div class="group-actions">
            <AppButton size="sm" variant="success" @click="shareGroup(group)">
              Share All
            </AppButton>
            <AppButton size="sm" variant="secondary" @click="showCopyOptions(group)">
              Copy
            </AppButton>
          </div>
        </div>
        
        <div class="group-items">
          <div
            v-for="item in group.items"
            :key="item.id"
            class="share-item"
          >
            <div class="item-content">
              <h4 class="item-title">{{ item.title }}</h4>
              <a :href="item.url" class="item-url" target="_blank" rel="noopener noreferrer">
                {{ item.url }}
              </a>
              <div class="item-meta">
                <span class="item-domain">{{ item.domain }}</span>
                <span class="item-time">{{ item.age }}</span>
              </div>
            </div>
            
            <div class="item-actions">
              <AppButton size="xs" variant="info" @click="previewItem(item)">
                👁️
              </AppButton>
              <AppButton size="xs" variant="secondary" @click="editItem(item)">
                ✏️
              </AppButton>
              <AppButton size="xs" variant="success" @click="completeItem(item)">
                Complete
              </AppButton>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Copy Options Modal (simplified) -->
    <div v-if="showCopyModal" class="copy-modal" @click="showCopyModal = false">
      <div class="copy-content" @click.stop>
        <h3>Copy {{ selectedGroup?.destination }} Items</h3>
        <div class="copy-formats">
          <AppButton @click="copyAs('markdown')">
            📝 Markdown
          </AppButton>
          <AppButton @click="copyAs('email')">
            📧 Email
          </AppButton>
          <AppButton @click="copyAs('slack')">
            💬 Slack
          </AppButton>
          <AppButton @click="copyAs('plain')">
            📄 Plain Text
          </AppButton>
        </div>
        <AppButton variant="secondary" @click="showCopyModal = false">
          Cancel
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useClipboard } from '@vueuse/core'
import type { ShareGroup, Bookmark } from '@/types'
import AppButton from '@/components/ui/AppButton.vue'

interface Props {
  groups: ShareGroup[]
}

defineProps<Props>()

const emit = defineEmits<{
  'share-group': [group: ShareGroup]
  'preview-item': [item: Bookmark]
  'edit-item': [item: Bookmark]
  'complete-item': [item: Bookmark]
}>()

// Copy functionality
const { copy } = useClipboard()
const showCopyModal = ref(false)
const selectedGroup = ref<ShareGroup | null>(null)

const showCopyOptions = (group: ShareGroup) => {
  selectedGroup.value = group
  showCopyModal.value = true
}

const copyAs = async (format: string) => {
  if (!selectedGroup.value) return
  
  const content = formatGroupContent(selectedGroup.value, format)
  await copy(content)
  showCopyModal.value = false
  
  // Show success feedback (you could use a toast notification here)
  console.log(`Copied ${selectedGroup.value.destination} items as ${format}`)
}

const formatGroupContent = (group: ShareGroup, format: string): string => {
  const { destination, items } = group
  
  switch (format) {
    case 'markdown':
      return `# ${destination}\n\n${items.map(item => 
        `- [${item.title}](${item.url})${item.description ? ` - ${item.description}` : ''}`
      ).join('\n')}`
    
    case 'email':
      return `${destination}\n\n${items.map(item => 
        `${item.title}\n${item.url}${item.description ? `\n${item.description}` : ''}\n`
      ).join('\n')}`
    
    case 'slack':
      return items.map(item => 
        `<${item.url}|${item.title}>${item.description ? ` - ${item.description}` : ''}`
      ).join('\n')
    
    case 'plain':
    default:
      return `${destination}\n\n${items.map(item => 
        `${item.title}\n${item.url}${item.description ? `\n${item.description}` : ''}\n`
      ).join('\n')}`
  }
}

// Event handlers
const shareGroup = (group: ShareGroup) => {
  emit('share-group', group)
}

const previewItem = (item: Bookmark) => {
  emit('preview-item', item)
}

const editItem = (item: Bookmark) => {
  emit('edit-item', item)
}

const completeItem = (item: Bookmark) => {
  emit('complete-item', item)
}
</script>

<style scoped>
.share-groups {
  width: 100%;
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.share-group {
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
}

.group-header {
  background: var(--color-gray-50);
  padding: var(--spacing-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--border-light);
}

.group-title {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  font-size: var(--font-size-lg);
}

.group-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.group-items {
  background: white;
  padding: var(--spacing-lg);
}

.share-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-md);
  background: var(--bg-card-hover);
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-md);
  gap: var(--spacing-md);
  transition: var(--transition-fast);
}

.share-item:last-child {
  margin-bottom: 0;
}

.share-item:hover {
  background: var(--color-gray-200);
  transform: translateY(-1px);
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-weight: var(--font-weight-semibold);
  margin-bottom: var(--spacing-xs);
  color: var(--color-gray-800);
  font-size: var(--font-size-base);
}

.item-url {
  color: var(--color-primary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-sm);
  text-decoration: none;
  display: block;
  word-break: break-all;
}

.item-url:hover {
  text-decoration: underline;
}

.item-meta {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  display: flex;
  gap: var(--spacing-md);
}

.item-domain {
  background: var(--color-gray-200);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-xs);
}

.item-actions {
  display: flex;
  gap: var(--spacing-xs);
  flex-shrink: 0;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.share-item:hover .item-actions {
  opacity: 1;
}

/* Copy Modal */
.copy-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
}

.copy-content {
  background: white;
  padding: var(--spacing-2xl);
  border-radius: var(--radius-xl);
  max-width: 400px;
  width: 90%;
  box-shadow: var(--shadow-xl);
}

.copy-content h3 {
  margin-bottom: var(--spacing-xl);
  text-align: center;
  color: var(--color-gray-800);
}

.copy-formats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xl);
}

/* Responsive */
@media (max-width: 768px) {
  .group-header {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .share-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .item-actions {
    opacity: 1;
    justify-content: flex-end;
    margin-top: var(--spacing-sm);
  }
  
  .copy-formats {
    grid-template-columns: 1fr;
  }
}
</style>
