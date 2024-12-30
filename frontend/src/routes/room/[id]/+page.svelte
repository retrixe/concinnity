<script lang="ts">
  import Button from '$lib/components/Button.svelte'
  import { onMount } from 'svelte'
  import Chat from './Chat.svelte'
  import { page } from '$app/state'
  import { connect } from './gateway'

  const mockMessages = [
    { userId: 'sanguineous', message: 'Hello', timestamp: '2018-11-05T00:54:15.000005125Z' },
    { userId: 'aelia', message: 'Hi :3', timestamp: '2024-12-30T04:43:53.156212954+05:30' },
  ]

  // FIXME: Implement chat
  // FIXME: Implement room info/player state handling
  // - Implement a way to select a video when no video is playing
  // - Implement a way to select a video when a video is already requested to play in room info
  // - Autoplay may not work on browsers, so a manual play button may be needed
  // FIXME: Implement video controls in chat via commands
  // TODO: Support watching remote files, local files, or no file
  // TODO: Implement UI controls

  // FIXME: If error, full-screen error or warning at the bottom of the video
  // FIXME: If connecting for the first time, full-screen loading
  // FIXME: Disable chat box if disconnected
  let ws: WebSocket | null = $state(null)
  let wsError: string | null = $state(null) // eslint-disable-line -- FIXME
  const id = page.params.id

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
  <div class="video">
    <div class="video-select">
      <h1>No video playing</h1>
      <br />
      <Button>Select local file</Button>
    </div>
  </div>
  <Chat
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

  .video {
    background-color: #000000;
    min-height: 280px; // FIXME: Remove if there's a video?
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    color: white;
    @media screen and (min-width: 768px) {
      flex: 1;
    }
  }

  .video-select {
    display: flex;
    flex-direction: column;
    align-items: center;
  }
</style>
