export interface Room {
  id: string
  createdAt: Date
  title: string
  type: 'localFile' | 'remoteFile'
  target: string
  chat: string[]
  members: string[]
  paused: boolean
  timestamp: number
  lastAction: Date
}
