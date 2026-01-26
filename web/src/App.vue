<template>
  <ResponsiveLayout v-if="isRouterReady && showLayout">
    <router-view />
  </ResponsiveLayout>
  <router-view v-else-if="isRouterReady" />
  <ToastContainer />
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ResponsiveLayout from '@/components/layout/ResponsiveLayout.vue'
import ToastContainer from '@/components/ui/ToastContainer.vue'

const route = useRoute()
const router = useRouter()
const isRouterReady = ref(false)

// 等待路由初始化完成，避免刷新时闪烁
onMounted(async () => {
  await router.isReady()
  isRouterReady.value = true
})

// 使用 path 判断更可靠，避免路由初始化时 name 未设置导致的闪烁
// route.path 在路由初始化时就会设置，比 route.name 更可靠
const showLayout = computed(() => {
  const path = route.path
  // 明确排除登录页面，避免初始渲染时显示布局
  return path !== '/login' && !path.startsWith('/login')
})
</script>

<style>
#app {
  width: 100%;
  height: 100vh;
  margin: 0;
  padding: 0;
}
</style>
