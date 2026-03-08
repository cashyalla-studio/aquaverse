import { apiClient } from './client'

export interface BadgeDefinition {
  code: string
  name: string
  description: string
  icon_emoji: string
  category: 'care' | 'community' | 'market' | 'collection' | 'special'
  condition_type: string
  condition_value: number
  is_active: boolean
}

export interface UserBadge {
  id: number
  badge_code: string
  name: string
  icon_emoji: string
  category: string
  earned_at: string
}

export interface Challenge {
  id: number
  title: string
  description: string
  badge_code?: string
  starts_at: string
  ends_at: string
  condition_type: string
  condition_value: number
}

export interface ChallengeParticipant {
  challenge_id: number
  progress: number
  completed: boolean
  joined_at: string
  completed_at?: string
}

export const badgesApi = {
  listBadges: () => apiClient.get<BadgeDefinition[]>('/badges'),
  getMyBadges: () => apiClient.get<UserBadge[]>('/users/me/badges'),
  getUserBadges: (userId: string) => apiClient.get<UserBadge[]>(`/users/${userId}/badges`),
  listChallenges: () => apiClient.get<Challenge[]>('/challenges'),
  joinChallenge: (challengeId: number) => apiClient.post(`/challenges/${challengeId}/join`),
  getChallengeProgress: (challengeId: number) =>
    apiClient.get<ChallengeParticipant>(`/challenges/${challengeId}/progress`),
}
