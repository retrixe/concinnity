<script lang="ts">
  import type { PlayerState } from '$lib/api/room'
  import { fade } from 'svelte/transition'
  import Play from 'phosphor-svelte/lib/Play'

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
    controlsVisible = true
  }}
>
  <!-- svelte-ignore a11y_media_has_caption -->
  <video class="video" {src} playsinline></video>
  {#if controlsVisible}
    <div class="controls" transition:fade>
      <button>
        <Play weight="bold" size="16px" />
      </button>
      <button>Rewind 10s</button>
      <button>Forward 10s</button>
    </div>
  {/if}
</div>

<style lang="scss">
  .player-container {
    max-width: 100%;
    max-height: 100%;
    position: relative;
  }

  .video {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .controls {
    background-color: rgba(0, 0, 0, 0.5);
    width: 100%;
    position: absolute;
    bottom: 0;
    display: flex;
    align-items: center;
  }

  // TODO: DRY with Button.svelte
  button {
    margin: 8px;
    padding: 8px;
    color: white;
    background-color: transparent;
    border: none;
    border-radius: 0.5rem;
    transition:
      background-color 0.2s ease-in-out,
      filter 0.2s ease-in-out;
    &:enabled {
      &:hover {
        background-color: var(--primary-color);
      }
      &:active {
        filter: brightness(0.8);
      }
    }
    &:disabled {
      background-color: var(--divider-color);
      cursor: not-allowed;
    }
    :global(svg) {
      display: block;
    }
  }
</style>
