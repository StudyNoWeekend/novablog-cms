import request from './request'
import type {
  Portfolio,
  PortfolioDetail,
  PortfolioListRes,
  CreatePortfolioReq,
  UpdatePortfolioReq,
  CreatePortfolioItemReq,
  UpdatePortfolioItemReq,
  SortPortfolioItemsReq,
} from '@/types/portfolio'

export const portfolioApi = {
  getList(params?: {
    page?: number
    page_size?: number
    keyword?: string
    status?: number
    category_id?: string
  }) {
    return request.get<PortfolioListRes>('/portfolios', { params })
  },
  getById(id: string) {
    return request.get<PortfolioDetail>(`/portfolios/${id}`)
  },
  create(data: CreatePortfolioReq) {
    return request.post<Portfolio>('/portfolios', data)
  },
  update(id: string, data: UpdatePortfolioReq) {
    return request.put<Portfolio>(`/portfolios/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/portfolios/${id}`)
  },
  addItem(id: string, data: CreatePortfolioItemReq) {
    return request.post(`/portfolios/${id}/items`, data)
  },
  updateItem(id: string, itemId: string, data: UpdatePortfolioItemReq) {
    return request.put(`/portfolios/${id}/items/${itemId}`, data)
  },
  removeItem(id: string, itemId: string) {
    return request.delete(`/portfolios/${id}/items/${itemId}`)
  },
  sortItems(id: string, data: SortPortfolioItemsReq) {
    return request.put(`/portfolios/${id}/items/sort`, data)
  },
}
