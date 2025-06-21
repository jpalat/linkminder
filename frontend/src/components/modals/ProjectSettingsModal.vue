<template>
  <AppModal
    :show="show"
    @close="$emit('close')"
    title="Project Settings"
    size="md"
  >
    <div class="project-settings">
      <form @submit.prevent="handleSave" class="settings-form">
        <!-- Project Name -->
        <div class="form-group">
          <label for="project-name" class="form-label">Project Name</label>
          <AppInput
            id="project-name"
            v-model="formData.name"
            type="text"
            placeholder="Enter project name..."
            required
          />
        </div>

        <!-- Project Description -->
        <div class="form-group">
          <label for="project-description" class="form-label">Description</label>
          <textarea
            id="project-description"
            v-model="formData.description"
            class="form-textarea"
            placeholder="Enter project description..."
            rows="3"
          ></textarea>
        </div>

        <!-- Project Status -->
        <div class="form-group">
          <label for="project-status" class="form-label">Status</label>
          <select id="project-status" v-model="formData.status" class="form-select">
            <option value="active">Active</option>
            <option value="stale">Stale</option>
            <option value="inactive">Inactive</option>
          </select>
        </div>

        <!-- Project Actions -->
        <div class="form-group">
          <label class="form-label">Project Actions</label>
          <div class="action-buttons">
            <AppButton
              type="button"
              variant="secondary"
              @click="exportProject"
              class="action-btn"
            >
              📤 Export Project
            </AppButton>
            <AppButton
              type="button"
              variant="secondary"
              @click="archiveProject"
              class="action-btn"
            >
              📁 Archive Project
            </AppButton>
            <AppButton
              type="button"
              variant="danger"
              @click="deleteProject"
              class="action-btn"
            >
              🗑️ Delete Project
            </AppButton>
          </div>
        </div>

        <!-- Form Actions -->
        <div class="form-actions">
          <AppButton
            type="button"
            variant="secondary"
            @click="$emit('close')"
          >
            Cancel
          </AppButton>
          <AppButton
            type="submit"
            variant="primary"
            :loading="saving"
          >
            Save Changes
          </AppButton>
        </div>
      </form>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import type { ProjectDetail } from '@/types'

interface Props {
  show: boolean
  project: ProjectDetail
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
  save: [project: Partial<ProjectDetail>]
  export: [project: ProjectDetail]
  archive: [project: ProjectDetail]
  delete: [project: ProjectDetail]
}>()

// Form state
const formData = ref({
  name: '',
  description: '',
  status: 'active' as 'active' | 'stale' | 'inactive'
})

const saving = ref(false)

// Initialize form data
onMounted(() => {
  formData.value = {
    name: props.project.topic,
    description: props.project.topic, // Default to topic name if no description
    status: props.project.status
  }
})

const handleSave = async () => {
  saving.value = true
  try {
    emit('save', {
      topic: formData.value.name,
      status: formData.value.status
    })
  } finally {
    saving.value = false
  }
}

const exportProject = () => {
  emit('export', props.project)
}

const archiveProject = () => {
  emit('archive', props.project)
}

const deleteProject = () => {
  emit('delete', props.project)
}
</script>

<style scoped>
.project-settings {
  padding: var(--spacing-md);
}

.settings-form {
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
  color: var(--color-gray-800);
  font-size: var(--font-size-sm);
}

.form-textarea {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--border-radius);
  font-size: var(--font-size-base);
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-light);
}

.form-select {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--border-light);
  border-radius: var(--border-radius);
  background: white;
  font-size: var(--font-size-base);
}

.form-select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-light);
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.action-btn {
  justify-content: flex-start;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--border-light);
}

@media (min-width: 480px) {
  .action-buttons {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--spacing-md);
  }
  
  .action-btn:last-child {
    grid-column: span 2;
  }
}
</style>