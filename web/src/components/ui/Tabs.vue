<template>
  <div class="ui-tabs">
    <div class="ui-tabs__header" ref="headerRef">
      <div
        v-for="(tab, index) in tabs"
        :key="tab.name"
        :class="['ui-tabs__item', { 'ui-tabs__item--active': activeTab === tab.name }]"
        @click="selectTab(tab.name)"
      >
        {{ tab.label }}
      </div>
      <div class="ui-tabs__indicator" :style="indicatorStyle"></div>
    </div>
    <div class="ui-tabs__content">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'

interface Tab {
  name: string
  label: string
}

interface Props {
  modelValue: string
  tabs: Tab[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
  change: [value: string]
}>()

const headerRef = ref<HTMLElement>()
const activeTab = ref(props.modelValue)
const indicatorStyle = ref({})

const selectTab = (name: string) => {
  activeTab.value = name
  emit('update:modelValue', name)
  emit('change', name)
  updateIndicator()
}

const updateIndicator = async () => {
  await nextTick()
  if (!headerRef.value) return
  
  const activeItem = headerRef.value.querySelector('.ui-tabs__item--active') as HTMLElement
  if (!activeItem) return
  
  const headerRect = headerRef.value.getBoundingClientRect()
  const itemRect = activeItem.getBoundingClientRect()
  
  indicatorStyle.value = {
    left: `${itemRect.left - headerRect.left}px`,
    width: `${itemRect.width}px`,
  }
}

watch(() => props.modelValue, (newVal) => {
  activeTab.value = newVal
  updateIndicator()
})

onMounted(() => {
  updateIndicator()
  window.addEventListener('resize', updateIndicator)
})
</script>

<style scoped>
.ui-tabs {
  width: 100%;
}

.ui-tabs__header {
  position: relative;
  display: flex;
  background-color: var(--color-bg-secondary);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  border-bottom: 0.5px solid var(--color-separator);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  border-radius: var(--border-radius-lg) var(--border-radius-lg) 0 0;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.05);
}

.ui-tabs__header::-webkit-scrollbar {
  display: none;
}

.ui-tabs__item {
  flex: 1;
  min-width: 80px;
  padding: var(--spacing-md) var(--spacing-lg);
  text-align: center;
  font-size: var(--font-size-md);
  font-family: var(--font-family);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-base);
  white-space: nowrap;
  position: relative;
  min-height: var(--touch-target);
  display: flex;
  align-items: center;
  justify-content: center;
}

.ui-tabs__item:active {
  opacity: 0.6;
}

.ui-tabs__item--active {
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
}

.ui-tabs__indicator {
  position: absolute;
  bottom: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-hover) 100%);
  border-radius: 2px 2px 0 0;
  transition: all var(--transition-base);
  box-shadow: 0 -2px 8px rgba(10, 132, 255, 0.3);
}

.ui-tabs__content {
  padding: var(--spacing-md);
}

@media (max-width: 767px) {
  .ui-tabs__item {
    padding: var(--spacing-sm) var(--spacing-md);
    font-size: var(--font-size-sm);
  }
  
  .ui-tabs__content {
    padding: var(--spacing-sm);
  }
}
</style>
