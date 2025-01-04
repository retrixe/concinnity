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

  const id = page.params.id
  const messages: ChatMessage[] = $state([])

  // TODO: Support watching remote files

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
        messages.push(...message.data)
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

  // Reconnect if there's an error
  $effect(() => {
    if (wsError) {
      // TODO: Food for thought - What if you reconnect in a time period longer than 10s?
      // TODO: Replace previous chats with missing messages.
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

<div class="container">
  {#if !roomInfo || roomInfo.type === RoomType.None}
    <RoomLanding bind:transientVideo error={wsError} connecting={wsInitialConnect} />
  {:else if roomInfo.type === RoomType.LocalFile}
    {#key roomInfo.target}
      <LocalFilePlayer bind:transientVideo {roomInfo} {playerState} error={wsError} />
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
    max-height: calc(100vh - 4rem);
    flex: 1;
    display: flex;
    flex-direction: column;
    @media screen and (min-width: 768px) {
      flex-direction: row;
    }
  }
</style>
