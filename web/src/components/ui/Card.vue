<template>
  <div :class="['ui-card', { 'ui-card--shadow': shadow, 'ui-card--hover': hover }]">
    <div v-if="$slots.header || header" class="ui-card__header">
      <slot name="header">{{ header }}</slot>
    </div>
    <div class="ui-card__body">
      <slot></slot>
    </div>
    <div v-if="$slots.footer" class="ui-card__footer">
      <slot name="footer"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'

interface Props {
  header?: string
  shadow?: boolean
  hover?: boolean
}

withDefaults(defineProps<Props>(), {
  shadow: true,
  hover: false,
})
</script>

<style scoped>
.ui-card {
  background-color: var(--color-bg);
  backdrop-filter: var(--glass-blur-strong);
  -webkit-backdrop-filter: var(--glass-blur-strong);
  border-radius: 24px;
  overflow: hidden;
  transition: all var(--transition-base);
  border: 0.5px solid rgba(0, 0, 0, 0.05);
  position: relative;
}

/* iOS 26 边框光效 */
.ui-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 24px;
  padding: 0.5px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.05) 50%, transparent 100%);
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
  opacity: 0;
  transition: opacity var(--transition-base);
}

.ui-card--shadow::before {
  opacity: 1;
}

.ui-card--shadow {
  box-shadow: var(--shadow-glass);
}

.ui-card--hover:hover {
  box-shadow: var(--shadow-glass-strong);
  transform: translateY(-2px);
  background-color: rgba(255, 255, 255, var(--glass-opacity-strong));
}

.ui-card--hover:hover::before {
  opacity: 1;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.15) 0%, rgba(255, 255, 255, 0.08) 50%, transparent 100%);
}

/* Apple Store 风格 - 统一使用浅色主题，移除深色模式 */

.ui-card__header {
  padding: var(--spacing-md);
  border-bottom: 0.5px solid var(--color-separator);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-lg);
  color: var(--color-text-primary);
  font-family: var(--font-family);
}

.ui-card__body {
  padding: var(--spacing-md);
}

.ui-card__footer {
  padding: var(--spacing-md);
  border-top: 0.5px solid var(--color-separator);
  background-color: var(--color-bg-secondary);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
}

@media (max-width: 767px) {
  .ui-card__header,
  .ui-card__body,
  .ui-card__footer {
    padding: var(--spacing-sm);
  }
}
</style>
