export interface Room {
  id: string
  chat: string[]
  createdAt: Date
  extra: string
  lastActionTime: Date
  members: string[]
  paused: boolean
  timestamp: number
  title: string
  type: 'localFile' | 'youtube' | 'netflix'
}
