export interface UserProfile {
  id: string
  nickname: string
  avatar: string
  bio: string
  following: number
  followers: number
  likes: number
}

export interface LoginPayload {
  mobile: string
  code: string
}

export interface LoginResult {
  token: string
  user: UserProfile
}
