<template>
  <div class="pull-refresh" ref="containerRef" @touchstart="handleTouchStart" @touchmove="handleTouchMove" @touchend="handleTouchEnd">
    <div
      class="pull-refresh__indicator"
      :style="{ transform: `translateY(${pullDistance}px)`, opacity: pullDistance > 0 ? 1 : 0 }"
    >
      <svg v-if="isRefreshing" class="pull-refresh__spinner" width="20" height="20" viewBox="0 0 50 50">
        <circle class="path" cx="25" cy="25" r="20" fill="none" stroke-width="4"></circle>
      </svg>
      <span v-else>{{ pullText }}</span>
    </div>
    <div class="pull-refresh__content" :style="{ transform: `translateY(${pullDistance}px)` }">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  disabled?: boolean
  pullingText?: string
  loosingText?: string
  loadingText?: string
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  pullingText: '下拉刷新',
  loosingText: '释放刷新',
  loadingText: '刷新中...',
})

const emit = defineEmits<{
  refresh: []
}>()

const containerRef = ref<HTMLElement>()
const startY = ref(0)
const pullDistance = ref(0)
const isRefreshing = ref(false)
const isPulling = ref(false)

const PULL_THRESHOLD = 50

const pullText = computed(() => {
  if (isRefreshing.value) return props.loadingText
  if (pullDistance.value >= PULL_THRESHOLD) return props.loosingText
  return props.pullingText
})

const handleTouchStart = (e: TouchEvent) => {
  if (props.disabled || isRefreshing.value) return
  
  const touch = e.touches[0]
  startY.value = touch.clientY
  isPulling.value = true
}

const handleTouchMove = (e: TouchEvent) => {
  if (!isPulling.value || props.disabled || isRefreshing.value) return
  
  const touch = e.touches[0]
  const deltaY = touch.clientY - startY.value
  
  if (deltaY > 0 && window.scrollY === 0) {
    e.preventDefault()
    pullDistance.value = Math.min(deltaY * 0.5, PULL_THRESHOLD * 1.5)
  }
}

const handleTouchEnd = () => {
  if (!isPulling.value || props.disabled) return
  
  isPulling.value = false
  
  if (pullDistance.value >= PULL_THRESHOLD) {
    isRefreshing.value = true
    emit('refresh')
  } else {
    pullDistance.value = 0
  }
}

const finishRefresh = () => {
  isRefreshing.value = false
  pullDistance.value = 0
}

defineExpose({ finishRefresh })
</script>

<style scoped>
.pull-refresh {
  position: relative;
  overflow: hidden;
}

.pull-refresh__indicator {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  padding: var(--spacing-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  transition: opacity var(--transition-fast);
  z-index: 1;
}

.pull-refresh__spinner {
  animation: rotate 1s linear infinite;
}

.pull-refresh__spinner .path {
  stroke: currentColor;
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

.pull-refresh__content {
  transition: transform var(--transition-base);
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
</style>
