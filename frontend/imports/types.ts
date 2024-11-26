export interface Room {
  id: string
  chat: string[]
  extra: string
  members: string[]
  paused: boolean
  timestamp: number
  modifiedAt: Date
  createdAt: Date
  title: string
  type: 'localFile' | 'remoteFile'
}
