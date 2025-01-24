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

  // TODO: Support watching remote files
  const id = page.params.id

  let containerEl = $state(null) as Element | null
  let visibilityState = $state('visible') as DocumentVisibilityState
  let messages: ChatMessage[] = $state([])
  let playerState = $state(initialPlayerState)
  let roomInfo: RoomInfo | null = $state(null)
  let subtitles: Record<string, string | null> = $state({})
  let transientVideo: File | null = $state(null)
  let typingStates: Record<string, boolean>[] = $state([])

  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null)
  const wsInitialConnect = $derived((ws === null && !wsError) || roomInfo === null)
  const { username } = $derived(page.data)

  const onMessage = (event: MessageEvent) => {
    try {
      if (typeof event.data !== 'string') throw new Error('Invalid message data type!')
      const message = JSON.parse(event.data) as GenericMessage
      if (isIncomingChatMessage(message)) {
        // TODO (low): Replace messages.length with IDs
        if (message.data.length === 1) messages.push(message.data[0])
        else messages.push(...message.data.slice(messages.length))
      } else if (isIncomingSubtitleMessage(message)) {
        message.data.forEach(name => {
          subtitles[name] = null
        })
      } else if (isIncomingRoomInfoMessage(message)) {
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
        const incomingKey = Object.keys(message.data)[0]
        const existingIndex = typingStates.findIndex(item => {
          const key = Object.keys(item)[0]
          return key === incomingKey
        })
        if (existingIndex !== -1) {
          typingStates[existingIndex] = message.data
        } else {
          typingStates.push(message.data)
        }
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
    onSendMessage={(message: string) => {
      ws?.send(JSON.stringify({ type: 'chat', data: message }))
    }}
    onTyping={(username: string, isTyping: boolean) => {
      ws?.send(JSON.stringify({ type: 'typing', data: { [username]: isTyping } }))
    }}
    {username}
    {typingStates}
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
