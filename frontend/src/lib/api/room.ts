import { PUBLIC_BACKEND_URL } from '$env/static/public'

export enum RoomType {
  None = '',
  LocalFile = 'local_file',
  RemoteFile = 'remote_file',
}

export interface ChatMessage {
  userId: string
  message: string
  timestamp: string
}

interface Handlers {
  onClose: (this: WebSocket, ev: CloseEvent) => void
  onError: (this: WebSocket, ev: Event) => void
  onMessage: (this: WebSocket, ev: MessageEvent) => void
}

export function connect(id: string, handlers: Handlers, reconnect = false): Promise<WebSocket> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(
      `${PUBLIC_BACKEND_URL.replace('http', 'ws')}/api/room/${id}/join`,
      'v0',
    )

    ws.onopen = () => {
      console.log('Connecting to room')
      // Send login message
      ws.send(JSON.stringify({ token: localStorage.getItem('concinnity:token'), reconnect }))
    }

    ws.onmessage = event => {
      console.log('Connected to room')
      // Set new handlers
      ws.onclose = handlers.onClose.bind(ws)
      ws.onerror = handlers.onError.bind(ws)
      ws.onmessage = handlers.onMessage.bind(ws)
      // Handle current event
      ws.onmessage(event)
      // Resolve WebSocket
      resolve(ws)
    }

    ws.onclose = event => {
      reject(new Error('WebSocket closed abruptly! ' + event.reason || `Code: ${event.code}`))
    }
  })
}
