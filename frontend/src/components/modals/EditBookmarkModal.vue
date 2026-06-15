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
        <label for="edit-shareTo" class="form-label">Share Destination *</label>
        <AppInput
          id="edit-shareTo"
          v-model="formData.shareTo"
          placeholder="Enter recipient or select from previous destinations"
          :error="errors.shareTo"
          list="share-destinations-edit"
          @input="clearError('shareTo')"
        />
        <datalist id="share-destinations-edit">
          <option v-for="destination in availableShareDestinations" :key="destination" :value="destination" />
          <!-- Common destination suggestions -->
          <option value="Team Slack" />
          <option value="Newsletter" />
          <option value="Dev Blog" />
          <option value="Social Media" />
          <option value="Email" />
          <option value="Unassigned" />
        </datalist>
        
        <!-- Quick Selection Chips -->
        <div v-if="availableShareDestinations.length > 0" class="destination-chips">
          <div class="chips-label">Previous destinations:</div>
          <div class="chips-container">
            <button
              v-for="destination in availableShareDestinations.slice(0, 5)"
              :key="destination"
              type="button"
              class="destination-chip"
              @click="selectShareDestination(destination)"
            >
              {{ destination }}
            </button>
          </div>
        </div>
      </div>

      <!-- Tags -->
      <div class="form-group">
        <label class="form-label">Tags</label>
        <div class="tag-input-wrapper" @click="focusTagInput">
          <button
            v-for="tag in formData.tags"
            :key="tag"
            type="button"
            class="tag-pill-edit"
            @click.stop="removeTag(tag)"
          >#{{ tag }} ×</button>
          <input
            ref="tagInputEl"
            v-model="tagInputValue"
            class="tag-text-input"
            placeholder="Add tag..."
            @keydown.enter.prevent="addTagFromInput"
            @keydown.exact.prevent.comma="addTagFromInput"
            @keydown.backspace="handleTagBackspace"
          />
        </div>
        <div v-if="suggestedTags.length > 0" class="tag-suggestions">
          <button
            v-for="tag in suggestedTags.slice(0, 12)"
            :key="tag"
            type="button"
            class="tag-suggestion-chip"
            @click="addTag(tag)"
          >#{{ tag }}</button>
        </div>
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
import { storeToRefs } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'

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

// Store access
const bookmarkStore = useBookmarkStore()
const { availableShareDestinations, availableTags } = storeToRefs(bookmarkStore)

// Form state
const formData = ref<FormData | null>(null)
const errors = ref<Record<string, string>>({})
const isSubmitting = ref(false)
const showContentEditor = ref(false)
const showAdvanced = ref(false)
const tagInputValue = ref('')
const tagInputEl = ref<HTMLInputElement | null>(null)
const customProperties = ref<CustomProperty[]>([])

const suggestedTags = computed(() =>
  availableTags.value.filter(t => !formData.value?.tags?.includes(t))
)

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
      formData.value = { ...newBookmark, tags: newBookmark.tags ? [...newBookmark.tags] : [] }
      tagInputValue.value = ''
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
  
  if (formData.value.action === 'share' && !formData.value.shareTo) {
    newErrors.shareTo = 'Share destination is required'
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
const focusTagInput = () => tagInputEl.value?.focus()

const addTagFromInput = () => {
  if (!formData.value) return
  const tag = tagInputValue.value.trim().replace(/,$/, '')
  if (!tag) return
  if (!formData.value.tags) formData.value.tags = []
  if (!formData.value.tags.includes(tag)) formData.value.tags.push(tag)
  tagInputValue.value = ''
}

const addTag = (tag: string) => {
  if (!formData.value) return
  if (!formData.value.tags) formData.value.tags = []
  if (!formData.value.tags.includes(tag)) formData.value.tags.push(tag)
}

const removeTag = (tagToRemove: string) => {
  if (!formData.value?.tags) return
  formData.value.tags = formData.value.tags.filter(t => t !== tagToRemove)
}

const handleTagBackspace = () => {
  if (tagInputValue.value === '' && formData.value?.tags?.length) {
    formData.value.tags.pop()
  }
}

// Custom properties handling
const addCustomProperty = () => {
  customProperties.value.push({ key: '', value: '' })
}

const removeCustomProperty = (index: number) => {
  customProperties.value.splice(index, 1)
}

const selectShareDestination = (destination: string) => {
  if (formData.value) {
    formData.value.shareTo = destination
    clearError('shareTo')
  }
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
  formData.value = null
  errors.value = {}
  isSubmitting.value = false
  showContentEditor.value = false
  showAdvanced.value = false
  tagInputValue.value = ''
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
.tag-input-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  cursor: text;
  min-height: 42px;
  align-items: center;
  transition: border-color var(--transition-fast);
}

.tag-input-wrapper:focus-within {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.tag-pill-edit {
  background: var(--color-primary-light, #ede9fe);
  color: var(--color-primary);
  border: none;
  padding: 2px 8px;
  border-radius: var(--radius-xl);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  white-space: nowrap;
  transition: background var(--transition-fast);
}

.tag-pill-edit:hover {
  background: var(--color-primary);
  color: white;
}

.tag-text-input {
  border: none;
  outline: none;
  font-size: var(--font-size-sm);
  background: transparent;
  min-width: 120px;
  flex: 1;
  padding: 2px 0;
  font-family: inherit;
}

.tag-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-xs);
}

.tag-suggestion-chip {
  background: var(--color-gray-100);
  color: var(--color-gray-600);
  border: 1px solid var(--border-light);
  padding: 2px 8px;
  border-radius: var(--radius-xl);
  font-size: var(--font-size-xs);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tag-suggestion-chip:hover {
  background: var(--color-primary-light, #ede9fe);
  color: var(--color-primary);
  border-color: var(--color-primary);
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

/* Share Destination Chips */
.destination-chips {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--border-light);
}

.chips-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-600);
  margin-bottom: var(--spacing-sm);
}

.chips-container {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.destination-chip {
  background: var(--color-gray-100);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xs) var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
  cursor: pointer;
  transition: var(--transition-fast);
}

.destination-chip:hover {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
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