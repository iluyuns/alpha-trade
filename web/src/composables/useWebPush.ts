import { ref, onMounted } from 'vue'
import request from '@/utils/request'

export function useWebPush() {
  const isSupported = ref('Notification' in window && 'serviceWorker' in navigator)
  const permission = ref<NotificationPermission>('default')
  const subscription = ref<PushSubscription | null>(null)
  const isSubscribed = ref(false)

  const checkPermission = async () => {
    if (!isSupported.value) {
      return false
    }
    
    permission.value = Notification.permission
    
    if ('serviceWorker' in navigator) {
      const registration = await navigator.serviceWorker.ready
      const sub = await registration.pushManager.getSubscription()
      subscription.value = sub
      isSubscribed.value = !!sub
    }
    
    return permission.value === 'granted'
  }

  const requestPermission = async () => {
    if (!isSupported.value) {
      throw new Error('浏览器不支持通知功能')
    }

    const result = await Notification.requestPermission()
    permission.value = result
    
    if (result === 'granted') {
      await subscribe()
    }
    
    return result === 'granted'
  }

  const subscribe = async () => {
    if (!isSupported.value || permission.value !== 'granted') {
      throw new Error('通知权限未授予')
    }

    const registration = await navigator.serviceWorker.ready
    
    const sub = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(getVapidPublicKey()),
    })

    subscription.value = sub
    
    // 发送订阅信息到后端
    try {
      await request.post('/notifications/subscribe', {
        subscription: JSON.stringify(sub),
      })
      isSubscribed.value = true
      return true
    } catch (error) {
      console.error('Failed to subscribe:', error)
      await unsubscribe()
      return false
    }
  }

  const unsubscribe = async () => {
    if (subscription.value) {
      await subscription.value.unsubscribe()
      subscription.value = null
      isSubscribed.value = false
      
      try {
        await request.post('/notifications/unsubscribe', {
          subscription: JSON.stringify(subscription.value),
        })
      } catch (error) {
        console.error('Failed to unsubscribe:', error)
      }
    }
  }

  const getVapidPublicKey = () => {
    // TODO: 从环境变量或配置获取 VAPID public key
    return import.meta.env.VITE_VAPID_PUBLIC_KEY || ''
  }

  const urlBase64ToUint8Array = (base64String: string) => {
    const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
    const rawData = window.atob(base64)
    const outputArray = new Uint8Array(rawData.length)
    
    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i)
    }
    
    return outputArray
  }

  // 监听推送消息
  const setupMessageListener = () => {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.addEventListener('message', (event) => {
        const data = event.data
        if (data && data.type === 'PUSH_NOTIFICATION') {
          showNotification(data.title, data.options)
        }
      })
    }
  }

  const showNotification = (title: string, options?: NotificationOptions) => {
    if (permission.value === 'granted') {
      new Notification(title, {
        icon: '/icons/icon-192.png',
        badge: '/icons/icon-192.png',
        ...options,
      })
    }
  }

  onMounted(async () => {
    await checkPermission()
    setupMessageListener()
  })

  return {
    isSupported,
    permission,
    isSubscribed,
    requestPermission,
    subscribe,
    unsubscribe,
    showNotification,
  }
}
