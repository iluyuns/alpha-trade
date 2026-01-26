import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'
import router from '@/router'

export interface UserInfo {
  id: number
  username: string
  displayName: string
  avatar: string
}

export interface LoginResponse {
  status: string
  pendingToken?: string
  mfaType?: string
  token?: string
  user?: UserInfo
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<UserInfo | null>(() => {
    const userStr = localStorage.getItem('user')
    return userStr ? JSON.parse(userStr) : null
  })

  const isAuthenticated = () => {
    return !!token.value
  }

  const login = async (username: string, password: string): Promise<LoginResponse> => {
    const response = await request.post<LoginResponse>('/auth/login', {
      username,
      password,
    })

    if (response.status === 'success' && response.token && response.user) {
      token.value = response.token
      user.value = response.user
      localStorage.setItem('token', response.token)
      localStorage.setItem('user', JSON.stringify(response.user))
      return response
    }

    return response
  }

  const logout = async () => {
    try {
      await request.post('/auth/logout')
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      token.value = null
      user.value = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
    }
  }

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (newUser: UserInfo) => {
    user.value = newUser
    localStorage.setItem('user', JSON.stringify(newUser))
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
    setToken,
    setUser,
  }
})
