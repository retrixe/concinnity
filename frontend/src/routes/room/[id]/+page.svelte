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
    type PlayerState,
    type RoomInfo,
  } from '$lib/api/room'
  import * as soundEffects from '$lib/utils/soundEffects'
  import { SvelteMap } from 'svelte/reactivity'

  const systemUUID = '00000000-0000-0000-0000-000000000000'

  // TODO: Support watching remote files
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
  const wsInitialConnect = $derived((ws === null && !wsError) || roomInfo === null)

  const clientId = Math.random().toString(36).substring(2)

  const onMessage = (event: MessageEvent) => {
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
        // TODO (low): Replace messages.length with IDs
        const newMessages = message.data.slice(message.data.length === 1 ? 0 : messages.length)
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
      } else if (message.type !== MessageType.Pong) {
        console.warn('Unhandled message type!', message)
      }
    } catch (e) {
      console.error('Failed to parse backend message!', event, e)
    }
  }

  const onClose = (event: CloseEvent) => {
    wsError = event.reason || `WebSocket closed with code: ${event.code}`
  }

  onMount(() => {
    connect(id, clientId, { onMessage, onClose })
      .then(socket => {
        ws = socket
      })
      .catch((e: unknown) => {
        if (e instanceof Error) wsError = e.message
      })
    const interval = setInterval(() => {
      if (ws?.readyState === WebSocket.OPEN)
        ws.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }))
    }, 10000)
    return () => {
      clearInterval(interval)
      ws?.close()
      typingIndicators.forEach(([, timeoutId]) => clearTimeout(timeoutId))
    }
  })

  $effect(() => {
    if (visibilityState === 'visible') unreadMessageCount = 0
  })

  // Reconnect if there's an error and the page is visible
  const isError = $derived(
    wsError && wsError !== 'You are not authenticated to access this resource!',
  ) // We don't care if the error message changed for this $effect, and don't reconnect if not authed.
  $effect(() => {
    if (isError && visibilityState === 'visible') {
      let timeout = -1
      const reconnect = async () => {
        try {
          ws = await connect(id, clientId, { onMessage, onClose }, true)
          wsError = null
        } catch (e: unknown) {
          if (e instanceof Error) wsError = e.message
          timeout = setTimeout(reconnect, 10000)
        }
      }
      // TODO: Implement exponential backoff
      if (wsInitialConnect) timeout = setTimeout(reconnect, 10000)
      else reconnect() // eslint-disable-line @typescript-eslint/no-floating-promises
      return () => clearTimeout(timeout)
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
    <RoomLanding bind:transientVideo error={wsError} connecting={wsInitialConnect} />
  {:else if roomInfo.type === RoomType.LocalFile || roomInfo.type === RoomType.RemoteFile}
    {#key roomInfo.target}
      <FilePlayer
        bind:transientVideo
        {roomInfo}
        {playerState}
        bind:subtitles
        {onPlayerStateChange}
        error={wsError}
        fullscreenEl={containerEl}
      />
    {/key}
  {:else}
    <RoomLanding bind:transientVideo error="Invalid room type!" connecting={false} />
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
