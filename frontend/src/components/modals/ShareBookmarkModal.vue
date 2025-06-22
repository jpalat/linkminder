<template>
  <AppModal
    :show="show"
    title="Share Bookmark"
    size="md"
    @update:show="$emit('update:show', $event)"
    @close="handleClose"
  >
    <div class="share-bookmark-content">
      <!-- Bookmark Info -->
      <div v-if="bookmarks && bookmarks.length === 1" class="bookmark-info">
        <h3 class="bookmark-title">{{ bookmarks[0].title }}</h3>
        <div class="bookmark-url">{{ bookmarks[0].url }}</div>
      </div>
      
      <!-- Multiple Bookmarks Info -->
      <div v-else-if="bookmarks && bookmarks.length > 1" class="bookmarks-info">
        <h3 class="bookmark-title">Share {{ bookmarks.length }} bookmarks</h3>
        <div class="bookmark-list">
          <div v-for="bookmark in bookmarks.slice(0, 3)" :key="bookmark.id" class="bookmark-item">
            {{ bookmark.title }}
          </div>
          <div v-if="bookmarks.length > 3" class="more-items">
            and {{ bookmarks.length - 3 }} more...
          </div>
        </div>
      </div>
      
      <!-- Share Destination Selection -->
      <div class="form-group">
        <label for="share-destination" class="form-label">Share Destination *</label>
        <AppInput
          id="share-destination"
          v-model="selectedDestination"
          placeholder="Enter recipient or select from previous destinations"
          :error="errors.destination"
          list="share-destinations"
          @input="clearError"
        />
        <datalist id="share-destinations">
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
              @click="selectDestination(destination)"
            >
              {{ destination }}
            </button>
          </div>
        </div>
      </div>

      <!-- Optional Notes -->
      <div class="form-group">
        <label for="share-notes" class="form-label">Notes (Optional)</label>
        <textarea
          id="share-notes"
          v-model="shareNotes"
          class="form-textarea"
          placeholder="Add any notes about why you're sharing this..."
          rows="3"
        />
      </div>
    </div>

    <template #footer>
      <AppButton variant="secondary" @click="handleClose">
        Cancel
      </AppButton>
      <AppButton 
        variant="primary" 
        @click="handleSubmit"
        :loading="isSubmitting"
        :disabled="!selectedDestination"
      >
        {{ bookmarks && bookmarks.length > 1 ? `Share ${bookmarks.length} Bookmarks` : 'Share Bookmark' }}
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useBookmarkStore } from '@/stores/bookmarks'
import type { Bookmark } from '@/types'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

interface Props {
  show: boolean
  bookmarks?: Bookmark | Bookmark[]
}

interface ShareData {
  destination: string
  notes?: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'submit': [data: ShareData]
}>()

// Store access
const bookmarkStore = useBookmarkStore()
const { availableShareDestinations } = storeToRefs(bookmarkStore)

// Reactive state
const selectedDestination = ref('')
const shareNotes = ref('')
const errors = ref<Record<string, string>>({})
const isSubmitting = ref(false)

// Computed properties
const bookmarks = computed(() => {
  if (!props.bookmarks) return null
  return Array.isArray(props.bookmarks) ? props.bookmarks : [props.bookmarks]
})

// Watch for modal show/hide to reset form
watch(() => props.show, (newShow) => {
  if (newShow) {
    // Reset form when modal opens
    selectedDestination.value = ''
    shareNotes.value = ''
    errors.value = {}
    isSubmitting.value = false
  }
})

// Methods
const clearError = () => {
  if (errors.value.destination) {
    delete errors.value.destination
  }
}

const selectDestination = (destination: string) => {
  selectedDestination.value = destination
  clearError()
}

const validateForm = (): boolean => {
  const newErrors: Record<string, string> = {}
  
  if (!selectedDestination.value) {
    newErrors.destination = 'Please select a share destination'
  }
  
  errors.value = newErrors
  return Object.keys(newErrors).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  isSubmitting.value = true
  
  try {
    emit('submit', {
      destination: selectedDestination.value,
      notes: shareNotes.value || undefined
    })
    handleClose()
  } catch (error) {
    console.error('Failed to share bookmark:', error)
    errors.value.submit = 'Failed to share bookmark. Please try again.'
  } finally {
    isSubmitting.value = false
  }
}

const handleClose = () => {
  emit('update:show', false)
}
</script>

<style scoped>
.share-bookmark-content {
  padding: var(--spacing-lg) 0;
}

.bookmark-info {
  background: var(--color-gray-50);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.bookmarks-info {
  background: var(--color-gray-50);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.bookmark-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  margin-bottom: var(--spacing-sm);
  line-height: var(--line-height-tight);
}

.bookmark-url {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  word-break: break-all;
}

.bookmark-list {
  margin-top: var(--spacing-sm);
}

.bookmark-item {
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
  margin-bottom: var(--spacing-xs);
  padding-left: var(--spacing-sm);
  border-left: 2px solid var(--color-gray-300);
}

.more-items {
  font-size: var(--font-size-sm);
  color: var(--color-gray-500);
  font-style: italic;
  margin-top: var(--spacing-xs);
}

.form-group {
  margin-bottom: var(--spacing-lg);
}

.form-label {
  display: block;
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-700);
  margin-bottom: var(--spacing-sm);
  font-size: var(--font-size-sm);
}

.form-select {
  width: 100%;
  padding: var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  background: white;
  color: var(--color-gray-800);
  transition: var(--transition-fast);
}

.form-select:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.form-select.error {
  border-color: var(--color-red-500);
}

.form-textarea {
  width: 100%;
  padding: var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
  transition: var(--transition-fast);
}

.form-textarea:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.form-error {
  color: var(--color-red-600);
  font-size: var(--font-size-sm);
  margin-top: var(--spacing-xs);
}

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

.destination-chip:active {
  transform: translateY(1px);
}
</style>