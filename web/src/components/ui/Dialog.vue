<template>
  <Teleport to="body">
    <Transition name="dialog-overlay">
      <div 
        v-if="visible" 
        class="ui-dialog-overlay" 
        @click="handleOverlayClick"
      >
        <Transition name="dialog-content">
          <div 
            v-if="visible"
            :class="['ui-dialog', { 'ui-dialog--mobile': isMobile }]" 
            @click.stop
          >
            <!-- 关闭按钮 -->
            <button 
              v-if="showCloseButton"
              class="ui-dialog__close"
              @click="handleClose"
              aria-label="关闭"
            >
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M15 5L5 15M5 5L15 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
            
            <!-- iOS 风格标题 -->
            <div class="ui-dialog__header" v-if="title || $slots.header">
              <slot name="header">
                <h3 class="ui-dialog__title">{{ title }}</h3>
              </slot>
            </div>
            
            <!-- iOS 风格内容 -->
            <div class="ui-dialog__body">
              <slot></slot>
            </div>
            
            <!-- iOS 风格底部按钮 -->
            <div class="ui-dialog__footer" v-if="$slots.footer">
              <slot name="footer"></slot>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

interface Props {
  modelValue: boolean
  title?: string
  closeOnClickOverlay?: boolean
  showCloseButton?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  closeOnClickOverlay: true,
  showCloseButton: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const isMobile = ref(window.innerWidth < 768)

const handleClose = () => {
  visible.value = false
  emit('close')
}

const handleOverlayClick = () => {
  if (props.closeOnClickOverlay) {
    handleClose()
  }
}

const handleEscape = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && visible.value) {
    handleClose()
  }
}

const updateMobile = () => {
  isMobile.value = window.innerWidth < 768
}

watch(visible, (val) => {
  if (val) {
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', handleEscape)
  } else {
    document.body.style.overflow = ''
    document.removeEventListener('keydown', handleEscape)
  }
})

onMounted(() => {
  window.addEventListener('resize', updateMobile)
})

onUnmounted(() => {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', handleEscape)
  window.removeEventListener('resize', updateMobile)
})
</script>

<style scoped>
/* Apple Store 风格遮罩层 - 柔和透明毛玻璃效果 */
/* 使用 !important 确保样式不被覆盖，渲染到 body 最顶层 */
.ui-dialog-overlay {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  bottom: 0 !important;
  background-color: rgba(0, 0, 0, 0.15) !important;
  backdrop-filter: blur(2px) saturate(110%) !important;
  -webkit-backdrop-filter: blur(2px) saturate(110%) !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  z-index: 9999 !important;
  padding: var(--spacing-lg) !important;
  overflow-y: auto !important;
  -webkit-overflow-scrolling: touch !important;
  margin: 0 !important;
  border: none !important;
  box-sizing: border-box !important;
}

/* Apple Store 风格对话框 - 柔和透明毛玻璃 */
/* 使用 !important 确保样式不被覆盖 */
.ui-dialog {
  background: rgba(255, 255, 255, 0.92) !important;
  backdrop-filter: blur(30px) saturate(170%) !important;
  -webkit-backdrop-filter: blur(30px) saturate(170%) !important;
  border-radius: 20px !important;
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.12),
    0 8px 24px rgba(0, 0, 0, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.7) !important;
  border: 0.5px solid rgba(255, 255, 255, 0.4) !important;
  max-width: 400px !important;
  width: 100% !important;
  max-height: 85vh !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
  position: relative !important;
  margin: auto !important;
  z-index: 10000 !important;
  box-sizing: border-box !important;
}

/* 优雅的关闭按钮 */
.ui-dialog__close {
  position: absolute;
  top: var(--spacing-md);
  right: var(--spacing-md);
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: rgba(0, 0, 0, 0.04);
  border-radius: 50%;
  cursor: pointer;
  color: rgba(0, 0, 0, 0.6);
  transition: all 0.2s cubic-bezier(0.4, 0.0, 0.2, 1);
  z-index: 10;
  padding: 0;
  flex-shrink: 0;
}

.ui-dialog__close:hover {
  background: rgba(0, 0, 0, 0.08);
  color: rgba(0, 0, 0, 0.8);
  transform: scale(1.05);
}

.ui-dialog__close:active {
  background: rgba(0, 0, 0, 0.12);
  transform: scale(0.95);
}

.ui-dialog__close svg {
  width: 16px;
  height: 16px;
  stroke-width: 2;
}

.ui-dialog--mobile {
  max-width: 100%;
  max-height: 90vh;
  border-radius: 20px 20px 0 0;
  margin-top: auto;
  box-shadow: 
    0 -8px 32px rgba(0, 0, 0, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
}

/* Apple Store 风格标题 */
.ui-dialog__header {
  padding: var(--spacing-xxl) var(--spacing-xxl) var(--spacing-lg);
  padding-top: calc(var(--spacing-xxl) + 8px);
  text-align: center;
}

.ui-dialog__title {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-family);
  color: var(--color-text-primary);
  line-height: 1.3;
  letter-spacing: -0.3px;
}

/* Apple Store 风格内容区域 */
.ui-dialog__body {
  flex: 1;
  padding: 0 var(--spacing-xxl) var(--spacing-lg);
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  text-align: center;
  color: var(--color-text-regular);
  font-size: var(--font-size-md);
  line-height: 1.5;
}

.ui-dialog__body :deep(p) {
  margin: 0 0 var(--spacing-sm);
  color: var(--color-text-regular);
}

.ui-dialog__body :deep(p:last-child) {
  margin-bottom: 0;
}

/* Apple Store 风格底部按钮区域 - 贴合底部 */
.ui-dialog__footer {
  padding: var(--spacing-lg) 0 0 0;
  border-top: none;
  display: flex;
  flex-direction: row;
  min-height: auto;
  margin-top: 0;
  overflow: hidden;
}

.ui-dialog__footer :deep(.ui-button) {
  flex: 1;
  border-radius: 0 !important;
  margin: 0 !important;
  font-weight: var(--font-weight-regular) !important;
  font-size: var(--font-size-md) !important;
  min-height: auto !important;
  height: auto !important;
  border: none !important;
  border-right: none !important;
  background: transparent !important;
  background-color: transparent !important;
  box-shadow: none !important;
  backdrop-filter: none !important;
  -webkit-backdrop-filter: none !important;
  transition: opacity var(--transition-fast) !important;
  padding: var(--spacing-sm) 0 !important;
  transform: none !important;
}

.ui-dialog__footer :deep(.ui-button:last-child) {
  border-right: none !important;
  border-radius: 0 0 20px 0 !important;
}

.ui-dialog__footer :deep(.ui-button:first-child) {
  border-radius: 0 0 0 20px !important;
}

/* iOS 风格确认按钮 - 贴合底部，圆角适配 */
.ui-dialog__footer :deep(.ui-button:only-child),
.ui-dialog__footer :deep(.ui-dialog__single-button) {
  border-radius: 0 0 20px 20px !important;
  border-right: none !important;
  color: #007AFF !important;
  font-weight: var(--font-weight-regular) !important;
  font-size: var(--font-size-md) !important;
  width: 100%;
  justify-content: center;
  background: transparent !important;
  background-color: transparent !important;
  box-shadow: none !important;
  transition: background-color 0.15s ease, opacity 0.15s ease !important;
  padding: var(--spacing-md) var(--spacing-xxl) !important;
  min-height: 52px !important;
  margin: 0 !important;
}

.ui-dialog__footer :deep(.ui-button:only-child:hover),
.ui-dialog__footer :deep(.ui-dialog__single-button:hover) {
  color: #007AFF !important;
  background-color: rgba(0, 0, 0, 0.03) !important;
  opacity: 1;
  transform: none !important;
}

.ui-dialog__footer :deep(.ui-button:only-child:active),
.ui-dialog__footer :deep(.ui-dialog__single-button:active) {
  color: #007AFF !important;
  background-color: rgba(0, 0, 0, 0.06) !important;
  opacity: 1;
  transform: none !important;
}

/* Apple Store 风格按钮 - 融合设计，带简单区分效果 */
.ui-dialog__footer :deep(.ui-button--default) {
  color: rgba(0, 0, 0, 0.7) !important;
  background: transparent !important;
  background-color: transparent !important;
  background-image: none !important;
  border: none !important;
  box-shadow: none !important;
  backdrop-filter: none !important;
  -webkit-backdrop-filter: none !important;
  transition: color 0.2s ease, opacity 0.2s ease !important;
}

/* iOS 风格按钮 - 主要按钮 */
.ui-dialog__footer :deep(.ui-button--primary) {
  color: #007AFF !important;
  font-weight: var(--font-weight-regular) !important;
  background: transparent !important;
  background-color: transparent !important;
  background-image: none !important;
  border: none !important;
  box-shadow: none !important;
  backdrop-filter: none !important;
  -webkit-backdrop-filter: none !important;
  transition: background-color 0.15s ease, opacity 0.15s ease !important;
  padding: var(--spacing-sm) var(--spacing-md) !important;
  min-height: 44px !important;
}

.ui-dialog__footer :deep(.ui-button--default:hover) {
  color: rgba(0, 0, 0, 0.85) !important;
  background: transparent !important;
  background-color: rgba(0, 0, 0, 0.03) !important;
  background-image: none !important;
  box-shadow: none !important;
  opacity: 1;
  transform: none !important;
}

.ui-dialog__footer :deep(.ui-button--primary:hover) {
  color: #007AFF !important;
  background: transparent !important;
  background-color: rgba(0, 0, 0, 0.03) !important;
  background-image: none !important;
  box-shadow: none !important;
  opacity: 1;
  transform: none !important;
}

.ui-dialog__footer :deep(.ui-button--default:active) {
  background: transparent !important;
  background-color: rgba(0, 0, 0, 0.06) !important;
  background-image: none !important;
  opacity: 1;
  transform: none !important;
  box-shadow: none !important;
}

.ui-dialog__footer :deep(.ui-button--primary:active) {
  color: #007AFF !important;
  background: transparent !important;
  background-color: rgba(0, 0, 0, 0.06) !important;
  background-image: none !important;
  opacity: 1;
  transform: none !important;
  box-shadow: none !important;
}

.ui-dialog__footer :deep(.ui-button--danger) {
  color: var(--color-danger) !important;
  background: transparent !important;
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

.ui-dialog__footer :deep(.ui-button--danger:active) {
  background-color: rgba(0, 0, 0, 0.05);
  opacity: 1;
}

/* Apple Store 风格动画 - 优雅流畅 */
.dialog-overlay-enter-active,
.dialog-overlay-leave-active {
  transition: opacity 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
}

.dialog-overlay-enter-from,
.dialog-overlay-leave-to {
  opacity: 0;
}

.dialog-content-enter-active {
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.dialog-content-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
}

.dialog-content-enter-from {
  transform: scale(0.92) translateY(10px);
  opacity: 0;
}

.dialog-content-leave-to {
  transform: scale(0.96) translateY(5px);
  opacity: 0;
}

/* 移动端样式 */
@media (max-width: 767px) {
  .ui-dialog-overlay {
    padding: 0;
    align-items: flex-end;
  }
  
  .ui-dialog {
    border-radius: 20px 20px 0 0;
    max-height: 90vh;
  }
  
  .ui-dialog__header {
    padding: var(--spacing-xl) var(--spacing-xl) var(--spacing-md);
  }
  
  .ui-dialog__body {
    padding: 0 var(--spacing-xl) var(--spacing-xl);
  }
  
  .dialog-content-enter-from {
    transform: translateY(100%);
  }
  
  .dialog-content-leave-to {
    transform: translateY(100%);
  }
}

/* Apple Store 风格 - 统一使用浅色主题，移除深色模式 */
</style>
