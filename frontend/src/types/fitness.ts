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
  type: 'strength' | 'cardio' | 'stretch'
  duration_min: number
  calories: number
  content: FitnessExercise[] | null
  notes: string
  status: number
  sort_order: number
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
  type?: 'strength' | 'cardio' | 'stretch'
  duration_min?: number
  calories?: number
  content?: FitnessExercise[]
  notes?: string
  status?: number
  sort_order?: number
}

export interface UpdateFitnessReq extends Partial<CreateFitnessReq> {}
