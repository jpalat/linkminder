<template>
  <div class="input-container">
    <div v-if="icon" class="input-icon">
      {{ icon }}
    </div>
    <input
      :class="[
        'input',
        {
          'input-with-icon': icon,
          'input-error': error
        }
      ]"
      :type="type"
      :placeholder="placeholder"
      :value="modelValue"
      :disabled="disabled"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @focus="$emit('focus', $event)"
      @blur="$emit('blur', $event)"
    />
  </div>
  <div v-if="error" class="input-error-text">
    {{ error }}
  </div>
</template>

<script setup lang="ts">
interface Props {
  modelValue?: string
  type?: string
  placeholder?: string
  icon?: string
  error?: string
  disabled?: boolean
}

withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false
})

defineEmits<{
  'update:modelValue': [value: string]
  focus: [event: FocusEvent]
  blur: [event: FocusEvent]
}>()
</script>

<style scoped>
.input-container {
  position: relative;
  width: 100%;
}

.input {
  width: 100%;
  padding: var(--spacing-md) var(--spacing-lg);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  font-size: var(--font-size-base);
  background: var(--bg-input);
  color: var(--color-gray-800);
  transition: var(--transition-fast);
}

.input:focus {
  outline: none;
  border-color: var(--border-focus);
  background: var(--bg-input-focus);
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.input-with-icon {
  padding-left: 2.5rem;
}

.input-icon {
  position: absolute;
  left: var(--spacing-md);
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-gray-600);
  font-size: var(--font-size-base);
  pointer-events: none;
  z-index: 1;
}

.input-error {
  border-color: var(--color-danger);
}

.input-error:focus {
  box-shadow: 0 0 0 3px rgba(229, 62, 62, 0.1);
}

.input-error-text {
  margin-top: var(--spacing-xs);
  font-size: var(--font-size-sm);
  color: var(--color-danger);
}

.input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: var(--color-gray-100);
}

.input::placeholder {
  color: var(--color-gray-500);
}
</style>
