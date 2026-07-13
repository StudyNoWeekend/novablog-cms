export interface Tag {
  id: string
  name: string
  created_at: string
}

export interface TagCreateReq {
  name: string
}

export interface TagUpdateReq {
  name: string
}