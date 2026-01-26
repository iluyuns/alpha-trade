import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/utils/request'
import { dashboardWS } from '@/utils/websocket'

export interface SystemHealthItem {
  name: string
  status: 'normal' | 'warning' | 'error'
  latency?: number
  lastHeartbeat?: string
  message?: string
}

export interface RiskStatus {
  consecutiveLosses: number
  maxConsecutiveLosses: number
  macroCoolingMode: 'active' | 'inactive'
  nextMacroWindow?: string
  leverageStatus: 'relaxed' | 'restricted'
  maxLeverage: number
  currentLeverage: number
}

export interface StrategyOverview {
  id: string
  name: string
  symbol: string
  direction: 'Long' | 'Short' | 'N/A'
  status: 'running' | 'stopped' | 'cooling'
  winRate: number
  reason?: string
}

export interface DashboardData {
  pnlDaily: string
  pnlPercent: string
  totalEquity: string
  riskExposure: string
  dailyDrawdown: string
  systemHealth: SystemHealthItem[]
  riskStatus: RiskStatus
  strategies: StrategyOverview[]
}

export const useDashboardStore = defineStore('dashboard', () => {
  const data = ref<DashboardData | null>(null)
  const loading = ref(false)
  const wsConnected = ref(false)

  const fetchDashboard = async () => {
    loading.value = true
    try {
      const response = await request.get<DashboardData>('/dashboard')
      data.value = response
    } catch (error) {
      console.error('Failed to fetch dashboard:', error)
    } finally {
      loading.value = false
    }
  }

  const connectWebSocket = (token: string) => {
    dashboardWS.connect(token)

    // 订阅 Dashboard 数据更新
    dashboardWS.subscribe('dashboard', (newData: DashboardData) => {
      data.value = newData
    })

    // 订阅订单状态更新
    dashboardWS.subscribe('order', (orderData: any) => {
      // 更新策略或订单相关数据
      console.log('Order update:', orderData)
    })

    // 订阅风控状态更新
    dashboardWS.subscribe('risk', (riskData: any) => {
      if (data.value) {
        data.value.riskStatus = riskData
      }
    })

    wsConnected.value = true
  }

  const disconnectWebSocket = () => {
    dashboardWS.disconnect()
    wsConnected.value = false
  }

  return {
    data,
    loading,
    wsConnected,
    fetchDashboard,
    connectWebSocket,
    disconnectWebSocket,
  }
})
