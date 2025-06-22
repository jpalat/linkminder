<template>
  <AppModal
    :show="show"
    title="Move to Project"
    size="md"
    @update:show="$emit('update:show', $event)"
    @close="handleClose"
  >
    <div class="move-to-project-content">
      <div v-if="bookmark" class="bookmark-info">
        <h3 class="bookmark-title">{{ bookmark.title }}</h3>
        <div class="bookmark-url">{{ bookmark.url }}</div>
      </div>
      
      <div class="form-group">
        <label for="project-topic" class="form-label">Select Project *</label>
        <div class="topic-input-container">
          <AppInput
            id="project-topic"
            v-model="selectedTopic"
            placeholder="Enter project name or select existing"
            :error="errors.topic"
            list="existing-projects"
            @input="clearError"
            @keydown.enter="handleSubmit"
          />
          <datalist id="existing-projects">
            <option v-for="topic in existingTopics" :key="topic" :value="topic" />
          </datalist>
        </div>
        <div v-if="existingTopics.length > 0" class="existing-topics">
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
    </div>

    <template #footer>
      <AppButton variant="secondary" @click="handleClose">
        Cancel
      </AppButton>
      <AppButton
        variant="primary"
        :loading="isSubmitting"
        @click="handleSubmit"
      >
        Move to Project
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Bookmark } from '@/types'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

interface Props {
  show: boolean
  bookmark?: Bookmark | null
  existingTopics: string[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'submit': [bookmarkId: string, topic: string]
}>()

// Form state
const selectedTopic = ref('')
const errors = ref<{ topic?: string }>({})
const isSubmitting = ref(false)

// Reset form when modal opens/closes or bookmark changes
watch(() => [props.show, props.bookmark], () => {
  if (props.show && props.bookmark) {
    selectedTopic.value = props.bookmark.topic || ''
    errors.value = {}
  } else {
    selectedTopic.value = ''
    errors.value = {}
  }
}, { immediate: true })

const selectTopic = (topic: string) => {
  selectedTopic.value = topic
  clearError()
}

const clearError = () => {
  if (errors.value.topic) {
    delete errors.value.topic
  }
}

const validateForm = (): boolean => {
  const newErrors: { topic?: string } = {}
  
  if (!selectedTopic.value.trim()) {
    newErrors.topic = 'Project/Topic is required'
  }
  
  errors.value = newErrors
  return Object.keys(newErrors).length === 0
}

const handleSubmit = async () => {
  if (!validateForm() || !props.bookmark) return
  
  isSubmitting.value = true
  
  try {
    emit('submit', props.bookmark.id, selectedTopic.value.trim())
    handleClose()
  } catch (error) {
    console.error('Failed to move bookmark to project:', error)
  } finally {
    isSubmitting.value = false
  }
}

const handleClose = () => {
  selectedTopic.value = ''
  errors.value = {}
  isSubmitting.value = false
  emit('update:show', false)
}
</script>

<style scoped>
.move-to-project-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.bookmark-info {
  padding: var(--spacing-md);
  background: var(--color-gray-50);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.bookmark-title {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
  margin: 0 0 var(--spacing-xs) 0;
  font-size: var(--font-size-base);
}

.bookmark-url {
  color: var(--color-primary);
  font-size: var(--font-size-sm);
  word-break: break-all;
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

/* Responsive */
@media (max-width: 768px) {
  .topic-chips {
    justify-content: flex-start;
  }
}
</style>