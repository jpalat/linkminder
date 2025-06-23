<template>
  <Teleport to="body">
    <div
      v-if="notifications.length > 0"
      class="toast-container"
      aria-live="polite"
      aria-label="Notifications"
    >
      <TransitionGroup
        name="toast-list"
        tag="div"
        class="toast-list"
      >
        <AppToast
          v-for="notification in notifications"
          :key="notification.id"
          v-bind="notification"
          @close="(id) => id && removeNotification(id)"
        />
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useNotifications } from '@/composables/useNotifications'
import AppToast from './AppToast.vue'

const { notifications, removeNotification } = useNotifications()
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: var(--spacing-xl);
  right: var(--spacing-xl);
  z-index: 9999;
  pointer-events: none;
}

.toast-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.toast-list > * {
  pointer-events: auto;
}

/* Toast List Transitions */
.toast-list-enter-active,
.toast-list-leave-active {
  transition: all 0.3s ease;
}

.toast-list-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.toast-list-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.toast-list-move {
  transition: transform 0.3s ease;
}

/* Responsive positioning */
@media (max-width: 768px) {
  .toast-container {
    top: var(--spacing-md);
    right: var(--spacing-md);
    left: var(--spacing-md);
  }
  
  .toast-list {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .toast-container {
    top: var(--spacing-sm);
    right: var(--spacing-sm);
    left: var(--spacing-sm);
  }
}
</style>