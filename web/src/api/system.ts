import request from '@/utils/request'

export interface SystemInfoResponse {
  version: string
  build_time: string
  commit_hash: string
}

export const systemAPI = {
  getInfo: () => {
    return request.get<SystemInfoResponse>('/system/info')
  },
}
