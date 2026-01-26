<template>
  <div class="ui-input-wrapper">
    <label v-if="label" class="ui-input__label">{{ label }}</label>
    <div class="ui-input__container" :class="{ 'ui-input--error': error, 'ui-input--disabled': disabled }">
      <span v-if="$slots.prefix" class="ui-input__prefix">
        <slot name="prefix"></slot>
      </span>
      <input
        :type="showPassword ? 'text' : type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        class="ui-input"
        @input="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
      />
      <span v-if="type === 'password' && showPasswordToggle" class="ui-input__suffix" @click="togglePassword">
        <svg v-if="showPassword" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
          <circle cx="12" cy="12" r="3"></circle>
        </svg>
        <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
          <line x1="1" y1="1" x2="23" y2="23"></line>
        </svg>
      </span>
      <span v-else-if="$slots.suffix" class="ui-input__suffix">
        <slot name="suffix"></slot>
      </span>
    </div>
    <div v-if="error" class="ui-input__error">{{ error }}</div>
    <div v-if="hint && !error" class="ui-input__hint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, defineEmits } from 'vue'

interface Props {
  modelValue: string | number
  type?: 'text' | 'password' | 'email' | 'number' | 'tel' | 'url'
  label?: string
  placeholder?: string
  error?: string
  hint?: string
  disabled?: boolean
  readonly?: boolean
  showPasswordToggle?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  showPasswordToggle: true,
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  focus: [event: FocusEvent]
  blur: [event: FocusEvent]
}>()

const showPassword = ref(false)
const isFocused = ref(false)

const togglePassword = () => {
  showPassword.value = !showPassword.value
}

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

const handleFocus = (event: FocusEvent) => {
  isFocused.value = true
  emit('focus', event)
}

const handleBlur = (event: FocusEvent) => {
  isFocused.value = false
  emit('blur', event)
}
</script>

<style scoped>
.ui-input-wrapper {
  width: 100%;
}

.ui-input__label {
  display: block;
  margin-bottom: var(--spacing-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-regular);
  font-weight: var(--font-weight-medium);
  font-family: var(--font-family);
}

.ui-input__container {
  position: relative;
  display: flex;
  align-items: center;
  background-color: var(--color-bg-secondary);
  backdrop-filter: var(--glass-blur-strong);
  -webkit-backdrop-filter: var(--glass-blur-strong);
  border: 0.5px solid var(--color-separator);
  border-radius: var(--border-radius-lg);
  transition: all var(--transition-base);
  min-height: var(--touch-target);
  transform: scale(1);
}

.ui-input__container:focus-within {
  background-color: var(--color-bg);
  border-color: var(--color-primary);
  box-shadow: 
    0 0 0 4px rgba(10, 132, 255, 0.12),
    0 4px 16px rgba(10, 132, 255, 0.15),
    0 2px 8px rgba(0, 0, 0, 0.1);
  transform: scale(1.01);
  border-width: 1px;
}

.ui-input--error .ui-input__container {
  border-color: var(--color-danger);
}

.ui-input--error .ui-input__container:focus-within {
  box-shadow: 
    0 0 0 4px rgba(255, 69, 58, 0.12),
    0 4px 16px rgba(255, 69, 58, 0.15),
    0 2px 8px rgba(0, 0, 0, 0.1);
  border-color: var(--color-danger);
}

@media (prefers-color-scheme: dark) {
  .ui-input__container:focus-within {
    box-shadow: 
      0 0 0 4px rgba(10, 132, 255, 0.2),
      0 4px 16px rgba(10, 132, 255, 0.25),
      0 2px 8px rgba(0, 0, 0, 0.3);
  }
  
  .ui-input--error .ui-input__container:focus-within {
    box-shadow: 
      0 0 0 4px rgba(255, 69, 58, 0.2),
      0 4px 16px rgba(255, 69, 58, 0.25),
      0 2px 8px rgba(0, 0, 0, 0.3);
  }
}

.ui-input--disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ui-input {
  flex: 1;
  width: 100%;
  padding: 0 var(--spacing-lg);
  font-size: var(--font-size-md);
  font-family: var(--font-family);
  color: var(--color-text-primary);
  background: transparent;
  border: none;
  outline: none;
  min-height: var(--touch-target);
  -webkit-appearance: none;
  appearance: none;
}

.ui-input::placeholder {
  color: var(--color-text-placeholder);
  font-size: var(--font-size-md);
}

.ui-input:disabled {
  cursor: not-allowed;
}

.ui-input__prefix,
.ui-input__suffix {
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-sm);
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.ui-input__suffix {
  cursor: pointer;
  user-select: none;
}

.ui-input__error {
  margin-top: var(--spacing-xs);
  font-size: var(--font-size-xs);
  color: var(--color-danger);
}

.ui-input__hint {
  margin-top: var(--spacing-xs);
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}
</style>
