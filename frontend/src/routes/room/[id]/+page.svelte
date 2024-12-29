<script>
  import Button from '$lib/components/Button.svelte'
  // FIXME: Implement connectivity to WebSocket
  // - onMount, connect to WebSocket
  // - if error, full-screen error
  // - if not connected yet, full-screen loading
  // - any disconnects will result in the chat box being disabled (also during load)
  //   and warning at the bottom of the video until reconnected
  // FIXME: Implement chat
  // FIXME: Implement room info/player state handling
  // - Autoplay may not work on browsers, so a manual play button may be needed
  // - Implement a way to select a video when a video is already playing
  // FIXME: Implement how to select a new video and video controls in chat
  // TODO: Support clearing the room and watching remote files
  // TODO: Implement UI controls
</script>

<div class="container">
  <div class="video">
    <div class="video-select">
      <h1>No video playing</h1>
      <br />
      <Button>Select local file</Button>
    </div>
  </div>
  <div class="chat"></div>
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

  .chat {
    @media screen and (width < 768px) {
      flex: 1;
    }
    @media screen and (min-width: 768px) {
      width: 280px;
    }
  }
</style>
