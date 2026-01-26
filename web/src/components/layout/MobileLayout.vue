<template>
  <div class="mobile-layout">
    <header class="mobile-layout__header safe-area-top">
      <button class="mobile-layout__menu-btn" @click="showMenu = true">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      </button>
      <h3 class="mobile-layout__title">Alpha-Trade</h3>
      <button class="mobile-layout__halt-btn" @click="handleGlobalHalt">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
      </button>
    </header>
    
    <main class="mobile-layout__content">
      <PullRefresh @refresh="handleRefresh">
        <slot></slot>
      </PullRefresh>
    </main>
    
    <MobileNav />
    
    <Menu :visible="showMenu" @update:visible="showMenu = $event" :items="menuItems" :router="true" />
    
    <Dialog v-model="showHaltDialog" title="紧急停机" @close="showHaltDialog = false">
      <p>确定要执行全局停机吗？这将停止所有交易活动。</p>
      <template #footer>
        <Button @click="showHaltDialog = false">取消</Button>
        <Button type="danger" @click="confirmHalt">停机</Button>
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import Menu from '@/components/ui/Menu.vue'
import MobileNav from './MobileNav.vue'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import PullRefresh from '@/components/ui/PullRefresh.vue'

const authStore = useAuthStore()
const showMenu = ref(false)
const showHaltDialog = ref(false)

const menuItems = [
  { path: '/dashboard', label: 'Dashboard' },
  { path: '/risk/config', label: '风控配置' },
  { path: '/orders', label: '订单管理' },
  { path: '/strategies', label: '策略管理' },
]

const handleGlobalHalt = () => {
  showHaltDialog.value = true
}

const confirmHalt = async () => {
  // TODO: 调用停机 API
  showHaltDialog.value = false
  if ((window as any).$toast) {
    ;(window as any).$toast.success('全局停机指令已发送')
  }
}

const handleRefresh = async () => {
  // TODO: 刷新数据
  await new Promise(resolve => setTimeout(resolve, 1000))
  if ((window as any).$toast) {
    ;(window as any).$toast.success('刷新成功')
  }
}
</script>

<style scoped>
.mobile-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background-color: var(--color-bg-page);
}

.mobile-layout__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--color-bg);
  border-bottom: 1px solid var(--color-border-light);
  padding: var(--spacing-sm) var(--spacing-md);
  height: 56px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: var(--z-index-sticky);
}

.mobile-layout__menu-btn,
.mobile-layout__halt-btn {
  width: var(--touch-target);
  height: var(--touch-target);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-primary);
  border-radius: var(--border-radius);
  transition: background-color var(--transition-fast);
}

.mobile-layout__menu-btn:active,
.mobile-layout__halt-btn:active {
  background-color: var(--color-bg-page);
}

.mobile-layout__title {
  flex: 1;
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--color-text-primary);
  text-align: center;
}

.mobile-layout__content {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  padding: var(--spacing-md);
  padding-bottom: calc(var(--spacing-md) + 60px + env(safe-area-inset-bottom));
}
</style>
