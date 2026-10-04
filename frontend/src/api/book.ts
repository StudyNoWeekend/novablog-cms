import request from './request'
import type { Book, BookListRes, CreateBookReq, UpdateBookReq } from '@/types/book'

export const bookApi = {
  getList(params?: { page?: number; page_size?: number; keyword?: string; reading_status?: string; status?: number }) {
    return request.get<BookListRes>('/books', { params })
  },
  getById(id: string) {
    return request.get<Book>(`/books/${id}`)
  },
  create(data: CreateBookReq) {
    return request.post<Book>('/books', data)
  },
  update(id: string, data: UpdateBookReq) {
    return request.put<Book>(`/books/${id}`, data)
  },
  remove(id: string) {
    return request.delete(`/books/${id}`)
  },
}
