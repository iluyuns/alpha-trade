import { ref, onMounted, onUnmounted } from 'vue'

export function useResponsive() {
  const isMobile = ref(window.innerWidth < 768)
  const isTablet = ref(window.innerWidth >= 768 && window.innerWidth < 1024)
  const isDesktop = ref(window.innerWidth >= 1024)
  
  const update = () => {
    const width = window.innerWidth
    isMobile.value = width < 768
    isTablet.value = width >= 768 && width < 1024
    isDesktop.value = width >= 1024
  }
  
  onMounted(() => {
    window.addEventListener('resize', update)
    update()
  })
  
  onUnmounted(() => {
    window.removeEventListener('resize', update)
  })
  
  return {
    isMobile,
    isTablet,
    isDesktop,
  }
}
