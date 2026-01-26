<template>
  <button
    :class="[
      'ui-button',
      `ui-button--${type}`,
      `ui-button--${size}`,
      {
        'ui-button--loading': loading,
        'ui-button--disabled': disabled,
        'ui-button--block': block,
        'ui-button--round': round,
      }
    ]"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <span v-if="loading" class="ui-button__loading">
      <svg class="ui-button__spinner" viewBox="0 0 50 50">
        <circle class="path" cx="25" cy="25" r="20" fill="none" stroke-width="4"></circle>
      </svg>
    </span>
    <span v-if="$slots.icon" class="ui-button__icon">
      <slot name="icon"></slot>
    </span>
    <span v-if="$slots.default" class="ui-button__text">
      <slot></slot>
    </span>
  </button>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'

interface Props {
  type?: 'primary' | 'danger' | 'success' | 'warning' | 'info' | 'default'
  size?: 'small' | 'medium' | 'large'
  loading?: boolean
  disabled?: boolean
  block?: boolean
  round?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'default',
  size: 'medium',
  loading: false,
  disabled: false,
  block: false,
  round: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const handleClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit('click', event)
  }
}
</script>

<style scoped>
.ui-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: 0 var(--spacing-lg);
  min-height: var(--touch-target);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-family);
  border-radius: var(--border-radius-lg);
  border: none;
  transition: all var(--transition-base);
  cursor: pointer;
  user-select: none;
  white-space: nowrap;
  position: relative;
  overflow: hidden;
  -webkit-tap-highlight-color: transparent;
}

.ui-button--block {
  width: 100%;
}

.ui-button--round {
  border-radius: var(--border-radius-round);
}

/* 尺寸 */
.ui-button--small {
  padding: 0 var(--spacing-sm);
  min-height: 32px;
  font-size: var(--font-size-sm);
}

.ui-button--medium {
  padding: 0 var(--spacing-md);
  min-height: 40px;
  font-size: var(--font-size-md);
}

.ui-button--large {
  padding: 0 var(--spacing-lg);
  min-height: 48px;
  font-size: var(--font-size-lg);
}

/* iOS 按钮类型 */
.ui-button--default {
  background-color: var(--color-bg-secondary);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  color: var(--color-primary);
  border: 0.5px solid var(--color-separator);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.ui-button--default:hover:not(.ui-button--disabled) {
  background-color: var(--color-bg-tertiary);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}

.ui-button--default:active:not(.ui-button--disabled) {
  background-color: var(--color-bg-tertiary);
  opacity: 0.8;
  transform: scale(0.97);
}

.ui-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-hover) 100%);
  color: white;
  border: none;
  box-shadow: 0 2px 8px rgba(10, 132, 255, 0.3), 0 1px 3px rgba(0, 0, 0, 0.2);
}

.ui-button--primary:hover:not(.ui-button--disabled) {
  background: linear-gradient(135deg, var(--color-primary-hover) 0%, var(--color-primary) 100%);
  box-shadow: 0 4px 12px rgba(10, 132, 255, 0.4), 0 2px 6px rgba(0, 0, 0, 0.2);
}

.ui-button--primary:active:not(.ui-button--disabled) {
  background: var(--color-primary-active);
  box-shadow: 0 1px 4px rgba(10, 132, 255, 0.3);
  transform: scale(0.97);
}

.ui-button--danger {
  background-color: var(--color-danger);
  color: white;
  border: none;
}

.ui-button--danger:hover:not(.ui-button--disabled) {
  background-color: var(--color-danger-hover);
}

.ui-button--danger:active:not(.ui-button--disabled) {
  background-color: var(--color-danger-active);
  opacity: 0.9;
}

.ui-button--success {
  background-color: var(--color-success);
  color: white;
  border: 1px solid var(--color-success);
}

.ui-button--success:hover:not(.ui-button--disabled) {
  background-color: var(--color-success-hover);
  border-color: var(--color-success-hover);
}

.ui-button--warning {
  background-color: var(--color-warning);
  color: white;
  border: 1px solid var(--color-warning);
}

.ui-button--warning:hover:not(.ui-button--disabled) {
  background-color: var(--color-warning-hover);
  border-color: var(--color-warning-hover);
}

.ui-button--info {
  background-color: var(--color-info);
  color: white;
  border: 1px solid var(--color-info);
}

.ui-button--info:hover:not(.ui-button--disabled) {
  background-color: var(--color-info-hover);
  border-color: var(--color-info-hover);
}

/* 状态 */
.ui-button--disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ui-button--loading {
  cursor: not-allowed;
}

.ui-button__loading {
  display: inline-flex;
  align-items: center;
}

.ui-button__spinner {
  width: 16px;
  height: 16px;
  animation: rotate 1s linear infinite;
}

.ui-button__spinner .path {
  stroke: currentColor;
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

.ui-button__icon {
  display: inline-flex;
  align-items: center;
  font-size: 1.2em;
}

.ui-button__text {
  flex: 1;
}

@keyframes rotate {
  100% {
    transform: rotate(360deg);
  }
}

@keyframes dash {
  0% {
    stroke-dasharray: 1, 150;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -35;
  }
  100% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -124;
  }
}

/* iOS 触摸反馈 */
.ui-button:active:not(.ui-button--disabled) {
  transform: scale(0.97);
  transition: transform 0.1s ease;
}

@media (hover: none) {
  .ui-button:active:not(.ui-button--disabled) {
    transform: scale(0.95);
    opacity: 0.8;
  }
}

/* Ripple 效果 */
.ui-button {
  position: relative;
  overflow: hidden;
}
</style>
