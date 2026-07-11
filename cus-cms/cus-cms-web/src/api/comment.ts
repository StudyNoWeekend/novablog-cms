import request from './request'
import type { Comment, CommentListReq, ReplyReq, UpdateCommentStatusReq } from '@/types/comment'
import type { PaginatedData } from '@/types/api'

export const commentApi = {
  getList(params: CommentListReq) {
    return request.get<PaginatedData<Comment>>('/comments', { params })
  },
  reply(id: string, data: ReplyReq) {
    return request.post<Comment>(`/comments/${id}/reply`, data)
  },
  remove(id: string) {
    return request.delete(`/comments/${id}`)
  },
  updateStatus(id: string, data: UpdateCommentStatusReq) {
    return request.put(`/comments/${id}/status`, data)
  },
}
