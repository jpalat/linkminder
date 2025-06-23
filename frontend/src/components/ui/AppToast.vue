<template>
  <Transition
    name="toast"
    enter-active-class="toast-enter-active"
    leave-active-class="toast-leave-active"
    enter-from-class="toast-enter-from"
    leave-to-class="toast-leave-to"
  >
    <div
      v-if="visible"
      class="toast"
      :class="[
        `toast--${type}`,
        { 'toast--dismissible': dismissible }
      ]"
      role="alert"
      :aria-live="type === 'error' ? 'assertive' : 'polite'"
    >
      <div class="toast__icon">
        <span v-if="type === 'success'">✅</span>
        <span v-else-if="type === 'error'">❌</span>
        <span v-else-if="type === 'warning'">⚠️</span>
        <span v-else-if="type === 'info'">ℹ️</span>
        <span v-else>📢</span>
      </div>
      
      <div class="toast__content">
        <div v-if="title" class="toast__title">{{ title }}</div>
        <div class="toast__message">{{ message }}</div>
      </div>
      
      <button
        v-if="dismissible"
        class="toast__close"
        @click="handleClose"
        aria-label="Close notification"
      >
        ×
      </button>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

export interface ToastProps {
  id?: string
  type?: 'success' | 'error' | 'warning' | 'info'
  title?: string
  message: string
  duration?: number
  dismissible?: boolean
  persistent?: boolean
}

const props = withDefaults(defineProps<ToastProps>(), {
  type: 'info',
  duration: 5000,
  dismissible: true,
  persistent: false
})

const emit = defineEmits<{
  close: [id?: string]
}>()

const visible = ref(true)
let timeoutId: ReturnType<typeof setTimeout> | null = null

const handleClose = () => {
  visible.value = false
  if (timeoutId) {
    clearTimeout(timeoutId)
    timeoutId = null
  }
  // Emit after transition completes
  setTimeout(() => {
    emit('close', props.id)
  }, 300)
}

onMounted(() => {
  if (!props.persistent && props.duration > 0) {
    timeoutId = setTimeout(() => {
      handleClose()
    }, props.duration)
  }
})
</script>

<style scoped>
.toast {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-sm);
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-left: 4px solid var(--border-light);
  max-width: 400px;
  min-width: 300px;
  position: relative;
}

.toast--success {
  border-left-color: var(--color-green-500);
  background: var(--color-green-50);
}

.toast--error {
  border-left-color: var(--color-red-500);
  background: var(--color-red-50);
}

.toast--warning {
  border-left-color: var(--color-yellow-500);
  background: var(--color-yellow-50);
}

.toast--info {
  border-left-color: var(--color-blue-500);
  background: var(--color-blue-50);
}

.toast__icon {
  flex-shrink: 0;
  font-size: var(--font-size-lg);
  line-height: 1;
}

.toast__content {
  flex: 1;
  min-width: 0;
}

.toast__title {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  margin-bottom: var(--spacing-xs);
  font-size: var(--font-size-sm);
}

.toast__message {
  color: var(--color-gray-700);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
  word-wrap: break-word;
}

.toast__close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: var(--color-gray-500);
  font-size: var(--font-size-xl);
  line-height: 1;
  cursor: pointer;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
}

.toast__close:hover {
  background: var(--color-gray-100);
  color: var(--color-gray-700);
}

/* Toast Transitions */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.toast-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

/* Responsive */
@media (max-width: 768px) {
  .toast {
    max-width: calc(100vw - 2rem);
    min-width: 280px;
  }
}
</style>