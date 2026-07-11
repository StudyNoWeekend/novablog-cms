export interface Comment {
  id: string
  target_type: 'article' | 'travel'
  target_id: string
  target_title: string
  parent_id: string | null
  blogger_id: string | null
  nickname: string
  email: string
  avatar: string
  content: string
  is_blogger: boolean
  status: number // 1=待审核 2=已通过 3=已拒绝
  ip_address: string
  created_at: string
  updated_at: string
}

export interface CommentListReq {
  page?: number
  page_size?: number
  target_type?: string
  target_id?: string
  status?: number
  keyword?: string
}

export interface ReplyReq {
  content: string
}

export interface UpdateCommentStatusReq {
  status: number
}
