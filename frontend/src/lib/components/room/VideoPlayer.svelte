<script lang="ts">
  import { fade } from 'svelte/transition'
  import {
    ArrowsIn,
    ArrowsOut,
    CaretLeft,
    Gear,
    Pause,
    PictureInPicture,
    Play,
    SpeakerHigh,
    SpeakerLow,
    SpeakerX,
    Stop,
  } from 'phosphor-svelte'
  import type { PlayerState } from '$lib/api/room'
  import { stringifyDuration } from '$lib/utils/duration'

  interface Props {
    video: File
    playerState: PlayerState
    fullscreenEl: Element
    onStop: () => void
  }
  const { video, playerState, fullscreenEl, onStop: handleStop }: Props = $props()
  const src = $derived(URL.createObjectURL(video))

  let controlsVisible = $state(false)

  let videoEl = $state(null) as HTMLVideoElement | null
  let paused = $state(true)
  let currentTime = $state(0)
  let duration = $state(0)
  let displayCurrentTime = $state(false)
  let muted = $state(false)
  let volume = $state(1)
  let playbackRate = $state(1)
  let fullscreenElement = $state(null) as Element | null
  let settingsMenu = $state<null | 'options' | 'speed'>(null)

  $inspect(playerState) // FIXME: Implement syncing with playerState

  const handlePlayPause = () => {
    paused = !paused
  }

  const handleTimeScrub = (e: KeyboardEvent) => {
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      e.preventDefault()
      currentTime += e.key === 'ArrowLeft' ? -5 : 5
    }
  }

  const handleDurationToggle = (e: KeyboardEvent | MouseEvent) => {
    if (e instanceof KeyboardEvent && e.key !== 'Enter' && e.key !== ' ') return
    displayCurrentTime = !displayCurrentTime
  }

  const handleMuteToggle = () => {
    muted = !muted
  }

  const handleVolumeScrub = (e: KeyboardEvent) => {
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      e.preventDefault()
      volume = e.key === 'ArrowLeft' ? Math.max(0, volume - 0.1) : Math.min(1, volume + 0.1)
    }
  }

  const handleSettingsOpen = () => {
    if (settingsMenu === null) settingsMenu = 'options'
    else settingsMenu = null
  }

  const handleSettingsNav = (menu: typeof settingsMenu) => () => {
    settingsMenu = menu
  }

  const handlePlayRateChange = (rate: number) => () => {
    playbackRate = rate
  }

  const handlePiPToggle = () => {
    // TODO: Implement the document picture-in-picture API
    // https://developer.chrome.com/docs/web-platform/document-picture-in-picture
    if (document.pictureInPictureElement === videoEl && videoEl) {
      document.exitPictureInPicture().catch(console.error)
    } else {
      videoEl?.requestPictureInPicture().catch(console.error)
    }
  }

  const handleFullScreenToggle = () => {
    if (fullscreenElement === fullscreenEl) {
      document.exitFullscreen().catch(console.error)
    } else {
      fullscreenEl.requestFullscreen().catch(console.error)
    }
  }

  const handleWindowClick = (event: MouseEvent) => {
    const outsideSettingsMenuBounds =
      event.target instanceof Element && !event.target.closest('.settings-menu')
    if (settingsMenu && outsideSettingsMenuBounds) settingsMenu = null
  }

  // Video controls:
  // - Play/Pause
  // - Seek timeline
  // - Time elapsed/time left (on tap)
  // - Volume control (mute bottom + range)
  // - Settings button (menu with speed control)
  // - Stop playing current video
  // - Picture-in-picture button
  // - Fullscreen button
  // TODO: Implement tooltips
  // TODO: Implement kb controls for all when cursor in bounds (except stop, to avoid accidents)
  // FIXME: Autoplay may not work on browsers, so a manual play button may be needed at first
</script>

<svelte:document bind:fullscreenElement />
<svelte:window onclickcapture={handleWindowClick} />
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
    bind:playbackRate
    playsinline
  ></video>
  <!-- TODO: Width of transiently passed videos are incorrect sometimes -->
  <!-- TODO: Controls are too wide on mobile in portrait -->
  {#if controlsVisible || settingsMenu}
    <div class="controls" transition:fade>
      <button onclick={handlePlayPause}>
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
        onkeydown={handleTimeScrub}
        style:flex="1"
      />
      <span
        role="button"
        tabindex="0"
        onkeypress={handleDurationToggle}
        onclick={handleDurationToggle}
      >
        {displayCurrentTime
          ? '-' + stringifyDuration(duration - currentTime)
          : stringifyDuration(currentTime)}
      </span>
      <button onclick={handleMuteToggle}>
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
        bind:value={volume}
        onkeydown={handleVolumeScrub}
        disabled={muted}
        style:width="80px"
      />
      <div style:position="relative">
        <button onclick={handleSettingsOpen}>
          <Gear weight="bold" size="16px" />
        </button>
        <div class="settings-menu" style:visibility={settingsMenu ? 'visible' : 'hidden'}>
          {#if settingsMenu == 'speed'}
            <button onclick={handleSettingsNav('options')} class="highlight">
              <CaretLeft weight="bold" size="16px" /> Back to options
            </button>
            {#each [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2, 4] as rate (rate)}
              <button class:highlight={playbackRate === rate} onclick={handlePlayRateChange(rate)}>
                {rate}x
              </button>
            {/each}
          {:else}
            <button onclick={handleSettingsNav('speed')}>
              <span>Speed</span>
              <span>{playbackRate}x</span>
            </button>
          {/if}
        </div>
      </div>
      <button onclick={handleStop}>
        <Stop weight="bold" size="16px" />
      </button>
      <button onclick={handlePiPToggle}>
        <PictureInPicture weight="bold" size="16px" />
      </button>
      <button onclick={handleFullScreenToggle}>
        {#if fullscreenElement === fullscreenEl}
          <ArrowsIn weight="bold" size="16px" />
        {:else}
          <ArrowsOut weight="bold" size="16px" />
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
    > span {
      margin: 8px;
    }
  }

  .settings-menu {
    position: absolute;
    bottom: 100%;
    background-color: rgba(0, 0, 0, 0.6);
    min-width: 160px;
    max-height: 50vh;
    overflow-y: scroll;
    button {
      width: calc(100% - 16px);
      display: flex;
      align-items: center;
      justify-content: space-between;
      &.highlight {
        background-color: var(--primary-color);
      }
      :global(svg) {
        display: inline;
      }
    }
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
