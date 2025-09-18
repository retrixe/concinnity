import { PUBLIC_BACKEND_URL } from '$env/static/public'

export enum RoomType {
  None = '',
  LocalFile = 'local_file',
  RemoteFile = 'remote_file',
}

export interface ChatMessage {
  id: number
  userId: string
  message: string
  timestamp: string
}

export interface UserProfile {
  username: string
  avatar: string | null
}

export interface RoomInfo {
  id: string
  createdAt: string
  modifiedAt: string
  type: RoomType
  target: string
}

export interface PlayerState {
  paused: boolean
  speed: number
  timestamp: number
  lastAction: string
}

export const initialPlayerState: PlayerState = {
  paused: true,
  timestamp: 0,
  speed: 1,
  lastAction: new Date(0).toISOString(),
}

export interface WSHandlers {
  onClose?: (this: WebSocket, ev: CloseEvent) => void
  onError?: (this: WebSocket, ev: Event) => void
  onMessage?: (this: WebSocket, ev: MessageEvent) => void
}

export function connect(
  id: string,
  clientId: string,
  handlers: WSHandlers,
  reconnect = false,
): Promise<WebSocket> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(
      `${PUBLIC_BACKEND_URL.replace('http', 'ws')}/api/room/${id}/join`,
      'v0',
    )

    ws.onopen = () => {
      console.log('Connecting to room')
      // Send login message
      const token = localStorage.getItem('concinnity:token')
      ws.send(JSON.stringify({ token, reconnect, clientId }))
    }

    ws.onmessage = event => {
      console.log('Connected to room')
      // If the first thing we receive is an error, then reject
      try {
        if (typeof event.data !== 'string') throw new Error('Invalid message data type!')
        const message = JSON.parse(event.data) as { error?: string }
        if ('error' in message) {
          return reject(new Error(message.error))
        }
      } catch (e: unknown) {
        console.error('Incoming message data has malformed JSON!', e) // Err on the side of safety
        ws.close(1002, 'Incoming message data has malformed JSON!')
        return reject(new Error('Incoming message data has malformed JSON!'))
      }
      // Set new handlers
      if (handlers.onClose) ws.onclose = handlers.onClose.bind(ws)
      if (handlers.onError) ws.onerror = handlers.onError.bind(ws)
      if (handlers.onMessage) ws.onmessage = handlers.onMessage.bind(ws)
      // Handle current event
      ws.onmessage?.(event)
      // Resolve WebSocket
      resolve(ws)
    }

    ws.onclose = event => {
      reject(new Error('WebSocket closed abruptly! ' + event.reason || `Code: ${event.code}`))
    }
  })
}

export enum MessageType {
  Typing = 'typing',
  Chat = 'chat',
  RoomInfo = 'room_info',
  PlayerState = 'player_state',
  Subtitle = 'subtitle',
  Pong = 'pong',
  UserProfileUpdate = 'user_profile_update',
}

export interface GenericMessage {
  type: MessageType
}

export interface IncomingChatMessage extends GenericMessage {
  type: MessageType.Chat
  data: ChatMessage[]
}

export interface IncomingTypingIndicator extends GenericMessage {
  type: MessageType.Typing
  userId: string
  timestamp: number
}

export interface IncomingPlayerStateMessage extends GenericMessage {
  type: MessageType.PlayerState
  data: PlayerState
}

export interface IncomingRoomInfoMessage extends GenericMessage {
  type: MessageType.RoomInfo
  data: RoomInfo
}

export interface IncomingSubtitleMessage extends GenericMessage {
  type: MessageType.Subtitle
  data: string[]
}

export interface IncomingUserProfileUpdateMessage extends GenericMessage {
  type: MessageType.UserProfileUpdate
  id: string
  data: Partial<UserProfile>
}

export const isIncomingChatMessage = (message: GenericMessage): message is IncomingChatMessage =>
  message.type === MessageType.Chat && Array.isArray((message as IncomingChatMessage).data)

export const isIncomingTypingIndicator = (
  message: GenericMessage,
): message is IncomingTypingIndicator => message.type === MessageType.Typing

export const isIncomingPlayerStateMessage = (
  message: GenericMessage,
): message is IncomingPlayerStateMessage => message.type === MessageType.PlayerState

export const isIncomingRoomInfoMessage = (
  message: GenericMessage,
): message is IncomingRoomInfoMessage => message.type === MessageType.RoomInfo

export const isIncomingSubtitleMessage = (
  message: GenericMessage,
): message is IncomingSubtitleMessage => message.type === MessageType.Subtitle

export const isIncomingUserProfileUpdateMessage = (
  message: GenericMessage,
): message is IncomingUserProfileUpdateMessage => message.type === MessageType.UserProfileUpdate
