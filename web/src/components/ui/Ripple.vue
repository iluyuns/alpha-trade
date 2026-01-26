<template>
  <button
    :class="['ui-ripple', { 'ui-ripple--active': isRippling }]"
    @click="handleClick"
    @mousedown="handleMouseDown"
    @mouseup="handleMouseUp"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
  >
    <span v-if="isRippling" class="ui-ripple__effect" :style="rippleStyle"></span>
    <slot></slot>
  </button>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const isRippling = ref(false)
const rippleX = ref(0)
const rippleY = ref(0)

const rippleStyle = computed(() => ({
  left: `${rippleX.value}px`,
  top: `${rippleY.value}px`,
}))

const createRipple = (event: MouseEvent | TouchEvent) => {
  if (props.disabled) return
  
  const button = (event.currentTarget as HTMLElement)
  const rect = button.getBoundingClientRect()
  
  const clientX = 'touches' in event ? event.touches[0].clientX : event.clientX
  const clientY = 'touches' in event ? event.touches[0].clientY : event.clientY
  
  rippleX.value = clientX - rect.left
  rippleY.value = clientY - rect.top
  
  isRippling.value = true
  
  setTimeout(() => {
    isRippling.value = false
  }, 600)
}

const handleClick = (event: MouseEvent) => {
  if (!props.disabled) {
    emit('click', event)
  }
}

const handleMouseDown = (event: MouseEvent) => {
  createRipple(event)
}

const handleMouseUp = () => {
  // Ripple effect handled by timeout
}

const handleTouchStart = (event: TouchEvent) => {
  createRipple(event)
}

const handleTouchEnd = () => {
  // Ripple effect handled by timeout
}
</script>

<style scoped>
.ui-ripple {
  position: relative;
  overflow: hidden;
}

.ui-ripple__effect {
  position: absolute;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.6);
  transform: scale(0);
  animation: ripple 0.6s ease-out;
  pointer-events: none;
  width: 100px;
  height: 100px;
  margin-left: -50px;
  margin-top: -50px;
}

@keyframes ripple {
  to {
    transform: scale(4);
    opacity: 0;
  }
}
</style>
