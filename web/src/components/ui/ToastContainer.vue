<template>
  <div class="ui-toast-container" ref="containerRef"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, render } from 'vue'
import Toast from './Toast.vue'

export interface ToastOptions {
  message: string
  type?: 'success' | 'error' | 'warning' | 'info'
  duration?: number
  closable?: boolean
}

const containerRef = ref<HTMLElement>()
const toasts = ref<Array<{ id: number; options: ToastOptions }>>([])
let toastId = 0

const show = (options: ToastOptions) => {
  const id = toastId++
  toasts.value.push({ id, options })
  
  const div = document.createElement('div')
  const vnode = h(Toast, {
    ...options,
    onClose: () => {
      remove(id)
      if (containerRef.value) {
        containerRef.value.removeChild(div)
      }
    },
  })
  
  render(vnode, div)
  if (containerRef.value) {
    containerRef.value.appendChild(div)
  }
}

const remove = (id: number) => {
  const index = toasts.value.findIndex(t => t.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

const success = (message: string, duration = 3000) => {
  show({ message, type: 'success', duration })
}

const error = (message: string, duration = 3000) => {
  show({ message, type: 'error', duration })
}

const warning = (message: string, duration = 3000) => {
  show({ message, type: 'warning', duration })
}

const info = (message: string, duration = 3000) => {
  show({ message, type: 'info', duration })
}

onMounted(() => {
  // 导出到全局
  ;(window as any).$toast = { show, success, error, warning, info }
})

defineExpose({ show, success, error, warning, info })
</script>

<style scoped>
/* 清新风格 Toast 容器 - 左上角定位 */
.ui-toast-container {
  position: fixed;
  top: 15px;
  left: 16px;
  z-index: var(--z-index-toast);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--spacing-xs);
  pointer-events: none;
}

.ui-toast-container > * {
  pointer-events: auto;
}

@media (max-width: 767px) {
  .ui-toast-container {
    top: 15px;
    left: 16px;
    right: 16px;
    align-items: stretch;
    gap: var(--spacing-xs);
  }
}
</style>
