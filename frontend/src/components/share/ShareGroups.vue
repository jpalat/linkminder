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
        <!-- Group Header -->
        <div class="share-group-header">
          <div class="share-group-info">
            <div class="share-group-title">{{ group.icon }} {{ group.destination }}</div>
            <div class="share-group-count">{{ group.items.length }}</div>
          </div>
          <div class="share-group-actions">
            <AppButton 
              size="sm" 
              variant="success" 
              @click="shareGroup(group)"
            >
              {{ getShareButtonText(group.destination) }}
            </AppButton>
            <AppButton 
              size="sm" 
              variant="primary" 
              @click="completeGroup(group)"
            >
              ✅ Complete
            </AppButton>
          </div>
        </div>
        
        <!-- Destination Info -->
        <div class="destination-info">
          <strong>Destination:</strong> {{ getDestinationDetails(group.destination) }}
        </div>
        
        <!-- Group Copy Section -->
        <div class="group-copy-section">
          <div class="group-copy-title">📋 Copy All {{ group.destination }} Items</div>
          <div class="group-copy-formats">
            <button 
              v-for="format in getGroupFormats()" 
              :key="format.key"
              class="group-copy-btn" 
              @click="copyGroupItems(group, format.key)"
            >
              {{ format.icon }} {{ format.label }}
            </button>
          </div>
        </div>
        
        <!-- Group Items -->
        <div class="share-group-items">
          <div
            v-for="item in group.items"
            :key="item.id"
            class="share-item"
          >
            <div class="share-item-header">
              <div class="item-info">
                <div class="item-title">{{ item.title }}</div>
                <div class="item-url">{{ item.url }}</div>
                <div class="item-meta">
                  <span>{{ item.domain }}</span>
                  <span>{{ item.age }}</span>
                </div>
              </div>
              <div class="item-actions">
                <button class="action-btn preview" @click="previewItem(item)">👁️</button>
                <button class="action-btn edit" @click="editItem(item)">✏️</button>
                <button class="action-btn share" @click="shareItem(item)">📤</button>
                <button class="action-btn archive" @click="archiveItem(item)">📦</button>
              </div>
            </div>
            
            <!-- Individual Item Format Buttons -->
            <div class="share-formats">
              <button 
                v-for="format in itemFormats"
                :key="format.key"
                class="format-btn" 
                @click="copyItemFormat(item, format.key)"
              >
                {{ format.icon }} {{ format.label }}
              </button>
              <button class="format-btn" @click="showItemPreview(item)">
                👁️ Preview
              </button>
            </div>
            
            <!-- Format Preview (expandable) -->
            <div v-if="item.id === previewItemId" class="format-preview">
              <div class="preview-content">
                {{ getItemPreview(item, previewFormat) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ShareGroup, Bookmark } from '@/types'
import AppButton from '@/components/ui/AppButton.vue'
import { useNotifications } from '@/composables/useNotifications'

interface Props {
  groups: ShareGroup[]
}

defineProps<Props>()

const emit = defineEmits<{
  'share-group': [group: ShareGroup]
  'complete-group': [group: ShareGroup]
  'preview-item': [item: Bookmark]
  'edit-item': [item: Bookmark]
  'share-item': [item: Bookmark]
  'archive-item': [item: Bookmark]
}>()

// Copy functionality - use fallback method that works reliably
const { success, error } = useNotifications()

// Reliable clipboard copy function using fallback method
const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    // Try modern Clipboard API first (if available and secure)
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
    
    // Fallback method using temporary textarea (works in more environments)
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.left = '-999999px'
    textarea.style.top = '-999999px'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    
    // Select and copy
    textarea.focus()
    textarea.select()
    textarea.setSelectionRange(0, 99999) // For mobile devices
    
    const successful = document.execCommand('copy')
    document.body.removeChild(textarea)
    
    return successful
  } catch (err) {
    console.error('Copy failed:', err)
    return false
  }
}
const previewItemId = ref<string | null>(null)
const previewFormat = ref('markdown')

// Format definitions
const itemFormats = [
  { key: 'rich-text', label: 'Rich Text', icon: '📄' },
  { key: 'markdown', label: 'Markdown', icon: '📝' },
  { key: 'plain', label: 'Plain Text', icon: '📃' }
]

// Group-specific format options
const getGroupFormats = () => {
  return [
    { key: 'rich-text', label: 'Rich Text', icon: '📄' },
    { key: 'markdown', label: 'Markdown', icon: '📝' },
    { key: 'plain', label: 'Plain Text', icon: '📃' }
  ]
}

const getShareButtonText = (destination: string): string => {
  const buttonTexts: Record<string, string> = {
    'Team Slack': '📤 Share to Slack',
    'Newsletter': '📧 Add to Newsletter',
    'Dev Blog': '📝 Add to Blog',
    'Unassigned': '📤 Share'
  }
  return buttonTexts[destination] || '📤 Share'
}

const getDestinationDetails = (destination: string): string => {
  const details: Record<string, string> = {
    'Team Slack': '#dev-resources channel • Webhook configured',
    'Newsletter': 'Weekly Dev Newsletter • Next send: Friday 9 AM',
    'Dev Blog': 'Company Tech Blog • Draft saved',
    'Unassigned': 'No destination configured'
  }
  return details[destination] || 'Custom destination'
}

// Formatting functions
const formatGroupContent = (group: ShareGroup, format: string): string => {
  const { destination, items } = group
  
  switch (format) {
    case 'rich-text':
      return `${destination}\n\n${items.map(item => 
        `${item.title}\n${item.url}${item.description ? `\n\n${item.description}` : ''}\n`
      ).join('\n')}`
    
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

const formatItemContent = (item: Bookmark, format: string): string => {
  switch (format) {
    case 'rich-text':
      // Rich text format suitable for pasting into rich text editors
      return `${item.title}\n${item.url}${item.description ? `\n\n${item.description}` : ''}`
    case 'markdown':
      return `[${item.title}](${item.url})${item.description ? ` - ${item.description}` : ''}`
    case 'plain':
      return `${item.title}\n${item.url}${item.description ? `\n${item.description}` : ''}`
    case 'email':
      return `${item.title}\n${item.url}${item.description ? `\n${item.description}` : ''}`
    case 'slack':
      return `<${item.url}|${item.title}>${item.description ? ` - ${item.description}` : ''}`
    default:
      return `${item.title}\n${item.url}${item.description ? `\n${item.description}` : ''}`
  }
}

// Event handlers
const shareGroup = (group: ShareGroup) => {
  emit('share-group', group)
}

const completeGroup = (group: ShareGroup) => {
  emit('complete-group', group)
}

const previewItem = (item: Bookmark) => {
  emit('preview-item', item)
}

const editItem = (item: Bookmark) => {
  emit('edit-item', item)
}

const shareItem = (item: Bookmark) => {
  emit('share-item', item)
}

const archiveItem = (item: Bookmark) => {
  emit('archive-item', item)
}

const copyGroupItems = async (group: ShareGroup, format: string) => {
  try {
    const content = formatGroupContent(group, format)
    console.log('Copying group content:', content) // Debug log
    
    const successful = await copyToClipboard(content)
    
    if (successful) {
      success(`Copied ${group.items.length} items as ${format} format`, {
        title: `${group.destination} Copied`
      })
    } else {
      error('Failed to copy to clipboard. Please try again.', {
        title: 'Copy Failed'
      })
    }
  } catch (err) {
    console.error('Failed to copy group items:', err)
    error('Failed to copy to clipboard. Please try again.', {
      title: 'Copy Failed'
    })
  }
}

const copyItemFormat = async (item: Bookmark, format: string) => {
  try {
    const content = formatItemContent(item, format)
    console.log('Copying item content:', content) // Debug log
    
    const successful = await copyToClipboard(content)
    
    if (successful) {
      success(`Copied "${item.title}" as ${format} format`, {
        title: 'Bookmark Copied'
      })
    } else {
      error('Failed to copy to clipboard. Please try again.', {
        title: 'Copy Failed'
      })
    }
  } catch (err) {
    console.error('Failed to copy item:', err)
    error('Failed to copy to clipboard. Please try again.', {
      title: 'Copy Failed'
    })
  }
}

const showItemPreview = (item: Bookmark) => {
  if (previewItemId.value === item.id) {
    previewItemId.value = null
  } else {
    previewItemId.value = item.id
    previewFormat.value = 'markdown'
  }
}

const getItemPreview = (item: Bookmark, format: string): string => {
  return formatItemContent(item, format)
}
</script>

<style scoped>
.share-groups {
  width: 100%;
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2xl);
}

.share-group {
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  background: white;
}

/* Group Header */
.share-group-header {
  background: #f8fafc;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--border-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.share-group-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.share-group-title {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  font-size: var(--font-size-lg);
}

.share-group-count {
  background: var(--color-primary);
  color: white;
  padding: 2px 8px;
  border-radius: var(--radius-xl);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  min-width: 20px;
  text-align: center;
}

.share-group-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* Destination Info */
.destination-info {
  background: #fffbeb;
  padding: var(--spacing-sm) var(--spacing-lg);
  border-bottom: 1px solid #fed7aa;
  font-size: var(--font-size-sm);
  color: #92400e;
}

/* Group Copy Section */
.group-copy-section {
  background: #f0f9ff;
  padding: var(--spacing-lg);
  border-bottom: 1px solid #bee3f8;
}

.group-copy-title {
  font-weight: var(--font-weight-semibold);
  color: #1e40af;
  margin-bottom: var(--spacing-md);
  font-size: var(--font-size-sm);
}

.group-copy-formats {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.group-copy-btn {
  background: white;
  color: #1e40af;
  border: 1px solid #bee3f8;
  padding: var(--spacing-xs) var(--spacing-md);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: var(--transition-fast);
}

.group-copy-btn:hover {
  background: #1e40af;
  color: white;
}

/* Share Group Items */
.share-group-items {
  padding: var(--spacing-lg);
}

.share-item {
  margin-bottom: var(--spacing-lg);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: var(--transition-fast);
}

.share-item:last-child {
  margin-bottom: 0;
}

.share-item:hover {
  box-shadow: var(--shadow-md);
}

/* Share Item Header */
.share-item-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-md);
  background: var(--bg-card-hover);
  gap: var(--spacing-md);
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-weight: var(--font-weight-semibold);
  margin-bottom: var(--spacing-xs);
  color: var(--color-gray-800);
  font-size: var(--font-size-base);
  line-height: var(--line-height-tight);
}

.item-url {
  color: var(--color-primary);
  font-size: var(--font-size-sm);
  margin-bottom: var(--spacing-xs);
  display: block;
  word-break: break-all;
  text-decoration: none;
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

.action-btn {
  background: var(--color-gray-200);
  color: var(--color-gray-700);
  border: none;
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  cursor: pointer;
  transition: var(--transition-fast);
}

.action-btn:hover {
  background: var(--color-gray-300);
}

.action-btn.preview {
  background: #e9d8fd;
  color: #553c9a;
}

.action-btn.edit {
  background: #fed7d7;
  color: #742a2a;
}

.action-btn.share {
  background: #c6f6d5;
  color: #22543d;
}

.action-btn.archive {
  background: #fed7d7;
  color: #742a2a;
}

/* Share Formats */
.share-formats {
  display: flex;
  gap: var(--spacing-xs);
  padding: var(--spacing-md);
  background: white;
  border-top: 1px solid var(--border-light);
  flex-wrap: wrap;
}

.format-btn {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  border: 1px solid var(--border-light);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: var(--transition-fast);
}

.format-btn:hover {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

/* Format Preview */
.format-preview {
  background: #f8fafc;
  border-top: 1px solid var(--border-light);
  padding: var(--spacing-md);
}

.preview-content {
  background: white;
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
  white-space: pre-wrap;
  word-break: break-word;
}

/* Responsive */
@media (max-width: 768px) {
  .share-group-header {
    flex-direction: column;
    gap: var(--spacing-md);
    align-items: stretch;
  }
  
  .share-group-info {
    justify-content: center;
  }
  
  .group-copy-formats {
    flex-direction: column;
  }
  
  .share-item-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .item-actions {
    opacity: 1;
    justify-content: flex-end;
    margin-top: var(--spacing-sm);
  }
  
  .share-formats {
    flex-direction: column;
  }
}
</style>
