<template>
  <component :is="layoutComponent">
    <slot></slot>
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useResponsive } from '@/composables/useResponsive'
import MobileLayout from './MobileLayout.vue'
import DesktopLayout from './DesktopLayout.vue'

const { isMobile } = useResponsive()

const layoutComponent = computed(() => {
  return isMobile.value ? MobileLayout : DesktopLayout
})
</script>
