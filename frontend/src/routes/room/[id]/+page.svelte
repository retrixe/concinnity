<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import Chat from '$lib/components/room/Chat.svelte'
  import RoomLanding from '$lib/components/room/RoomLanding.svelte'
  import LocalFilePlayer from '$lib/components/room/LocalFilePlayer.svelte'
  import {
    connect,
    isIncomingChatMessage,
    MessageType,
    RoomType,
    type ChatMessage,
    type GenericMessage,
  } from '$lib/api/room'

  const id = page.params.id
  const messages: ChatMessage[] = $state([])

  // FIXME: Implement room info/player state handling
  // - Implement a way to select a video when no video is playing
  // - Implement a way to select a video when a video is already requested to play in room info
  // - Autoplay may not work on browsers, so a manual play button may be needed
  // FIXME: Implement video controls in chat via commands
  // FIXME: Implement no file and local file watching
  // TODO: Support watching remote files
  // TODO: Implement UI controls

  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null)
  const wsConnecting = $derived(!wsError && ws === null)

  let roomType: RoomType = $state(RoomType.None)

  onMount(() => {
    connect(id, {
      onMessage: message => {
        try {
          if (typeof message.data !== 'string') throw new Error('Invalid message data type!')
          const data = JSON.parse(message.data) as GenericMessage
          if (isIncomingChatMessage(data)) {
            messages.push(...data.data)
          } else if (data.type !== MessageType.Pong) {
            console.warn('Unhandled message type!', data)
          }
        } catch (e) {
          console.error('Failed to parse backend message!', message, e)
        }
      },
      onClose: event => {
        wsError = event.reason || `WebSocket closed with code: ${event.code}`
      },
      onError: () => {
        /* no-op */
      },
    })
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
</script>

<div class="container">
  {#if roomType === RoomType.None}
    <RoomLanding error={wsError} connecting={wsConnecting} />
  {:else if roomType === RoomType.LocalFile}
    <LocalFilePlayer error={wsError} />
  {:else}
    <RoomLanding error="Invalid room type!" connecting={false} />
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
    max-height: calc(100vh - 4rem); // FIXME: Chat overflows on mobile...
    flex: 1;
    display: flex;
    flex-direction: column;
    @media screen and (min-width: 768px) {
      flex-direction: row;
    }
  }
</style>
