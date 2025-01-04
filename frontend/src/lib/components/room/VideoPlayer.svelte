<script lang="ts">
  import type { PlayerState } from '$lib/api/room'
  import { fade } from 'svelte/transition'
  import {
    ArrowsIn,
    ArrowsOut,
    Gear,
    Pause,
    PictureInPicture,
    Play,
    SpeakerHigh,
    SpeakerLow,
    SpeakerX,
    Stop,
  } from 'phosphor-svelte'
  import { stringifyDuration } from '$lib/utils/duration'

  const { video, playerState }: { video: File; playerState: PlayerState } = $props()
  const src = $derived(URL.createObjectURL(video))
  const roomEl = document.getElementsByClassName('room')[0] // TODO: Use a ref

  let controlsVisible = $state(false)

  let videoEl = $state(null) as HTMLVideoElement | null
  let paused = $state(true)
  let currentTime = $state(0)
  let duration = $state(0)
  let displayCurrentTime = $state(false)
  let muted = $state(false)
  let volume = $state(1)
  let fullscreenElement = $state(null) as Element | null

  $inspect(playerState) // FIXME: Implement syncing with playerState

  const onPlayPause = () => {
    paused = !paused
  }

  const onDurationToggle = (e: KeyboardEvent | MouseEvent) => {
    if (e instanceof KeyboardEvent && e.key !== 'Enter') return
    displayCurrentTime = !displayCurrentTime
  }

  const onMuteToggle = () => {
    muted = !muted
  }

  const onPiPToggle = () => {
    // TODO: Implement the document picture-in-picture API
    // https://developer.chrome.com/docs/web-platform/document-picture-in-picture
    if (document.pictureInPictureElement === videoEl && videoEl) {
      document.exitPictureInPicture().catch(console.error)
    } else {
      videoEl?.requestPictureInPicture().catch(console.error)
    }
  }

  const onFullScreenToggle = () => {
    if (fullscreenElement === roomEl) {
      document.exitFullscreen().catch(console.error)
    } else {
      roomEl.requestFullscreen().catch(console.error)
    }
  }

  // TODO: Implement tooltips
  // Implement video controls:
  // - Play/Pause
  // - Seek timeline
  // - Time elapsed/time left (on tap)
  // - Volume control (mute bottom + range)
  // - FIXME: Settings button (menu with speed control)
  // - Picture-in-picture button
  // - FIXME: Button to stop playing
  // - Fullscreen button
  // - FIXME: Arrow keys to rewind/forward/mute/unmute
  // FIXME: Autoplay may not work on browsers, so a manual play button may be needed at first
</script>

<svelte:document bind:fullscreenElement />
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
  <video
    class="video"
    {src}
    bind:this={videoEl}
    bind:duration
    bind:currentTime
    bind:paused
    bind:muted
    bind:volume
    playsinline
  ></video>
  <!-- TODO: Controls are too wide on mobile in portrait -->
  {#if controlsVisible}
    <div class="controls" transition:fade>
      <button onclick={onPlayPause}>
        {#if paused}
          <Play weight="bold" size="16px" />
        {:else}
          <Pause weight="bold" size="16px" />
        {/if}
      </button>
      <input
        type="range"
        min="0"
        max={isFinite(duration) ? duration : 0}
        step="0.01"
        bind:value={currentTime}
        style:flex="1"
      />
      <span role="button" tabindex="0" onkeypress={onDurationToggle} onclick={onDurationToggle}>
        {displayCurrentTime
          ? '-' + stringifyDuration(duration - currentTime)
          : stringifyDuration(currentTime)}
      </span>
      <button onclick={onMuteToggle}>
        {#if muted}
          <SpeakerX weight="bold" size="16px" />
        {:else if volume < 0.5}
          <SpeakerLow weight="bold" size="16px" />
        {:else}
          <SpeakerHigh weight="bold" size="16px" />
        {/if}
      </button>
      <input
        type="range"
        min="0"
        max="1"
        step="0.01"
        style:width="80px"
        disabled={muted}
        bind:value={volume}
      />
      <button>
        <Gear weight="bold" size="16px" />
      </button>
      <button>
        <Stop weight="bold" size="16px" />
      </button>
      <button onclick={onPiPToggle}>
        <PictureInPicture weight="bold" size="16px" />
      </button>
      <button onclick={onFullScreenToggle}>
        {#if fullscreenElement === roomEl}
          <ArrowsOut weight="bold" size="16px" />
        {:else}
          <ArrowsIn weight="bold" size="16px" />
        {/if}
      </button>
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

  span {
    margin: 8px;
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
