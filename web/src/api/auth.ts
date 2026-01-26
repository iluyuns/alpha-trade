import request from '@/utils/request'
import type { LoginResponse, UserInfo } from '@/stores/auth'

export const authAPI = {
  login: (username: string, password: string) => {
    return request.post<LoginResponse>('/auth/login', { username, password })
  },

  logout: () => {
    return request.post('/auth/logout')
  },

  oauth2Init: (provider: string) => {
    return request.get<{ redirect_url: string }>(`/auth/oauth2/init?provider=${provider}`)
  },

  oauth2Callback: (code: string, state: string) => {
    return request.get<LoginResponse>(`/auth/oauth2/callback?code=${code}&state=${state}`)
  },
}
