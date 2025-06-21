<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
      <div
        class="modal-container"
        :class="[`modal-size-${size}`, { 'modal-full-screen': fullScreen }]"
        @click.stop
        role="dialog"
        :aria-labelledby="titleId"
        :aria-describedby="contentId"
        aria-modal="true"
      >
        <!-- Modal Header -->
        <div class="modal-header">
          <h2 :id="titleId" class="modal-title">
            <slot name="title">{{ title }}</slot>
          </h2>
          <button
            class="modal-close"
            @click="handleClose"
            aria-label="Close modal"
          >
            ✕
          </button>
        </div>

        <!-- Modal Content -->
        <div :id="contentId" class="modal-content">
          <slot />
        </div>

        <!-- Modal Footer -->
        <div v-if="$slots.footer" class="modal-footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch, nextTick } from 'vue'

interface Props {
  show: boolean
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  fullScreen?: boolean
  closeOnOverlay?: boolean
  closeOnEscape?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  size: 'md',
  fullScreen: false,
  closeOnOverlay: true,
  closeOnEscape: true
})

const emit = defineEmits<{
  'update:show': [value: boolean]
  close: []
}>()

const titleId = computed(() => `modal-title-${Math.random().toString(36).substr(2, 9)}`)
const contentId = computed(() => `modal-content-${Math.random().toString(36).substr(2, 9)}`)

const handleClose = () => {
  emit('update:show', false)
  emit('close')
}

const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    handleClose()
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && props.closeOnEscape && props.show) {
    handleClose()
  }
}

// Focus management
const focusModal = async () => {
  await nextTick()
  const modal = document.querySelector('.modal-container') as HTMLElement
  if (modal) {
    modal.focus()
  }
}

watch(
  () => props.show,
  (newShow) => {
    if (newShow) {
      document.addEventListener('keydown', handleKeydown)
      document.body.style.overflow = 'hidden'
      focusModal()
    } else {
      document.removeEventListener('keydown', handleKeydown)
      document.body.style.overflow = ''
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal, 1000);
  padding: var(--spacing-lg);
  backdrop-filter: blur(4px);
}

.modal-container {
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  width: 100%;
  outline: none;
  animation: modalEnter 0.2s ease-out;
}

@keyframes modalEnter {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Modal Sizes */
.modal-size-sm {
  max-width: 400px;
}

.modal-size-md {
  max-width: 600px;
}

.modal-size-lg {
  max-width: 800px;
}

.modal-size-xl {
  max-width: 1000px;
}

.modal-full-screen {
  max-width: none;
  max-height: none;
  width: 100vw;
  height: 100vh;
  border-radius: 0;
}

/* Modal Header */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg) var(--spacing-xl);
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.modal-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  margin: 0;
  color: var(--color-gray-800);
}

.modal-close {
  background: none;
  border: none;
  font-size: var(--font-size-xl);
  color: var(--color-gray-500);
  cursor: pointer;
  padding: var(--spacing-xs);
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.modal-close:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
}

.modal-close:focus {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

/* Modal Content */
.modal-content {
  padding: var(--spacing-xl);
  overflow-y: auto;
  flex: 1;
}

/* Modal Footer */
.modal-footer {
  padding: var(--spacing-lg) var(--spacing-xl);
  border-top: 1px solid var(--border-light);
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

/* Responsive */
@media (max-width: 768px) {
  .modal-overlay {
    padding: var(--spacing-md);
  }
  
  .modal-container {
    max-height: 95vh;
  }
  
  .modal-header,
  .modal-content,
  .modal-footer {
    padding-left: var(--spacing-lg);
    padding-right: var(--spacing-lg);
  }
  
  .modal-size-sm,
  .modal-size-md,
  .modal-size-lg,
  .modal-size-xl {
    max-width: none;
  }
}

@media (max-width: 480px) {
  .modal-overlay {
    padding: var(--spacing-sm);
  }
  
  .modal-container {
    border-radius: var(--radius-md);
  }
  
  .modal-header,
  .modal-content,
  .modal-footer {
    padding-left: var(--spacing-md);
    padding-right: var(--spacing-md);
  }
}
</style>