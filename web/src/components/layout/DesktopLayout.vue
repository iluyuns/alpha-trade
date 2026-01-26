<template>
  <div class="desktop-layout">
    <aside class="desktop-layout__sidebar">
      <div class="desktop-layout__logo">
        <h3>Alpha-Trade</h3>
      </div>
      <Menu :items="menuItems" :router="true" />
    </aside>
    
    <div class="desktop-layout__main">
      <header class="desktop-layout__header">
        <div class="desktop-layout__header-left">
          <h3>Command Center</h3>
        </div>
        <div class="desktop-layout__header-right">
          <Button type="danger" size="small" @click="handleGlobalHalt">
            <template #icon>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                <line x1="12" y1="9" x2="12" y2="13"></line>
                <line x1="12" y1="17" x2="12.01" y2="17"></line>
              </svg>
            </template>
            紧急停机
          </Button>
          <div class="desktop-layout__user" @click="showUserMenu = !showUserMenu">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <span>{{ user?.displayName || user?.username }}</span>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
            <div v-if="showUserMenu" class="desktop-layout__user-menu">
              <div class="desktop-layout__user-menu-item" @click="handleLogout">退出登录</div>
            </div>
          </div>
        </div>
      </header>
      
      <main class="desktop-layout__content">
        <slot></slot>
      </main>
    </div>
    
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
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Menu from '@/components/ui/Menu.vue'
import Button from '@/components/ui/Button.vue'
import Dialog from '@/components/ui/Dialog.vue'

const router = useRouter()
const authStore = useAuthStore()
const showUserMenu = ref(false)
const showHaltDialog = ref(false)

const user = computed(() => authStore.user)

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

const handleLogout = async () => {
  showUserMenu.value = false
  await authStore.logout()
  if ((window as any).$toast) {
    ;(window as any).$toast.success('已退出登录')
  }
}
</script>

<style scoped>
.desktop-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.desktop-layout__sidebar {
  width: 240px;
  background-color: var(--color-sidebar-bg);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 0.5px solid var(--color-separator);
  box-shadow: 1px 0 0 rgba(0, 0, 0, 0.03);
}

.desktop-layout__sidebar :deep(.ui-menu) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.desktop-layout__sidebar :deep(.ui-menu__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.desktop-layout__sidebar :deep(.ui-menu__list) {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.desktop-layout__logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-primary);
  background-color: transparent;
  padding: var(--spacing-md);
  border-bottom: 0.5px solid var(--color-separator);
}

.desktop-layout__logo h3 {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-family);
}

.desktop-layout__main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.desktop-layout__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: rgba(255, 255, 255, var(--glass-opacity));
  backdrop-filter: var(--glass-blur-strong);
  -webkit-backdrop-filter: var(--glass-blur-strong);
  border-bottom: 0.5px solid var(--color-separator);
  padding: 0 var(--spacing-lg);
  height: 64px;
  flex-shrink: 0;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.05);
}

.desktop-layout__header-left h3 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-family);
}

@media (prefers-color-scheme: dark) {
  .desktop-layout__header {
    background-color: rgba(28, 28, 30, var(--glass-opacity));
    box-shadow: 0 1px 0 rgba(255, 255, 255, 0.05);
  }
}

.desktop-layout__header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.desktop-layout__user {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius);
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.desktop-layout__user:hover {
  background-color: var(--color-bg-page);
}

.desktop-layout__user-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: var(--spacing-xs);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-md);
  min-width: 120px;
  overflow: hidden;
}

.desktop-layout__user-menu-item {
  padding: var(--spacing-md);
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.desktop-layout__user-menu-item:hover {
  background-color: var(--color-bg-page);
}

.desktop-layout__content {
  flex: 1;
  background: var(--color-bg-page-gradient, var(--color-bg-page));
  padding: var(--spacing-xl);
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  position: relative;
}
</style>
