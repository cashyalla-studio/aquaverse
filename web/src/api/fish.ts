import { apiClient } from './client'

export interface CreatureCategory {
  code: string
  name_ko: string
  name_en: string
  icon_emoji: string
  sort_order: number
}

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
  creature_category?: string
  extra_attributes?: {
    humidity_min?: number
    humidity_max?: number
    uv_requirement?: string
    basking_temp_c?: number
    substrate_type?: string
    enclosure_type?: string
    colony_size_min?: number
    colony_size_max?: number
    queen_required?: boolean
    venom_level?: string
    lifespan_years_min?: number
    lifespan_years_max?: number
    adult_size_cm?: number
    legal_status_kr?: string
    cites_appendix?: string
    notes?: string
  }
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
    category?: string
  }) => apiClient.get<FishListResponse>('/fish', { params }),

  get: (id: number, locale?: string) =>
    apiClient.get<FishDetail>(`/fish/${id}`, { params: { locale } }),

  search: (q: string, locale?: string, category?: string) =>
    apiClient.get<FishListItem[]>('/fish/search', { params: { q, locale, category } }),

  families: () => apiClient.get<string[]>('/fish/families'),

  categories: () => apiClient.get<CreatureCategory[]>('/fish/categories'),
}
