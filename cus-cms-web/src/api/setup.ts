import request from './request'

export interface StatusRes {
  initialized: boolean
}

export interface InitReq {
  username: string
  password: string
  nickname?: string
}

export interface InitRes {
  success: boolean
  message: string
}

export const setupApi = {
  getStatus(): Promise<StatusRes> {
    return request.get('/public/setup/status') as Promise<StatusRes>
  },
  init(data: InitReq): Promise<InitRes> {
    return request.post('/public/setup/init', data) as Promise<InitRes>
  },
}