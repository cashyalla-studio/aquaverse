import { apiClient } from './client'

export interface CareSchedule {
  id: number
  tank_id: number
  schedule_type: 'feeding' | 'water_change' | 'filter_clean' | 'medication' | 'custom'
  title: string
  description?: string
  frequency: 'daily' | 'weekly' | 'biweekly' | 'monthly' | 'custom'
  interval_days?: number
  next_due_at: string
  last_done_at?: string
  is_active: boolean
}

export interface CareLog {
  id: number
  tank_id: number
  care_type: string
  notes?: string
  done_at: string
}

export interface CareStreak {
  user_id: string
  current_streak: number
  longest_streak: number
  last_care_date?: string
}

export interface TodayTasksResponse {
  date: string
  tasks: CareSchedule[]
  count: number
}

export interface SchedulesResponse {
  tank_id: number
  schedules: CareSchedule[]
}

export const careApi = {
  listSchedules: (tankId: number) =>
    apiClient.get<SchedulesResponse>(`/tanks/${tankId}/schedules`),

  createSchedule: (tankId: number, data: Partial<CareSchedule>) =>
    apiClient.post<CareSchedule>(`/tanks/${tankId}/schedules`, data),

  completeSchedule: (scheduleId: number, notes?: string) =>
    apiClient.post(`/schedules/${scheduleId}/complete`, { notes }),

  deleteSchedule: (scheduleId: number) =>
    apiClient.delete(`/schedules/${scheduleId}`),

  getTodayTasks: () =>
    apiClient.get<TodayTasksResponse>('/users/me/care-today'),

  getStreak: () =>
    apiClient.get<CareStreak>('/users/me/streak'),
}
