<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import Chat from '$lib/components/room/Chat.svelte'
  import FilePlayer from '$lib/components/room/FilePlayer.svelte'
  import RoomLanding from '$lib/components/room/RoomLanding.svelte'
  import {
    connect,
    initialPlayerState,
    isIncomingChatMessage,
    isIncomingPlayerStateMessage,
    isIncomingRoomInfoMessage,
    isIncomingSubtitleMessage,
    isIncomingTypingIndicator,
    MessageType,
    RoomType,
    type ChatMessage,
    type GenericMessage,
    type WSHandlers,
    type PlayerState,
    type RoomInfo,
  } from '$lib/api/room'
  import * as soundEffects from '$lib/utils/soundEffects'
  import { SvelteMap } from 'svelte/reactivity'

  const systemUUID = '00000000-0000-0000-0000-000000000000'
  const timeout = 30000

  const id = page.params.id

  let messages: ChatMessage[] = $state([])
  let playerState = $state(initialPlayerState)
  let roomInfo: RoomInfo | null = $state(null)
  let subtitles: Record<string, string | null> = $state({})
  let typingIndicators = new SvelteMap<string, [number, number]>()

  let containerEl = $state(null) as Element | null
  let visibilityState = $state('visible') as DocumentVisibilityState
  let transientVideo: File | null = $state(null)
  let lastNotificationSound = Number.NEGATIVE_INFINITY
  let unreadMessageCount = $state(0)

  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null)
  let pongDeadline = 0
  let reconnectWait = $state(0)
  const wsInitialConnect = $derived((ws === null && !wsError) || roomInfo === null)

  const clientId = Math.random().toString(36).substring(2)

  const onMessage: WSHandlers['onMessage'] = function (event: MessageEvent) {
    // This sounds counterintuitive, but we don't want to handle messages if ws.close() was called.
    if (this.readyState === WebSocket.CLOSING || this.readyState === WebSocket.CLOSED) return
    try {
      if (typeof event.data !== 'string') throw new Error('Invalid message data type!')
      const message = JSON.parse(event.data) as GenericMessage
      if (isIncomingRoomInfoMessage(message)) {
        if (roomInfo === null) {
          roomInfo = message.data // On first run, we expect player state to come up afterwards
        } else {
          Object.assign(roomInfo, message.data)
          playerState = initialPlayerState
          subtitles = {}
        }
      } else if (isIncomingPlayerStateMessage(message)) {
        playerState = message.data
      } else if (isIncomingTypingIndicator(message)) {
        const existing = typingIndicators.get(message.userId)
        if (existing) clearTimeout(existing[1])
        const timeoutId = setTimeout(() => {
          if (typingIndicators.get(message.userId)?.[0] === message.timestamp) {
            typingIndicators.delete(message.userId)
          }
        }, 5000)
        typingIndicators.set(message.userId, [message.timestamp, timeoutId])
      } else if (isIncomingChatMessage(message)) {
        const lastKnownId = messages[messages.length - 1]?.id ?? -1
        const newMessagesIdx =
          lastKnownId === -1 ? -1 : message.data.findLastIndex(({ id }) => id === lastKnownId)
        const newMessages = message.data.slice(newMessagesIdx + 1)
        if (newMessages.length === 0) return
        for (const message of newMessages) {
          const typingIndicator = typingIndicators.get(message.userId)
          if (typingIndicator) {
            clearTimeout(typingIndicator[1])
            typingIndicators.delete(message.userId)
          }
        }
        const currentTimestamp = new Date().getTime()
        if (newMessages.length > 1) {
          soundEffects.join?.play().catch(console.warn) // A reconnect, just play the join sound
          lastNotificationSound = currentTimestamp
        } else {
          const { message, userId } = newMessages[0]
          if (userId === systemUUID) {
            if (message.endsWith('joined') || message.endsWith('reconnected')) {
              soundEffects.join?.play().catch(console.warn)
              lastNotificationSound = currentTimestamp
            } else if (message.endsWith('left') || message.endsWith('disconnected')) {
              soundEffects.leave?.play().catch(console.warn)
              lastNotificationSound = currentTimestamp
            }
          } else if (currentTimestamp - lastNotificationSound > 5000 && !document.hasFocus()) {
            soundEffects.message?.play().catch(console.warn)
            lastNotificationSound = currentTimestamp
          }
        }
        messages.push(...newMessages)
        if (visibilityState !== 'visible') unreadMessageCount += newMessages.length
      } else if (isIncomingSubtitleMessage(message)) {
        message.data.forEach(name => (subtitles[name] = null))
      } else if (message.type === MessageType.Pong) {
        pongDeadline = Date.now() + timeout // Reset pong deadline
      } else {
        console.warn('Unhandled message type!', message)
      }
    } catch (e) {
      console.error('Failed to parse backend message!', event, e)
    }
  }

  const onClose: WSHandlers['onClose'] = function (event: CloseEvent) {
    if (this !== ws) return
    wsError = event.reason || `WebSocket closed with code: ${event.code}`
  }

  onMount(() => {
    connect(id, clientId, { onMessage, onClose })
      .then(socket => {
        ws = socket
        pongDeadline = Date.now() + timeout // Reset pong deadline upon connect
      })
      .catch((e: unknown) => {
        if (e instanceof Error) wsError = e.message
      })
    const pingInterval = setInterval(() => {
      if (ws?.readyState === WebSocket.OPEN) {
        if (Date.now() > pongDeadline) {
          ws.close(4408, 'Connection timed out!')
          wsError = 'Connection timed out!'
        } else {
          ws.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }))
        }
      }
    }, 10000)
    return () => {
      clearInterval(pingInterval)
      ws?.close()
      typingIndicators.forEach(([, timeoutId]) => clearTimeout(timeoutId))
    }
  })

  $effect(() => {
    if (visibilityState === 'visible') unreadMessageCount = 0
  })

  // Reconnect if there's an error and the page is visible
  // TODO: Add Room not found! to this list.
  const isError = $derived(
    wsError && wsError !== 'You are not authenticated to access this resource!',
  ) // We don't care if the error message changed for this $effect, and don't reconnect if not authed.
  const reconnecting = $derived(isError ? reconnectWait : -1)
  $effect(() => {
    if (isError && visibilityState === 'visible') {
      let reconnectInterval = -1
      const reconnect = (isInterval: boolean) => async () => {
        if (isInterval && --reconnectWait > 0) return // Decrement 1 second every interval
        try {
          ws = await connect(id, clientId, { onMessage, onClose }, true)
          wsError = null
          pongDeadline = Date.now() + timeout // Reset pong deadline upon connect
        } catch (e: unknown) {
          if (e instanceof Error) wsError = e.message
          reconnectWait = 10 // TODO (low): Implement exponential backoff
          if (!isInterval) reconnectInterval = setInterval(reconnect(true), 1000)
        }
      }
      // TODO (low): Don't reset this whole thing when the page is made visible, maybe?
      if (wsInitialConnect) {
        reconnectWait = 10
        reconnectInterval = setInterval(reconnect(true), 1000)
      } else reconnect(false)() // eslint-disable-line @typescript-eslint/no-floating-promises
      return () => clearInterval(reconnectInterval)
    }
  })

  const onPlayerStateChange = (newState: PlayerState) => {
    ws?.send(JSON.stringify({ type: 'player_state', data: newState }))
  }

  let typingTimeout: number | null = null
  const onTyping = () => {
    if (typeof typingTimeout === 'number') return // Return if the function is throttled
    typingTimeout = setTimeout(() => (typingTimeout = null), 3000) // One typing msg every 3 seconds
    ws?.send(JSON.stringify({ type: 'typing', timestamp: Date.now() }))
  }

  const onSendMessage = (message: string) => {
    // Remove throttle if user sends a message
    if (typeof typingTimeout === 'number') {
      clearTimeout(typingTimeout)
      typingTimeout = null
    }
    ws?.send(JSON.stringify({ type: 'chat', data: message }))
  }
</script>

<svelte:document bind:visibilityState />
<svelte:head>
  <title>
    {(unreadMessageCount ? `(${unreadMessageCount}) ` : '') + (page.data.title as string)}
  </title>
</svelte:head>
<div class="container room" bind:this={containerEl}>
  {#if !roomInfo || roomInfo.type === RoomType.None}
    <RoomLanding bind:transientVideo error={wsError} {reconnecting} connecting={wsInitialConnect} />
  {:else if roomInfo.type === RoomType.LocalFile || roomInfo.type === RoomType.RemoteFile}
    {#key roomInfo.target}
      <FilePlayer
        bind:transientVideo
        {roomInfo}
        {playerState}
        bind:subtitles
        {onPlayerStateChange}
        error={wsError}
        {reconnecting}
        fullscreenEl={containerEl}
      />
    {/key}
  {:else}
    <RoomLanding bind:transientVideo error="Invalid room type!" {reconnecting} connecting={false} />
  {/if}
  <Chat
    disabled={wsError !== null || ws === null}
    {messages}
    {onSendMessage}
    {onTyping}
    {typingIndicators}
  />
</div>

<style lang="scss">
  .container {
    &:fullscreen,
    &::backdrop {
      background-color: var(--background-color);
    }
    max-height: calc(100vh - 4rem);
    flex: 1;
    display: flex;
    flex-direction: column;
    @media screen and (min-width: 768px) {
      flex-direction: row;
    }
  }
</style>
