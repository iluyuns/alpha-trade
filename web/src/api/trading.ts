import request from '@/utils/request'

export interface TradingStatusResponse {
  enabled: boolean
  started: boolean
  mode: string
  symbols: string[]
  interval: string
  strategy: string
  message?: string
}

export interface TradingStartResponse {
  success: boolean
  message: string
}

export interface TradingStopResponse {
  success: boolean
  message: string
}

export const tradingAPI = {
  getStatus: () => {
    return request.get<TradingStatusResponse>('/trading/status')
  },

  start: () => {
    return request.post<TradingStartResponse>('/trading/start')
  },

  stop: () => {
    return request.post<TradingStopResponse>('/trading/stop')
  },
}
