import { apiRequest } from './http'
import type { LoginResponse, RegisterResponse, User } from './types'

export type LoginPayload = {
  email: string
  password: string
}

export type RegisterPayload = {
  email: string
  phone?: string
  password: string
}

export const authApi = {
  login(payload: LoginPayload) {
    return apiRequest<LoginResponse>('/auth/login', {
      method: 'POST',
      body: payload,
      auth: false,
    })
  },
  register(payload: RegisterPayload) {
    return apiRequest<RegisterResponse>('/users', {
      method: 'POST',
      body: payload,
      auth: false,
    })
  },
  me() {
    return apiRequest<User>('/users/me')
  },
}
