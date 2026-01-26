<template>
  <Transition name="toast">
    <div v-if="visible" :class="['ui-toast', `ui-toast--${type}`]" @click="handleClick">
      <span class="ui-toast__icon">
        <slot name="icon">
          <svg v-if="type === 'success'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
          <svg v-else-if="type === 'error'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <svg v-else-if="type === 'warning'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
            <line x1="12" y1="9" x2="12" y2="13"></line>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
          <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="16" x2="12" y2="12"></line>
            <line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
        </slot>
      </span>
      <span class="ui-toast__message">{{ message }}</span>
      <button v-if="closable" class="ui-toast__close" @click.stop="handleClose">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Props {
  message: string
  type?: 'success' | 'error' | 'warning' | 'info'
  duration?: number
  closable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  duration: 3000,
  closable: true,
})

const emit = defineEmits<{
  close: []
  click: []
}>()

const visible = ref(false)
let timer: number | null = null

const handleClose = () => {
  visible.value = false
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
  setTimeout(() => emit('close'), 300)
}

const handleClick = () => {
  emit('click')
}

onMounted(() => {
  visible.value = true
  if (props.duration > 0) {
    timer = window.setTimeout(() => {
      handleClose()
    }, props.duration)
  }
})
</script>

<style scoped>
/* 清新风格 Toast - 与登录页面保持一致 */
.ui-toast {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-radius: 14px;
  box-shadow: 
    0 4px 16px rgba(0, 0, 0, 0.06),
    0 2px 8px rgba(0, 0, 0, 0.04),
    0 1px 2px rgba(0, 0, 0, 0.02),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
  border: 0.5px solid rgba(0, 0, 0, 0.08);
  min-width: 280px;
  max-width: calc(100vw - 32px);
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0.0, 0.2, 1);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
}

.ui-toast:hover {
  transform: translateY(-1px);
  box-shadow: 
    0 6px 20px rgba(0, 0, 0, 0.08),
    0 3px 10px rgba(0, 0, 0, 0.05),
    0 1px 3px rgba(0, 0, 0, 0.03),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

/* Material Design + 中式美学 - 错误类型（朱砂红） */
.ui-toast--error {
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-color: rgba(211, 47, 47, 0.18);
  box-shadow: 
    0 4px 16px rgba(211, 47, 47, 0.1),
    0 2px 8px rgba(211, 47, 47, 0.06),
    0 1px 2px rgba(0, 0, 0, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.ui-toast--error:hover {
  box-shadow: 
    0 6px 20px rgba(211, 47, 47, 0.12),
    0 3px 10px rgba(211, 47, 47, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.ui-toast--error .ui-toast__icon {
  color: #D32F2F;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(211, 47, 47, 0.1);
  border-radius: 50%;
  padding: 4px;
}

.ui-toast--error .ui-toast__message {
  color: #B71C1C;
  font-weight: 500;
  font-size: 14px;
  line-height: 1.5;
  letter-spacing: -0.01em;
}

.ui-toast--error .ui-toast__close {
  color: rgba(183, 28, 28, 0.5);
}

.ui-toast--error .ui-toast__close:hover {
  background-color: rgba(211, 47, 47, 0.1);
  color: #B71C1C;
}

/* Material Design + 中式美学 - 成功类型（竹青） */
.ui-toast--success {
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-color: rgba(46, 125, 50, 0.18);
  box-shadow: 
    0 4px 16px rgba(46, 125, 50, 0.1),
    0 2px 8px rgba(46, 125, 50, 0.06),
    0 1px 2px rgba(0, 0, 0, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.ui-toast--success:hover {
  box-shadow: 
    0 6px 20px rgba(46, 125, 50, 0.12),
    0 3px 10px rgba(46, 125, 50, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.ui-toast--success .ui-toast__icon {
  color: #2E7D32;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(46, 125, 50, 0.1);
  border-radius: 50%;
  padding: 4px;
}

.ui-toast--success .ui-toast__message {
  color: #1B5E20;
  font-weight: 500;
  font-size: 14px;
  line-height: 1.5;
  letter-spacing: -0.01em;
}

.ui-toast--success .ui-toast__close {
  color: rgba(27, 94, 32, 0.5);
}

.ui-toast--success .ui-toast__close:hover {
  background-color: rgba(46, 125, 50, 0.1);
  color: #1B5E20;
}

/* Material Design + 中式美学 - 警告类型（琥珀） */
.ui-toast--warning {
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-color: rgba(237, 108, 2, 0.18);
  box-shadow: 
    0 4px 16px rgba(237, 108, 2, 0.1),
    0 2px 8px rgba(237, 108, 2, 0.06),
    0 1px 2px rgba(0, 0, 0, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.ui-toast--warning:hover {
  box-shadow: 
    0 6px 20px rgba(237, 108, 2, 0.12),
    0 3px 10px rgba(237, 108, 2, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.ui-toast--warning .ui-toast__icon {
  color: #ED6C02;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(237, 108, 2, 0.1);
  border-radius: 50%;
  padding: 4px;
}

.ui-toast--warning .ui-toast__message {
  color: #E65100;
  font-weight: 500;
  font-size: 14px;
  line-height: 1.5;
  letter-spacing: -0.01em;
}

.ui-toast--warning .ui-toast__close {
  color: rgba(230, 81, 0, 0.5);
}

.ui-toast--warning .ui-toast__close:hover {
  background-color: rgba(237, 108, 2, 0.1);
  color: #E65100;
}

/* Material Design + 中式美学 - 信息类型（天青） */
.ui-toast--info {
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-color: rgba(2, 136, 209, 0.18);
  box-shadow: 
    0 4px 16px rgba(2, 136, 209, 0.1),
    0 2px 8px rgba(2, 136, 209, 0.06),
    0 1px 2px rgba(0, 0, 0, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.ui-toast--info:hover {
  box-shadow: 
    0 6px 20px rgba(2, 136, 209, 0.12),
    0 3px 10px rgba(2, 136, 209, 0.08),
    0 1px 3px rgba(0, 0, 0, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.ui-toast--info .ui-toast__icon {
  color: #0288D1;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(2, 136, 209, 0.1);
  border-radius: 50%;
  padding: 4px;
}

.ui-toast--info .ui-toast__message {
  color: #01579B;
  font-weight: 500;
  font-size: 14px;
  line-height: 1.5;
  letter-spacing: -0.01em;
}

.ui-toast--info .ui-toast__close {
  color: rgba(1, 87, 155, 0.5);
}

.ui-toast--info .ui-toast__close:hover {
  background-color: rgba(2, 136, 209, 0.1);
  color: #01579B;
}

.ui-toast__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ui-toast__message {
  flex: 1;
  font-size: 14px;
  line-height: 1.4;
}

.ui-toast__close {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: rgba(0, 0, 0, 0.45);
  transition: all 0.2s ease;
  cursor: pointer;
  opacity: 0.6;
  background: transparent;
  border: none;
  padding: 0;
}

.ui-toast__close:hover {
  background-color: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.7);
  opacity: 1;
}

/* 清新风格动画 - 从左侧滑入 */
.toast-enter-active {
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.toast-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0.0, 0.2, 1);
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(-20px) scale(0.95);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(-10px) scale(0.98);
}

@media (max-width: 767px) {
  .ui-toast {
    min-width: auto;
    width: calc(100vw - 32px);
    padding: 12px 16px;
    gap: 10px;
    border-radius: 12px;
  }
  
  .ui-toast__icon {
    width: 18px;
    height: 18px;
  }
  
  .ui-toast__icon[class*="error"],
  .ui-toast__icon[class*="success"],
  .ui-toast__icon[class*="warning"],
  .ui-toast__icon[class*="info"] {
    padding: 3px;
  }
  
  .ui-toast__message {
    font-size: 13px;
  }
  
  .ui-toast__close {
    width: 18px;
    height: 18px;
  }
}
</style>
