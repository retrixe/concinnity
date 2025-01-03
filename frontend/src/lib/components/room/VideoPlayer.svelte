<script lang="ts">
  import type { PlayerState } from '$lib/api/room'
  import { fade } from 'svelte/transition'

  const { video, playerState }: { video: File; playerState: PlayerState } = $props()
  $inspect(playerState) // FIXME: Implement syncing with playerState
  const src = $derived(URL.createObjectURL(video))

  let controlsVisible = $state(false)

  // FIXME: Implement video controls:
  // - Play/Pause
  // - Seek timeline
  // - Time elapsed/time left (on tap)
  // - Settings button (speed control)
  // - Picture-in-picture button
  // - Button to stop playing
  // - Fullscreen button
  // - Arrow keys to rewind/forward/mute/unmute
  // FIXME: Autoplay may not work on browsers, so a manual play button may be needed at first
</script>

<div
  role="presentation"
  class="player-container"
  onmouseenter={() => {
    controlsVisible = true
  }}
  onmouseleave={() => {
    controlsVisible = false
  }}
>
  <!-- svelte-ignore a11y_media_has_caption -->
  <video class="video" {src} playsinline></video>
  {#if controlsVisible}
    <div class="controls" transition:fade>
      <button>Play/Pause</button>
      <button>Rewind 10s</button>
      <button>Forward 10s</button>
    </div>
  {/if}
</div>

<style lang="scss">
  .player-container {
    position: relative;
  }

  .video {
    display: block;
    width: 100%;
    object-fit: contain;
  }

  .controls {
    position: absolute;
    bottom: 0;
  }
</style>
