import request from '@/utils/request'
import type { DashboardData } from '@/stores/dashboard'

export const dashboardAPI = {
  getDashboard: () => {
    return request.get<DashboardData>('/dashboard')
  },
}
