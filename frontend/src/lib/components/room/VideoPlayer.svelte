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
  import Button from '../Button.svelte'

  interface Props {
    video: File
    name: string
    playerState: PlayerState
    onPlayerStateChange: (newState: PlayerState) => void
    subtitles: Record<string, string | null>
    fullscreenEl: Element
    onStop: () => void
  }
  const {
    video,
    name,
    playerState,
    onPlayerStateChange,
    subtitles,
    fullscreenEl,
    onStop: handleStop,
  }: Props = $props()
  const src = $derived(URL.createObjectURL(video))

  let controlsVisible = $state(false)
  let lastLocalAction: Date | null = null

  let videoEl = $state(null) as HTMLVideoElement | null
  let paused = $state(true)
  let currentTime = $state(0)
  let duration = $state(0)
  let displayCurrentTime = $state(false)
  let muted = $state(false)
  let volume = $state(1)
  let playbackRate = $state(1)
  let subtitle = $state<null | string>(null)
  let fullscreenElement = $state(null) as Element | null
  let settingsMenu = $state<null | 'options' | 'speed' | 'subtitles'>(null)
  let autoplayNotif = $state(false)

  // Synchronise to incoming player state changes
  const synchroniseToPlayerState = () => {
    const lastAction = new Date(playerState.lastAction).getTime()
    if (lastLocalAction && lastLocalAction.getTime() > lastAction) return // Don't override local changes
    const currentTimeDelta = playerState.paused ? 0 : Math.max((Date.now() - lastAction) / 1000, 0)
    currentTime = playerState.timestamp + currentTimeDelta
    playbackRate = playerState.speed
    if (playerState.paused) {
      paused = true
    } else {
      const promise = videoEl?.play()
      promise
        ?.then(() => {
          autoplayNotif = false
        })
        .catch(() => {
          autoplayNotif = true
        })
    }
  }
  $effect(synchroniseToPlayerState)

  // Send player state changes on pause or speed change
  // TODO: This doesn't interact with extensions like Video Speed Controller
  const handlePlayerStateChange = () => {
    lastLocalAction = new Date()
    const time = lastLocalAction.toISOString()
    onPlayerStateChange({ paused, speed: playbackRate, timestamp: currentTime, lastAction: time })
  }

  const handlePlayPause = () => {
    paused = !paused
    handlePlayerStateChange()
  }

  const handleTimeScrub = (e: Event) => {
    if (e instanceof KeyboardEvent && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
      e.preventDefault()
      currentTime += e.key === 'ArrowLeft' ? -5 : 5
      handlePlayerStateChange()
    } else if (!(e instanceof KeyboardEvent)) {
      currentTime = Number((e.target as HTMLInputElement).value)
      handlePlayerStateChange()
    }
  }

  const handleDurationToggle = (e: KeyboardEvent | MouseEvent) => {
    if (e instanceof KeyboardEvent && e.key !== 'Enter' && e.key !== ' ') return
    displayCurrentTime = !displayCurrentTime
  }

  const handleMuteToggle = () => (muted = !muted)

  const handleVolumeScrub = (e: KeyboardEvent) => {
    if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
      e.preventDefault()
      volume = e.key === 'ArrowLeft' ? Math.max(0, volume - 0.1) : Math.min(1, volume + 0.1)
      handlePlayerStateChange()
    }
  }

  const handleSettingsOpen = () => (settingsMenu = settingsMenu === null ? 'options' : null)

  const handleSettingsNav = (menu: typeof settingsMenu) => () => (settingsMenu = menu)

  const handlePlayRateChange = (rate: number) => () => {
    playbackRate = rate
    handlePlayerStateChange()
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
</script>

<svelte:document bind:fullscreenElement />
<svelte:window onclickcapture={handleWindowClick} />
<div
  role="presentation"
  class="player-container"
  onmouseenter={() => (controlsVisible = true)}
  onmouseleave={() => (controlsVisible = false)}
>
  {#if autoplayNotif}
    <div role="presentation" class="autoplay" onclick={synchroniseToPlayerState}>
      <h1>Autoplay was blocked.<br />Press to begin playing.</h1>
    </div>
  {/if}
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
    <div class="controls top" transition:fade>
      <span>{name}</span>
    </div>
    <div class="controls bottom" transition:fade>
      <Button onclick={handlePlayPause}>
        {#if paused}
          <Play weight="bold" size="16px" />
        {:else}
          <Pause weight="bold" size="16px" />
        {/if}
      </Button>
      <input
        type="range"
        min="0"
        max={isFinite(duration) ? duration : 0}
        step="0.01"
        value={currentTime}
        oninput={handleTimeScrub}
        onkeydown={handleTimeScrub}
        style:flex="1"
      />
      <!-- TODO: The constantly changing width of this thing bugs me -->
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
      <Button onclick={handleMuteToggle}>
        {#if muted}
          <SpeakerX weight="bold" size="16px" />
        {:else if volume < 0.5}
          <SpeakerLow weight="bold" size="16px" />
        {:else}
          <SpeakerHigh weight="bold" size="16px" />
        {/if}
      </Button>
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
        <Button onclick={handleSettingsOpen}>
          <Gear weight="bold" size="16px" />
        </Button>
        <div class="settings-menu" style:visibility={settingsMenu ? 'visible' : 'hidden'}>
          {#if settingsMenu == 'speed'}
            <Button onclick={handleSettingsNav('options')} class="highlight">
              <CaretLeft weight="bold" size="16px" /> Back to options
            </Button>
            {#each [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2, 4] as rate (rate)}
              <Button
                class={playbackRate === rate ? 'highlight' : ''}
                onclick={handlePlayRateChange(rate)}
              >
                {rate}x
              </Button>
            {/each}
          {:else if settingsMenu == 'subtitles'}
            <Button onclick={handleSettingsNav('options')} class="highlight">
              <CaretLeft weight="bold" size="16px" /> Back to options
            </Button>
            {#each Object.keys(subtitles) as sub}
              <Button onclick={() => (subtitle = sub)} class={subtitle === sub ? 'highlight' : ''}>
                <span>{sub}</span>
              </Button>
            {/each}
            <Button onclick={() => (subtitle = null)} class={subtitle === null ? 'highlight' : ''}>
              <span>None</span>
            </Button>
          {:else}
            <Button onclick={synchroniseToPlayerState}>
              <span>Sync to others</span>
            </Button>
            <Button onclick={handleSettingsNav('speed')}>
              <span>Speed</span>
              <span>{playbackRate}x</span>
            </Button>
            <Button onclick={handleSettingsNav('subtitles')}>
              <span>Subtitles</span>
              <span>{subtitle ?? 'None'}</span>
            </Button>
          {/if}
        </div>
      </div>
      <Button onclick={handleStop}>
        <Stop weight="bold" size="16px" />
      </Button>
      <Button onclick={handlePiPToggle}>
        <PictureInPicture weight="bold" size="16px" />
      </Button>
      <Button onclick={handleFullScreenToggle}>
        {#if fullscreenElement === fullscreenEl}
          <ArrowsIn weight="bold" size="16px" />
        {:else}
          <ArrowsOut weight="bold" size="16px" />
        {/if}
      </Button>
    </div>
  {/if}
</div>

<style lang="scss">
  .player-container {
    max-width: 100%;
    max-height: 100%;
    position: relative;
  }

  .autoplay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    user-select: none;
    z-index: 2;
    background-color: rgba(0, 0, 0, 0.5);
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
    display: flex;
    align-items: center;
    > span {
      margin: 8px;
    }
    :global(button) {
      margin: 8px;
      padding: 8px;
      background-color: transparent;
      transition:
        background-color 0.2s ease-in-out,
        filter 0.2s ease-in-out;
      &:enabled {
        &:hover {
          background-color: var(--primary-color);
        }
      }
      :global(svg) {
        display: block;
      }
    }
  }

  .bottom {
    bottom: 0;
  }

  .top {
    top: 0;
    overflow: hidden;
  }

  .settings-menu {
    position: absolute;
    right: calc(100% - 48px);
    bottom: 100%;
    background-color: rgba(0, 0, 0, 0.6);
    min-width: 180px;
    max-width: 50vh;
    max-height: 50vh; // TODO: Might exceed height of the video container
    overflow-x: hidden;
    overflow-y: scroll;
    :global(button) {
      gap: 1rem;
      width: calc(100% - 16px);
      display: flex;
      justify-content: space-between;
      :global(svg) {
        display: inline;
      }
    }
    :global(button.highlight) {
      background-color: var(--primary-color);
    }
  }
</style>
