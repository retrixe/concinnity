<script lang="ts">
  // TODO: Width of transiently passed videos are incorrect sometimes
  import { fade } from 'svelte/transition'
  import { Button, Tooltip } from 'heliodor'
  import {
    ArrowsIn,
    ArrowsOut,
    CaretLeft,
    Gear,
    Pause,
    PictureInPicture,
    Play,
    Plus,
    SpeakerHigh,
    SpeakerLow,
    SpeakerX,
    Stop,
    Subtitles,
    SubtitlesSlash,
  } from 'phosphor-svelte'
  import ky from '$lib/api/ky'
  import type { PlayerState } from '$lib/api/room'
  import { stringifyDuration } from '$lib/utils/duration'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import { srt2webvtt } from '$lib/utils/srt'
  import { page } from '$app/state'

  interface Props {
    src: string
    name: string
    playerState: PlayerState
    onPlayerStateChange: (newState: PlayerState) => void
    subtitles: Record<string, string | null>
    fullscreenEl: Element
    onStop: () => void
  }
  const id = page.params.id
  const {
    src,
    name,
    playerState,
    onPlayerStateChange,
    subtitles = $bindable(), // TODO: Remove the bindables and rework how this data flow works...
    fullscreenEl,
    onStop: handleStop,
  }: Props = $props()

  let videoEl = $state(null) as HTMLVideoElement | null
  let controlsVisible = $state(false)
  let displayCurrentTime = $state(false)
  let settingsMenu = $state<null | 'options' | 'speed' | 'subtitles'>(null)
  let fullscreenElement = $state(null) as Element | null
  let autoplayNotif = $state(false)
  let lastLocalAction: Date | null = null

  let paused = $state(true)
  let currentTime = $state(0)
  let duration = $state(0)
  let muted = $state(false)
  let volume = $state(1)
  let playbackRate = $state(1)
  let subtitle = $state<null | [boolean, string]>(null)

  // Synchronise to incoming player state changes
  const synchroniseToPlayerState = () => {
    const lastAction = new Date(playerState.lastAction).getTime()
    // API quirk: If the last action locally took place after the last action from the server, ignore the server action
    if (lastLocalAction && lastLocalAction.getTime() > lastAction) return
    const latencyDelta = playerState.paused ? 0 : Math.max((Date.now() - lastAction) / 1000, 0)
    currentTime = playerState.timestamp + latencyDelta
    playbackRate = playerState.speed
    if (playerState.paused) {
      paused = true
      autoplayNotif = false
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
  // TODO: This doesn't interact with e.g. extensions like Video Speed Controller or closing PiP
  // We could update the player state on ratechange/pause/play/durationchange in the video?
  // (Though we'll need to ignore any of our own changes when these events are fired...)
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
    const increase = e.key === 'ArrowUp' || e.key === 'ArrowRight'
    const decrease = e.key === 'ArrowDown' || e.key === 'ArrowLeft'
    if (increase || decrease) {
      e.preventDefault()
      volume = decrease ? Math.max(0, volume - 0.1) : Math.min(1, volume + 0.1)
    }
  }

  const handleSettingsOpen = () => (settingsMenu = settingsMenu === null ? 'options' : null)

  const handleSettingsNav = (menu: typeof settingsMenu) => () => (settingsMenu = menu)

  const handlePlayRateChange = (rate: number) => () => {
    playbackRate = rate
    handlePlayerStateChange()
  }

  const handleSubtitleSelect = (name: string) => () => (subtitle = [true, name])

  const handleSubtitleToggle = () => {
    subtitle = subtitle ? [!subtitle[0], subtitle[1]] : [true, Object.keys(subtitles)[0]]
  }

  // If there weren't subtitles before, but there are now, then use the first subtitle in the list.
  let subtitleCount = 0
  $effect(() => {
    if (!subtitle && !subtitleCount && Object.keys(subtitles).length) {
      subtitle = [true, Object.keys(subtitles)[0]]
    }
    subtitleCount = Object.keys(subtitles).length
  })

  // In case the subtitles are replaced at the parent component, we use $effect here
  $effect(() => {
    if (subtitle?.[0] && !subtitles[subtitle[1]]) {
      const name = subtitle[1]
      subtitles[name] = ''
      ky(`api/room/${id}/subtitle?name=${encodeURIComponent(name)}`)
        .text()
        .then(text => (subtitles[name] = text))
        .catch((e: unknown) => {
          alert('Failed to retrieve subtitles!')
          console.error('Failed to retrieve subtitles!', e)
          subtitles[name] = null
        })
    }
  })

  const subtitleUrl = $derived.by(() => {
    if (!subtitle?.[0]) return null
    const subs = subtitles[subtitle[1]]
    return subs ? URL.createObjectURL(new Blob([srt2webvtt(subs)], { type: 'text/plain' })) : null
  })

  const handleSubtitleUpload = async () => {
    const file = await openFileOrFiles({
      types: [
        {
          description: 'Subtitles',
          accept: { 'text/vtt': ['.vtt'], 'application/x-subrip': ['.srt'] },
        },
      ],
    })
    if (!file) return
    if (file.size > 1024 * 1024) return alert('Subtitles must be less than 1MB!')
    const filename = encodeURIComponent(file.name)
    try {
      await ky.post(`api/room/${id}/subtitle?name=${filename}`, { body: file })
    } catch (e: unknown) {
      alert('Failed to upload subtitle!')
      console.error('Failed to upload subtitle!', e)
    }
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

  $effect(() => {
    try {
      // Request video to automatically enter picture-in-picture when eligible.
      // @ts-expect-error -- enterpictureinpicture is not yet standardised
      navigator.mediaSession.setActionHandler('enterpictureinpicture', () => {
        videoEl?.requestPictureInPicture().catch(console.error)
      })
      // @ts-expect-error -- enterpictureinpicture is not yet standardised
      return () => navigator.mediaSession.setActionHandler('enterpictureinpicture', null)
    } catch {
      /* Ignore any errors */
    }
  })

  const handleFullScreenToggle = () => {
    if (fullscreenElement === fullscreenEl) {
      document.exitFullscreen().catch(console.error)
    } else {
      fullscreenEl.requestFullscreen().catch(console.error)
    }
  }

  const handleWindowClick = (event: MouseEvent) => {
    const outsideSettingsMenuBounds =
      event.target instanceof Element &&
      !event.target.closest('.settings-menu') &&
      !event.target.closest('.settings-open-btn') // Exclude the settings button itself
    if (settingsMenu && outsideSettingsMenuBounds) settingsMenu = null
  }

  const handleKeyboardControls = (event: KeyboardEvent) => {
    if (event.target !== document.body && event.target !== videoEl) return
    if (event.key === ' ') handlePlayPause()
    if (event.key === 'm') handleMuteToggle()
    if (event.key === 'f') handleFullScreenToggle()
    if (event.key === 'p') handlePiPToggle()
    if (event.key === 'ArrowUp' || event.key === 'ArrowDown') handleVolumeScrub(event)
    if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') handleTimeScrub(event)
  }
</script>

<svelte:document bind:fullscreenElement />
<svelte:window onclickcapture={handleWindowClick} onkeydown={handleKeyboardControls} />
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
  >
    {#if subtitleUrl}
      <track kind="subtitles" src={subtitleUrl} label={subtitle?.[1] ?? 'N/A'} default />
    {/if}
  </video>
  {#if controlsVisible || settingsMenu}
    <div class="controls top" transition:fade>
      <span>{name}</span>
    </div>
    <div class="controls bottom" transition:fade>
      <Tooltip text={paused ? 'Play' : 'Pause'}>
        <Button onclick={handlePlayPause}>
          {#if paused}
            <Play weight="bold" size="1rem" />
          {:else}
            <Pause weight="bold" size="1rem" />
          {/if}
        </Button>
      </Tooltip>
      <input
        type="range"
        min="0"
        max={isFinite(duration) ? duration : 0}
        step="0.01"
        value={currentTime}
        oninput={handleTimeScrub}
        onkeydown={handleTimeScrub}
        style:flex="1"
        style:min-width="50px"
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
      <Tooltip text={muted ? 'Unmute' : 'Mute'}>
        <Button class="hide-on-mobile" onclick={handleMuteToggle}>
          {#if muted}
            <SpeakerX weight="bold" size="1rem" />
          {:else if volume < 0.5}
            <SpeakerLow weight="bold" size="1rem" />
          {:else}
            <SpeakerHigh weight="bold" size="1rem" />
          {/if}
        </Button>
      </Tooltip>
      <input
        type="range"
        min="0"
        max="1"
        step="0.01"
        bind:value={volume}
        onkeydown={handleVolumeScrub}
        disabled={muted}
        style:width="80px"
        class="hide-on-mobile"
      />
      <div style:position="relative">
        <Tooltip text="Settings">
          <Button class="settings-open-btn" onclick={handleSettingsOpen}>
            <Gear weight="bold" size="1rem" />
          </Button>
        </Tooltip>
        <div class="settings-menu" style:visibility={settingsMenu ? 'visible' : 'hidden'}>
          {#if settingsMenu == 'speed'}
            <Button onclick={handleSettingsNav('options')} class="highlight">
              <CaretLeft weight="bold" size="1rem" /> Back to options
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
              <CaretLeft weight="bold" size="1rem" /> Back to options
            </Button>
            {#each Object.keys(subtitles) as sub (sub)}
              <Button
                onclick={handleSubtitleSelect(sub)}
                class={subtitle?.[0] && subtitle[1] === sub ? 'highlight' : ''}
              >
                <span>{sub}</span>
              </Button>
            {/each}
            <Button onclick={handleSubtitleUpload}>
              <span>Upload</span>
              <Plus weight="bold" size="1rem" />
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
              <span>{subtitle?.[0] ? subtitle[1] : 'None'}</span>
            </Button>
          {/if}
        </div>
      </div>
      <Tooltip text="Stop playback">
        <Button onclick={handleStop}>
          <Stop weight="bold" size="1rem" />
        </Button>
      </Tooltip>
      {#if Object.keys(subtitles).length}
        <Tooltip text={subtitle?.[0] ? 'Hide subtitles' : 'Show subtitles'}>
          <Button onclick={handleSubtitleToggle}>
            {#if subtitle?.[0]}
              <Subtitles weight="bold" size="1rem" />
            {:else}
              <SubtitlesSlash weight="bold" size="1rem" />
            {/if}
          </Button>
        </Tooltip>
      {/if}
      <Tooltip text="Picture-in-picture">
        <Button onclick={handlePiPToggle}>
          <PictureInPicture weight="bold" size="1rem" />
        </Button>
      </Tooltip>
      <Tooltip text={fullscreenElement === fullscreenEl ? 'Exit fullscreen' : 'Enter fullscreen'}>
        <Button onclick={handleFullScreenToggle}>
          {#if fullscreenElement === fullscreenEl}
            <ArrowsIn weight="bold" size="1rem" />
          {:else}
            <ArrowsOut weight="bold" size="1rem" />
          {/if}
        </Button>
      </Tooltip>
    </div>
  {/if}
</div>

<style lang="scss">
  :global(.hide-on-mobile) {
    @media screen and (max-width: 600px) {
      display: none;
    }
  }

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
    z-index: 100;
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
