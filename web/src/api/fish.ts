import { apiClient } from './client'

export interface FishListItem {
  id: number
  scientific_name: string
  common_name: string
  family: string
  care_level?: string
  temperament?: string
  max_size_cm?: number
  min_tank_size_liters?: number
  primary_image_url?: string
  quality_score: number
}

export interface FishDetail {
  id: number
  scientific_name: string
  genus: string
  species: string
  family: string
  primary_common_name: string
  care_level?: string
  temperament?: string
  max_size_cm?: number
  lifespan_years?: number
  ph_min?: number
  ph_max?: number
  temp_min_c?: number
  temp_max_c?: number
  min_tank_size_liters?: number
  diet_type?: string
  diet_notes?: string
  breeding_notes?: string
  care_notes?: string
  primary_image_url?: string
  license?: string
  attribution?: string
  // 번역된 필드 (API가 locale에 맞춰 반환)
  translation?: {
    common_name?: string
    care_notes?: string
    breeding_notes?: string
    diet_notes?: string
  }
}

export interface FishListResponse {
  items: FishListItem[]
  total_count: number
  page: number
  limit: number
}

export const fishApi = {
  list: (params: {
    page?: number
    limit?: number
    family?: string
    care_level?: string
    q?: string
    locale?: string
  }) => apiClient.get<FishListResponse>('/fish', { params }),

  get: (id: number, locale?: string) =>
    apiClient.get<FishDetail>(`/fish/${id}`, { params: { locale } }),

  search: (q: string, locale?: string) =>
    apiClient.get<FishListItem[]>('/fish/search', { params: { q, locale } }),

  families: () => apiClient.get<string[]>('/fish/families'),
}
