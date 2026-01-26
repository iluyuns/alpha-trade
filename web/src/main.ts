import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import './styles/variables.css'
import './styles/reset.css'
import './styles/mobile.css'
import './style.css'
import App from './App.vue'
import router from './router'
import ToastContainer from './components/ui/ToastContainer.vue'

const app = createApp(App)
const pinia = createPinia()

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.component('ToastContainer', ToastContainer)

app.mount('#app')
