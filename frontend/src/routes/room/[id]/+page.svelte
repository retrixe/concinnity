<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/state'
  import Chat from '$lib/components/room/Chat.svelte'
  import RoomLanding from '$lib/components/room/RoomLanding.svelte'
  import LocalFilePlayer from '$lib/components/room/LocalFilePlayer.svelte'
  import { connect, RoomType } from '$lib/api/room'

  const id = page.params.id
  const mockMessages = [
    {
      userId: '20e08df5-948a-4f5d-b8b4-aae20c0ff54b',
      message: 'Hello',
      timestamp: '2018-11-05T00:54:15.000005125Z',
    },
    {
      userId: 'e96626e7-0513-470a-af74-a47aed8ca7a8',
      message: 'Hi :3',
      timestamp: '2024-12-30T04:43:53.156212954+05:30',
    },
  ]

  // FIXME: Implement chat
  // FIXME: Implement room info/player state handling
  // - Implement a way to select a video when no video is playing
  // - Implement a way to select a video when a video is already requested to play in room info
  // - Autoplay may not work on browsers, so a manual play button may be needed
  // FIXME: Implement video controls in chat via commands
  // FIXME: Implement no file and local file watching
  // TODO: Support watching remote files
  // TODO: Implement UI controls

  // FIXME: If error, warning at the bottom of the video
  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null)
  let roomType: RoomType = $state(RoomType.None)

  let initialError = $derived(wsError && ws === null ? wsError : null)
  let initialLoading = $derived(!wsError && ws === null)

  onMount(() => {
    connect(id, {
      // FIXME
      onMessage: message => {
        console.log(message)
      },
      onClose: () => {
        console.log('Disconnected')
      },
      onError: e => {
        console.error(e)
      },
    })
      .then(socket => {
        ws = socket
      })
      .catch((e: unknown) => {
        if (e instanceof Error) wsError = e.message
      })
    return () => ws?.close()
  })
</script>

<div class="container">
  {#if roomType === RoomType.None}
    <RoomLanding error={initialError} loading={initialLoading} />
  {:else if roomType === RoomType.LocalFile}
    <LocalFilePlayer error={wsError} />
  {:else}
    <RoomLanding error="Invalid room type!" loading={false} />
  {/if}
  <Chat
    disabled={wsError !== null || ws === null}
    messages={mockMessages}
    onSendMessage={message => {
      console.log(message)
    }}
  />
</div>

<style lang="scss">
  .container {
    flex: 1;
    display: flex;
    flex-direction: column;
    @media screen and (min-width: 768px) {
      flex-direction: row;
    }
  }
</style>
