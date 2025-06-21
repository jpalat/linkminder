<template>
  <span
    :class="[
      'badge',
      variant,
      size
    ]"
  >
    <slot />
  </span>
</template>

<script setup lang="ts">
interface Props {
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'
  size?: 'sm' | 'md' | 'lg'
}

withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'md'
})
</script>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-xl);
  white-space: nowrap;
}

/* Variants */
.default {
  background: var(--color-gray-200);
  color: var(--color-gray-700);
}

.primary {
  background: var(--color-primary);
  color: white;
}

.success {
  background: #c6f6d5;
  color: #22543d;
}

.warning {
  background: #fed7aa;
  color: #9c4221;
}

.danger {
  background: #fed7d7;
  color: #742a2a;
}

.info {
  background: #e9d8fd;
  color: #553c9a;
}

/* Sizes */
.sm {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--font-size-xs);
}

.md {
  padding: var(--spacing-xs) var(--spacing-md);
  font-size: var(--font-size-sm);
}

.lg {
  padding: var(--spacing-sm) var(--spacing-lg);
  font-size: var(--font-size-base);
}
</style>
