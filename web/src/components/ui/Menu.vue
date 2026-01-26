<template>
  <div :class="['ui-menu', { 'ui-menu--mobile': isMobile, 'ui-menu--open': isOpen }]">
    <button
      v-if="isMobile"
      class="ui-menu__trigger"
      @click="toggleMenu"
    >
      <slot name="trigger">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      </slot>
    </button>
    
    <div v-if="!isMobile || isOpen" class="ui-menu__content">
      <div
        v-if="isMobile"
        class="ui-menu__overlay"
        @click="closeMenu"
      ></div>
      <div class="ui-menu__list">
        <div
          v-for="item in items"
          :key="item.path"
          :class="['ui-menu__item', { 'ui-menu__item--active': activePath === item.path }]"
          @click="handleItemClick(item)"
        >
          <span v-if="item.icon" class="ui-menu__icon">
            <component :is="item.icon" />
          </span>
          <span class="ui-menu__label">{{ item.label }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export interface MenuItem {
  path: string
  label: string
  icon?: any
}

interface Props {
  items: MenuItem[]
  router?: boolean
  visible?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  router: true,
  visible: false,
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const route = useRoute()
const router = useRouter()
const isMobile = ref(window.innerWidth < 768)
const isOpen = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const activePath = computed(() => route.path)

const toggleMenu = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
}

const closeMenu = () => {
  isOpen.value = false
  document.body.style.overflow = ''
}

watch(() => props.visible, (val) => {
  if (val) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

const handleItemClick = (item: MenuItem) => {
  if (props.router) {
    router.push(item.path)
  }
  closeMenu()
}

const updateMobile = () => {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) {
    closeMenu()
  }
}

onMounted(() => {
  window.addEventListener('resize', updateMobile)
})

onUnmounted(() => {
  document.body.style.overflow = ''
  window.removeEventListener('resize', updateMobile)
})
</script>

<style scoped>
.ui-menu {
  position: relative;
}

.ui-menu__trigger {
  width: var(--touch-target);
  height: var(--touch-target);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-primary);
  border-radius: var(--border-radius);
  transition: background-color var(--transition-fast);
}

.ui-menu__trigger:active {
  background-color: var(--color-bg-page);
}

.ui-menu__content {
  position: relative;
}

.ui-menu--mobile .ui-menu__content {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: var(--z-index-modal);
  display: flex;
}

.ui-menu__overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-bg-overlay);
}

.ui-menu__list {
  position: relative;
  background-color: var(--color-bg);
  backdrop-filter: var(--glass-blur-strong);
  -webkit-backdrop-filter: var(--glass-blur-strong);
  width: 280px;
  max-width: 80vw;
  height: 100%;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15), 0 2px 8px rgba(0, 0, 0, 0.1);
  animation: slideInLeft var(--transition-base);
  border-right: 0.5px solid var(--color-separator);
}

.ui-menu__item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  color: var(--color-text-primary);
  cursor: pointer;
  transition: all var(--transition-base);
  min-height: var(--touch-target);
  border-left: none;
  font-size: var(--font-size-md);
  font-family: var(--font-family);
  position: relative;
}

.ui-menu__item:hover {
  background-color: var(--color-bg-secondary);
}

.ui-menu__item:active {
  background-color: var(--color-bg-tertiary);
  opacity: 0.7;
}

.ui-menu__item--active {
  background-color: var(--color-sidebar-selected);
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}

.ui-menu__icon {
  display: flex;
  align-items: center;
  font-size: 1.2em;
  flex-shrink: 0;
}

.ui-menu__label {
  flex: 1;
}

@keyframes slideInLeft {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(0);
  }
}

@media (min-width: 768px) {
  .ui-menu__list {
    position: static;
    width: 100%;
    box-shadow: none;
    animation: none;
    background-color: transparent;
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
    padding: var(--spacing-sm) 0;
    border-right: none;
  }
  
  .ui-menu__item {
    border-left: none;
    border-radius: var(--border-radius-lg);
    margin: 0 var(--spacing-sm) var(--spacing-xs);
  }
  
  .ui-menu__item--active {
    border-left: none;
    background-color: var(--color-sidebar-selected);
  }
}
</style>
