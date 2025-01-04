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
    MessageType,
    RoomType,
    type ChatMessage,
    type GenericMessage,
    type RoomInfo,
  } from '$lib/api/room'

  // TODO: Support watching remote files
  const id = page.params.id

  let containerEl = $state(null) as Element | null
  let visibilityState = $state('visible') as DocumentVisibilityState
  let messages: ChatMessage[] = $state([])
  let playerState = $state(initialPlayerState)
  let roomInfo: RoomInfo | null = $state(null)
  let transientVideo: File | null = $state(null)

  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null)
  const wsInitialConnect = $derived((ws === null && !wsError) || roomInfo === null)

  const onMessage = (event: MessageEvent) => {
    try {
      if (typeof event.data !== 'string') throw new Error('Invalid message data type!')
      const message = JSON.parse(event.data) as GenericMessage
      if (isIncomingChatMessage(message)) {
        if (message.data.length === 1) messages.push(message.data[0])
        else messages = message.data // TODO (low): Have IDs on messages to just push new messages
      } else if (isIncomingRoomInfoMessage(message)) {
        if (roomInfo === null) {
          roomInfo = message.data
          playerState = initialPlayerState
        } else {
          Object.assign(roomInfo, message.data)
        }
      } else if (isIncomingPlayerStateMessage(message)) {
        playerState = message.data
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
