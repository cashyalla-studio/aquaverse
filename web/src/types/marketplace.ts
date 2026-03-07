export interface Listing {
  id: number
  seller_id: string
  fish_data_id?: number
  scientific_name?: string
  common_name: string
  quantity: number
  age_months?: number
  size_cm?: number
  sex: 'MALE' | 'FEMALE' | 'UNKNOWN' | 'MIXED'
  health_status: 'EXCELLENT' | 'GOOD' | 'DISEASE_HISTORY' | 'UNDER_TREATMENT'
  disease_history?: string
  bred_by_seller: boolean
  price: string | number
  currency: string
  price_negotiable: boolean
  trade_type: 'DIRECT' | 'COURIER' | 'AQUA_COURIER' | 'ALL'
  allow_international: boolean
  allowed_countries?: string[]
  location_text: string
  country_code: string
  title: string
  description?: string
  image_urls: string[]
  status: string
  view_count: number
  favorite_count: number
  distance_km?: number
  seller_trust_score?: number
  created_at: string
}

export interface Trade {
  id: number
  listing_id: number
  seller_id: string
  buyer_id: string
  trade_type: string
  agreed_price: string
  currency: string
  escrow_enabled: boolean
  status: string
  created_at: string
}
