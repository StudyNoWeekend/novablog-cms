export interface FitnessExercise {
  name: string
  sets?: string
  reps?: string
  note?: string
}

export interface FitnessRecord {
  id: string
  date: string
  title: string
  /** 兼容镜像：通常等于分类 slug（strength/cardio/stretch），新数据以 category_id 为准 */
  type: string
  duration_min: number
  calories: number
  content: FitnessExercise[] | null
  notes: string
  status: number
  sort_order: number
  category_id: string
  category_name: string
  created_at: string
  updated_at: string
}

export interface FitnessListRes {
  list: FitnessRecord[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CreateFitnessReq {
  date?: string
  title: string
  category_id?: string
  /** 兼容保留，建议用 category_id */
  type?: 'strength' | 'cardio' | 'stretch'
  duration_min?: number
  calories?: number
  content?: FitnessExercise[]
  notes?: string
  status?: number
  sort_order?: number
}

export interface UpdateFitnessReq extends Partial<CreateFitnessReq> {}
