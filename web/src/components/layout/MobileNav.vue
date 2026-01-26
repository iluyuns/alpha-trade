<template>
  <nav class="mobile-nav safe-area-bottom">
    <router-link
      v-for="item in navItems"
      :key="item.path"
      :to="item.path"
      :class="['mobile-nav__item', { 'mobile-nav__item--active': isActive(item.path) }]"
    >
      <span class="mobile-nav__icon">
        <component :is="item.icon" />
      </span>
      <span class="mobile-nav__label">{{ item.label }}</span>
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const navItems = [
  {
    path: '/dashboard',
    label: 'Dashboard',
    icon: () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor' }, [
      h('rect', { x: 3, y: 3, width: 7, height: 7 }),
      h('rect', { x: 14, y: 3, width: 7, height: 7 }),
      h('rect', { x: 14, y: 14, width: 7, height: 7 }),
      h('rect', { x: 3, y: 14, width: 7, height: 7 }),
    ]),
  },
  {
    path: '/orders',
    label: '订单',
    icon: () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor' }, [
      h('path', { d: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z' }),
      h('polyline', { points: '14 2 14 8 20 8' }),
      h('line', { x1: 16, y1: 13, x2: 8, y2: 13 }),
      h('line', { x1: 16, y1: 17, x2: 8, y2: 17 }),
    ]),
  },
  {
    path: '/strategies',
    label: '策略',
    icon: () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor' }, [
      h('polygon', { points: '12 2 2 7 12 12 22 7 12 2' }),
      h('polyline', { points: '2 17 12 22 22 17' }),
      h('polyline', { points: '2 12 12 17 22 12' }),
    ]),
  },
  {
    path: '/risk/config',
    label: '设置',
    icon: () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor' }, [
      h('circle', { cx: 12, cy: 12, r: 3 }),
      h('path', { d: 'M12 1v6m0 6v6m9-9h-6m-6 0H3m15.364 6.364l-4.243-4.243m-4.242 0L5.636 17.364m12.728 0l-4.243-4.243m-4.242 0L5.636 6.636' }),
    ]),
  },
]

const isActive = (path: string) => {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<style scoped>
.mobile-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background-color: var(--color-bg);
  border-top: 1px solid var(--color-border-light);
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
  z-index: var(--z-index-fixed);
  height: 60px;
}

.mobile-nav__item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: var(--spacing-xs);
  color: var(--color-text-secondary);
  transition: color var(--transition-fast);
  text-decoration: none;
  min-height: var(--touch-target);
}

.mobile-nav__item--active {
  color: var(--color-primary);
}

.mobile-nav__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.mobile-nav__label {
  font-size: var(--font-size-xs);
  font-weight: 500;
}
</style>
