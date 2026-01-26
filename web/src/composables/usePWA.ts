import { ref, onMounted } from 'vue'
import { registerSW } from 'virtual:pwa-register'

export function usePWA() {
  const isInstalled = ref(false)
  const canInstall = ref(false)
  const deferredPrompt = ref<any>(null)
  const updateSW = ref<(() => Promise<void>) | null>(null)
  const needRefresh = ref(false)

  const checkInstalled = () => {
    // 检查是否已安装
    if (window.matchMedia('(display-mode: standalone)').matches) {
      isInstalled.value = true
    }
    
    // 检查是否可以通过 beforeinstallprompt 安装
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault()
      deferredPrompt.value = e
      canInstall.value = true
    })
  }

  const install = async () => {
    if (!deferredPrompt.value) {
      return false
    }

    deferredPrompt.value.prompt()
    const { outcome } = await deferredPrompt.value.userChoice
    
    if (outcome === 'accepted') {
      isInstalled.value = true
      canInstall.value = false
      deferredPrompt.value = null
      return true
    }
    
    return false
  }

  const checkUpdate = () => {
    if (updateSW.value) {
      updateSW.value()
    }
  }

  onMounted(() => {
    checkInstalled()
    
    // 注册 Service Worker
    updateSW.value = registerSW({
      immediate: true,
      onNeedRefresh() {
        needRefresh.value = true
      },
      onOfflineReady() {
        console.log('PWA offline ready')
      },
    })
  })

  return {
    isInstalled,
    canInstall,
    needRefresh,
    install,
    checkUpdate,
  }
}
