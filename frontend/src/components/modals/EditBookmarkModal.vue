<template>
  <AppModal
    :show="show"
    title="Edit Bookmark"
    size="lg"
    @update:show="$emit('update:show', $event)"
    @close="handleClose"
  >
    <form v-if="formData" @submit.prevent="handleSubmit" class="edit-bookmark-form">
      <!-- URL Input -->
      <div class="form-group">
        <label for="edit-url" class="form-label">URL *</label>
        <AppInput
          id="edit-url"
          v-model="formData.url"
          type="url"
          placeholder="https://example.com/article"
          :error="errors.url"
          required
          @input="clearError('url')"
          @blur="validateUrl"
        />
      </div>

      <!-- Title Input -->
      <div class="form-group">
        <label for="edit-title" class="form-label">Title *</label>
        <AppInput
          id="edit-title"
          v-model="formData.title"
          placeholder="Enter bookmark title"
          :error="errors.title"
          required
          @input="clearError('title')"
        />
      </div>

      <!-- Description Input -->
      <div class="form-group">
        <label for="edit-description" class="form-label">Description</label>
        <textarea
          id="edit-description"
          v-model="formData.description"
          class="form-textarea"
          placeholder="Optional description or notes"
          rows="3"
          @input="clearError('description')"
        />
      </div>

      <!-- Content Input (if editing content is allowed) -->
      <div v-if="showContentEditor" class="form-group">
        <label for="edit-content" class="form-label">
          Content
          <button
            type="button"
            class="toggle-content"
            @click="showContentEditor = false"
          >
            Hide Editor
          </button>
        </label>
        <textarea
          id="edit-content"
          v-model="formData.content"
          class="form-textarea content-textarea"
          placeholder="Page content or notes"
          rows="6"
        />
      </div>
      <div v-else class="form-group">
        <div class="content-toggle">
          <span class="form-label">Content</span>
          <button
            type="button"
            class="toggle-content"
            @click="showContentEditor = true"
          >
            Show Editor
          </button>
        </div>
        <div class="content-preview">
          {{ formData.content ? formData.content.substring(0, 200) + '...' : 'No content' }}
        </div>
      </div>

      <!-- Action Selection -->
      <div class="form-group">
        <label class="form-label">Action</label>
        <div class="action-grid">
          <label
            v-for="action in actionOptions"
            :key="action.value"
            class="action-option"
            :class="{ active: formData.action === action.value }"
          >
            <input
              type="radio"
              :value="action.value"
              v-model="formData.action"
              @change="handleActionChange"
            />
            <div class="action-content">
              <div class="action-icon">{{ action.icon }}</div>
              <div class="action-text">
                <div class="action-label">{{ action.label }}</div>
                <div class="action-description">{{ action.description }}</div>
              </div>
            </div>
          </label>
        </div>
      </div>

      <!-- Topic/Project Selection (conditional) -->
      <div v-if="formData.action === 'working'" class="form-group">
        <label for="edit-topic" class="form-label">Project/Topic *</label>
        <div class="topic-input-container">
          <AppInput
            id="edit-topic"
            v-model="formData.topic"
            placeholder="Enter project name or select existing"
            :error="errors.topic"
            list="existing-topics-edit"
            @input="clearError('topic')"
          />
          <datalist id="existing-topics-edit">
            <option v-for="topic in existingTopics" :key="topic" :value="topic" />
          </datalist>
        </div>
        <div class="existing-topics">
          <div class="existing-topics-label">Existing Projects:</div>
          <div class="topic-chips">
            <button
              v-for="topic in existingTopics"
              :key="topic"
              type="button"
              class="topic-chip"
              @click="selectTopic(topic)"
            >
              {{ topic }}
            </button>
          </div>
        </div>
      </div>

      <!-- Share Destination (conditional) -->
      <div v-if="formData.action === 'share'" class="form-group">
        <label for="edit-shareTo" class="form-label">Share Destination</label>
        <select
          id="edit-shareTo"
          v-model="formData.shareTo"
          class="form-select"
        >
          <option value="">Select destination</option>
          <option value="Team Slack">📢 Team Slack</option>
          <option value="Newsletter">📧 Newsletter</option>
          <option value="Dev Blog">📝 Dev Blog</option>
          <option value="Unassigned">📤 Unassigned</option>
        </select>
      </div>

      <!-- Metadata Section -->
      <div class="form-group metadata-section">
        <h4 class="metadata-title">Metadata</h4>
        <div class="metadata-grid">
          <div class="metadata-item">
            <span class="metadata-label">Domain:</span>
            <span class="metadata-value">{{ formData.domain }}</span>
          </div>
          <div class="metadata-item">
            <span class="metadata-label">Added:</span>
            <span class="metadata-value">{{ formatDate(formData.timestamp) }}</span>
          </div>
          <div class="metadata-item">
            <span class="metadata-label">Age:</span>
            <span class="metadata-value">{{ formData.age }}</span>
          </div>
          <div class="metadata-item">
            <span class="metadata-label">ID:</span>
            <span class="metadata-value">{{ formData.id }}</span>
          </div>
        </div>
      </div>

      <!-- Advanced Options (collapsible) -->
      <div class="form-group">
        <button
          type="button"
          class="advanced-toggle"
          @click="showAdvanced = !showAdvanced"
        >
          {{ showAdvanced ? '▼' : '▶' }} Advanced Options
        </button>
        
        <div v-if="showAdvanced" class="advanced-options">
          <!-- Tags Input -->
          <div class="advanced-field">
            <label for="edit-tags" class="form-label">Tags (comma-separated)</label>
            <AppInput
              id="edit-tags"
              v-model="tagsInput"
              placeholder="tag1, tag2, tag3"
              @input="updateTags"
            />
            <div v-if="formData.tags && formData.tags.length > 0" class="current-tags">
              <span
                v-for="tag in formData.tags"
                :key="tag"
                class="tag-chip"
              >
                #{{ tag }}
                <button
                  type="button"
                  class="tag-remove"
                  @click="removeTag(tag)"
                >
                  ×
                </button>
              </span>
            </div>
          </div>

          <!-- Custom Fields -->
          <div class="advanced-field">
            <label class="form-label">Custom Properties</label>
            <div class="custom-properties">
              <div
                v-for="(prop, index) in customProperties"
                :key="index"
                class="custom-property"
              >
                <AppInput
                  v-model="prop.key"
                  placeholder="Property name"
                  size="sm"
                />
                <AppInput
                  v-model="prop.value"
                  placeholder="Property value"
                  size="sm"
                />
                <button
                  type="button"
                  class="remove-property"
                  @click="removeCustomProperty(index)"
                >
                  ×
                </button>
              </div>
              <button
                type="button"
                class="add-property"
                @click="addCustomProperty"
              >
                + Add Property
              </button>
            </div>
          </div>
        </div>
      </div>
    </form>

    <div v-else class="no-bookmark">
      <div class="no-bookmark-icon">❌</div>
      <p>No bookmark data available for editing</p>
    </div>

    <template #footer>
      <div class="footer-left">
        <AppButton
          variant="danger"
          size="sm"
          @click="handleDelete"
        >
          🗑️ Delete
        </AppButton>
      </div>
      <div class="footer-right">
        <AppButton variant="secondary" @click="handleClose">
          Cancel
        </AppButton>
        <AppButton
          variant="primary"
          :loading="isSubmitting"
          @click="handleSubmit"
        >
          Save Changes
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Bookmark } from '@/types'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

interface Props {
  show: boolean
  bookmark?: Bookmark | null
  existingTopics: string[]
}

interface FormData extends Bookmark {
  tags?: string[]
}

interface CustomProperty {
  key: string
  value: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'submit': [data: FormData]
  'delete': [bookmarkId: string]
}>()

// Form state
const formData = ref<FormData | null>(null)
const errors = ref<Record<string, string>>({})
const isSubmitting = ref(false)
const showContentEditor = ref(false)
const showAdvanced = ref(false)
const tagsInput = ref('')
const customProperties = ref<CustomProperty[]>([])

// Action options
const actionOptions = [
  {
    value: 'read-later',
    label: 'Read Later',
    icon: '📚',
    description: 'Save for later review'
  },
  {
    value: 'working',
    label: 'Working',
    icon: '🚀',
    description: 'Active project resource'
  },
  {
    value: 'share',
    label: 'Share',
    icon: '📤',
    description: 'Ready to share with others'
  },
  {
    value: 'archived',
    label: 'Archive',
    icon: '📦',
    description: 'Completed or finished'
  }
]

// Watch for bookmark changes
watch(
  () => props.bookmark,
  (newBookmark) => {
    if (newBookmark) {
      formData.value = { ...newBookmark }
      tagsInput.value = newBookmark.tags ? newBookmark.tags.join(', ') : ''
      // Reset advanced options
      showAdvanced.value = false
      showContentEditor.value = false
      customProperties.value = []
    }
  },
  { immediate: true }
)

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

// Validation
const validateUrl = () => {
  if (!formData.value?.url) {
    errors.value.url = 'URL is required'
    return false
  }
  
  try {
    new URL(formData.value.url)
    errors.value.url = ''
    return true
  } catch {
    errors.value.url = 'Please enter a valid URL'
    return false
  }
}

const validateForm = (): boolean => {
  if (!formData.value) return false
  
  const newErrors: Record<string, string> = {}
  
  if (!formData.value.url) {
    newErrors.url = 'URL is required'
  } else {
    try {
      new URL(formData.value.url)
    } catch {
      newErrors.url = 'Please enter a valid URL'
    }
  }
  
  if (!formData.value.title) {
    newErrors.title = 'Title is required'
  }
  
  if (formData.value.action === 'working' && !formData.value.topic) {
    newErrors.topic = 'Project/Topic is required for working items'
  }
  
  errors.value = newErrors
  return Object.keys(newErrors).length === 0
}

// Event handlers
const clearError = (field: string) => {
  if (errors.value[field]) {
    delete errors.value[field]
  }
}

const handleActionChange = () => {
  if (!formData.value) return
  
  // Clear topic when switching away from working
  if (formData.value.action !== 'working') {
    formData.value.topic = ''
  }
  // Clear shareTo when switching away from share
  if (formData.value.action !== 'share') {
    formData.value.shareTo = ''
  }
}

const selectTopic = (topic: string) => {
  if (formData.value) {
    formData.value.topic = topic
    clearError('topic')
  }
}

// Tags handling
const updateTags = () => {
  if (!formData.value) return
  
  const tags = tagsInput.value
    .split(',')
    .map(tag => tag.trim())
    .filter(tag => tag.length > 0)
  
  formData.value.tags = tags.length > 0 ? tags : undefined
}

const removeTag = (tagToRemove: string) => {
  if (!formData.value?.tags) return
  
  formData.value.tags = formData.value.tags.filter(tag => tag !== tagToRemove)
  tagsInput.value = formData.value.tags.join(', ')
}

// Custom properties handling
const addCustomProperty = () => {
  customProperties.value.push({ key: '', value: '' })
}

const removeCustomProperty = (index: number) => {
  customProperties.value.splice(index, 1)
}

const handleSubmit = async () => {
  if (!validateForm() || !formData.value) return
  
  isSubmitting.value = true
  
  try {
    // Update domain if URL changed
    const domain = new URL(formData.value.url).hostname
    formData.value.domain = domain
    
    // Process custom properties
    const customProps = customProperties.value
      .filter(prop => prop.key && prop.value)
      .reduce((acc, prop) => {
        acc[prop.key] = prop.value
        return acc
      }, {} as Record<string, string>)
    
    const submitData = {
      ...formData.value,
      ...customProps
    }
    
    emit('submit', submitData)
    handleClose()
  } catch (error) {
    console.error('Failed to submit bookmark:', error)
    errors.value.submit = 'Failed to save changes. Please try again.'
  } finally {
    isSubmitting.value = false
  }
}

const handleDelete = () => {
  if (props.bookmark?.id) {
    emit('delete', props.bookmark.id)
  }
}

const handleClose = () => {
  // Reset form state
  formData.value = null
  errors.value = {}
  isSubmitting.value = false
  showContentEditor.value = false
  showAdvanced.value = false
  tagsInput.value = ''
  customProperties.value = []
  
  emit('update:show', false)
}
</script>

<style scoped>
.edit-bookmark-form {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.form-label {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-700);
  font-size: var(--font-size-sm);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-textarea {
  padding: var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  font-family: inherit;
  font-size: var(--font-size-base);
  resize: vertical;
  min-height: 80px;
  transition: var(--transition-fast);
}

.form-textarea.content-textarea {
  min-height: 120px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: var(--font-size-sm);
}

.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.form-select {
  padding: var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  background: white;
  transition: var(--transition-fast);
}

.form-select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

/* Content toggle */
.content-toggle {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toggle-content {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  padding: var(--spacing-xs);
  border-radius: var(--radius-sm);
}

.toggle-content:hover {
  background: var(--color-gray-100);
}

.content-preview {
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  font-style: italic;
}

/* Action Grid - reuse from AddBookmarkModal */
.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.action-option {
  position: relative;
  display: block;
  cursor: pointer;
  border: 2px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: var(--spacing-md);
  transition: var(--transition-fast);
  background: white;
}

.action-option:hover {
  border-color: var(--color-primary);
  background: var(--color-gray-50);
}

.action-option.active {
  border-color: var(--color-primary);
  background: rgba(66, 153, 225, 0.05);
}

.action-option input[type="radio"] {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.action-content {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-md);
}

.action-icon {
  font-size: var(--font-size-xl);
  flex-shrink: 0;
}

.action-text {
  flex: 1;
}

.action-label {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  margin-bottom: var(--spacing-xs);
}

.action-description {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  line-height: var(--line-height-relaxed);
}

/* Topic selection - reuse from AddBookmarkModal */
.topic-input-container {
  position: relative;
}

.existing-topics {
  margin-top: var(--spacing-md);
}

.existing-topics-label {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  margin-bottom: var(--spacing-sm);
}

.topic-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.topic-chip {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  border: 1px solid var(--border-light);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: var(--transition-fast);
}

.topic-chip:hover {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

/* Metadata section */
.metadata-section {
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
}

.metadata-title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  margin: 0 0 var(--spacing-md) 0;
  color: var(--color-gray-700);
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-sm);
}

.metadata-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--font-size-sm);
}

.metadata-label {
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-600);
}

.metadata-value {
  color: var(--color-gray-800);
}

/* Advanced options */
.advanced-toggle {
  background: none;
  border: none;
  color: var(--color-gray-600);
  font-size: var(--font-size-sm);
  cursor: pointer;
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  transition: var(--transition-fast);
}

.advanced-toggle:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-800);
}

.advanced-options {
  margin-top: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.advanced-field {
  margin-bottom: var(--spacing-lg);
}

.advanced-field:last-child {
  margin-bottom: 0;
}

/* Tags */
.current-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-sm);
}

.tag-chip {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-xl);
  font-size: var(--font-size-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.tag-remove {
  background: none;
  border: none;
  color: var(--color-gray-500);
  cursor: pointer;
  font-size: var(--font-size-sm);
  padding: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tag-remove:hover {
  background: var(--color-gray-300);
  color: var(--color-gray-700);
}

/* Custom properties */
.custom-properties {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.custom-property {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.remove-property {
  background: var(--color-red-100);
  color: var(--color-red-600);
  border: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  font-size: var(--font-size-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-property:hover {
  background: var(--color-red-200);
}

.add-property {
  background: var(--color-gray-100);
  color: var(--color-gray-600);
  border: 1px dashed var(--border-light);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-size-sm);
  margin-top: var(--spacing-sm);
}

.add-property:hover {
  background: var(--color-gray-200);
  border-color: var(--color-gray-400);
}

/* Footer layout */
.footer-left {
  margin-right: auto;
}

.footer-right {
  display: flex;
  gap: var(--spacing-md);
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
  .action-grid {
    grid-template-columns: 1fr;
  }
  
  .metadata-grid {
    grid-template-columns: 1fr;
  }
  
  .custom-property {
    flex-direction: column;
    align-items: stretch;
  }
  
  .footer-left,
  .footer-right {
    display: flex;
    gap: var(--spacing-sm);
  }
}
</style>