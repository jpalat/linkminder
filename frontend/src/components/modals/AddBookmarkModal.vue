<template>
  <AppModal
    :show="show"
    title="Add New Bookmark"
    size="lg"
    @update:show="$emit('update:show', $event)"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="add-bookmark-form">
      <!-- URL Input -->
      <div class="form-group">
        <label for="url" class="form-label">URL *</label>
        <AppInput
          id="url"
          v-model="formData.url"
          type="url"
          placeholder="https://example.com/article"
          :error="errors.url"
          required
          @input="clearError('url')"
          @blur="validateUrl"
        />
        <div v-if="isLoadingMetadata" class="url-status">
          🔄 Loading page information...
        </div>
        <div v-if="urlPreview" class="url-preview">
          <div class="url-preview-title">{{ urlPreview.title }}</div>
          <div class="url-preview-description">{{ urlPreview.description }}</div>
        </div>
      </div>

      <!-- Title Input -->
      <div class="form-group">
        <label for="title" class="form-label">Title *</label>
        <AppInput
          id="title"
          v-model="formData.title"
          placeholder="Enter bookmark title"
          :error="errors.title"
          required
          @input="clearError('title')"
        />
      </div>

      <!-- Description Input -->
      <div class="form-group">
        <label for="description" class="form-label">Description</label>
        <textarea
          id="description"
          v-model="formData.description"
          class="form-textarea"
          placeholder="Optional description or notes"
          rows="3"
          @input="clearError('description')"
        />
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
        <label for="topic" class="form-label">Project/Topic *</label>
        <div class="topic-input-container">
          <AppInput
            id="topic"
            v-model="formData.topic"
            placeholder="Enter project name or select existing"
            :error="errors.topic"
            list="existing-topics"
            @input="clearError('topic')"
          />
          <datalist id="existing-topics">
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
        <label for="shareTo" class="form-label">Share Destination</label>
        <select
          id="shareTo"
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

      <!-- Content Preview (if available) -->
      <div v-if="formData.content" class="form-group">
        <label class="form-label">Content Preview</label>
        <div class="content-preview">
          {{ formData.content.substring(0, 300) }}
          <span v-if="formData.content.length > 300">...</span>
        </div>
      </div>
    </form>

    <template #footer>
      <AppButton variant="secondary" @click="handleClose">
        Cancel
      </AppButton>
      <AppButton
        variant="primary"
        :loading="isSubmitting"
        @click="handleSubmit"
      >
        Add Bookmark
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useNotifications } from '@/composables/useNotifications'

interface Props {
  show: boolean
  existingTopics: string[]
}

interface FormData {
  url: string
  title: string
  description: string
  action: string
  topic: string
  shareTo: string
  content: string
}

interface UrlPreview {
  title: string
  description: string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'submit': [data: FormData]
}>()

const { validationError } = useNotifications()

// Form state
const formData = ref<FormData>({
  url: '',
  title: '',
  description: '',
  action: 'read-later',
  topic: '',
  shareTo: '',
  content: ''
})

const errors = ref<Record<string, string>>({})
const isSubmitting = ref(false)
const isLoadingMetadata = ref(false)
const urlPreview = ref<UrlPreview | null>(null)

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

// Validation
const validateUrl = () => {
  if (!formData.value.url) {
    errors.value.url = 'URL is required'
    return false
  }
  
  try {
    new URL(formData.value.url)
    errors.value.url = ''
    loadUrlMetadata()
    return true
  } catch {
    errors.value.url = 'Please enter a valid URL'
    return false
  }
}

const validateForm = (): boolean => {
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

// URL metadata loading (mock implementation)
const loadUrlMetadata = async () => {
  if (!formData.value.url) return
  
  isLoadingMetadata.value = true
  
  try {
    // Mock API call - in real implementation, this would fetch page metadata
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // Mock response based on URL
    const url = new URL(formData.value.url)
    urlPreview.value = {
      title: `Page Title from ${url.hostname}`,
      description: 'This is a mock description that would be extracted from the page metadata.'
    }
    
    // Auto-fill title if empty
    if (!formData.value.title) {
      formData.value.title = urlPreview.value.title
    }
  } catch (error) {
    console.error('Failed to load URL metadata:', error)
  } finally {
    isLoadingMetadata.value = false
  }
}

// Event handlers
const clearError = (field: string) => {
  if (errors.value[field]) {
    delete errors.value[field]
  }
}

const handleActionChange = () => {
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
  formData.value.topic = topic
  clearError('topic')
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  isSubmitting.value = true
  
  try {
    // Add domain extraction
    const domain = new URL(formData.value.url).hostname
    
    const submitData = {
      ...formData.value,
      domain,
      timestamp: new Date().toISOString()
    }
    
    emit('submit', submitData)
    handleClose()
  } catch (error) {
    console.error('Failed to submit bookmark:', error)
    errors.value.submit = 'Failed to add bookmark. Please try again.'
    validationError('Please check your bookmark details and try again.')
  } finally {
    isSubmitting.value = false
  }
}

const handleClose = () => {
  // Reset form
  formData.value = {
    url: '',
    title: '',
    description: '',
    action: 'read-later',
    topic: '',
    shareTo: '',
    content: ''
  }
  errors.value = {}
  urlPreview.value = null
  isSubmitting.value = false
  isLoadingMetadata.value = false
  
  emit('update:show', false)
}

// Watch for URL changes to trigger metadata loading
watch(() => formData.value.url, (newUrl) => {
  if (newUrl && urlPreview.value) {
    // Clear previous preview when URL changes
    urlPreview.value = null
  }
})
</script>

<style scoped>
.add-bookmark-form {
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

/* URL Preview */
.url-status {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  padding: var(--spacing-sm);
  background: var(--color-gray-50);
  border-radius: var(--radius-sm);
}

.url-preview {
  padding: var(--spacing-md);
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.url-preview-title {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  margin-bottom: var(--spacing-xs);
}

.url-preview-description {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  line-height: var(--line-height-relaxed);
}

/* Action Grid */
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

/* Topic Selection */
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

/* Content Preview */
.content-preview {
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
  line-height: var(--line-height-relaxed);
  max-height: 150px;
  overflow-y: auto;
}

/* Responsive */
@media (max-width: 768px) {
  .action-grid {
    grid-template-columns: 1fr;
  }
  
  .topic-chips {
    justify-content: flex-start;
  }
}
</style>