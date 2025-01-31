<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import Chat from '$lib/components/room/Chat.svelte'
  import RoomLanding from '$lib/components/room/RoomLanding.svelte'
  import LocalFilePlayer from '$lib/components/room/LocalFilePlayer.svelte'
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
  import { SvelteMap } from 'svelte/reactivity'

  // TODO: Support watching remote files
  const id = page.params.id

  let containerEl = $state(null) as Element | null
  let visibilityState = $state('visible') as DocumentVisibilityState
  let messages: ChatMessage[] = $state([])
  let playerState = $state(initialPlayerState)
  let roomInfo: RoomInfo | null = $state(null)
  let subtitles: Record<string, string | null> = $state({})
  let transientVideo: File | null = $state(null)
  let typingIndicators = new SvelteMap<string, [number, number]>()

  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null)
  const wsInitialConnect = $derived((ws === null && !wsError) || roomInfo === null)

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
        for (const message of newMessages) {
          const typingIndicator = typingIndicators.get(message.userId)
          if (typingIndicator) {
            clearTimeout(typingIndicator[1])
            typingIndicators.delete(message.userId)
          }
        }
        messages.push(...newMessages)
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
    connect(id, { onMessage, onClose })
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

  // Reconnect if there's an error and the page is visible
  // TODO: Get rid of the 10s delay on first reconnect after page is visible
  $effect(() => {
    if (wsError && visibilityState === 'visible') {
      const interval = setInterval(async () => {
        try {
          ws = await connect(id, { onMessage, onClose }, true)
          wsError = null
        } catch (e: unknown) {
          if (e instanceof Error) wsError = e.message
        }
      }, 10000)
      return () => clearInterval(interval)
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
<div class="container room" bind:this={containerEl}>
  {#if !roomInfo || roomInfo.type === RoomType.None}
    <RoomLanding bind:transientVideo error={wsError} connecting={wsInitialConnect} />
  {:else if roomInfo.type === RoomType.LocalFile}
    {#key roomInfo.target}
      <LocalFilePlayer
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
