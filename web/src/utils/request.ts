import axios from 'axios'
import router from '@/router'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    // 所有错误信息输出到控制台
    console.error('请求错误:', error)
    if (error.response) {
      const { status, data } = error.response
      console.error('响应状态:', status)
      console.error('响应数据:', data)
      
      if (status === 401) {
        // Token 过期或未登录 - 静默处理，不显示错误提示
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        router.push('/login')
        // 不显示错误提示，让登录页面自己处理
      }
      // 其他错误只记录到控制台，不显示 Toast
      // 让业务代码自己决定是否显示用户友好的错误提示
    } else {
      console.error('网络错误:', error.message)
      // 网络错误只记录到控制台
    }
    return Promise.reject(error)
  }
)

export default request
