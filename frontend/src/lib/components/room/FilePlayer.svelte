<script lang="ts">
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import { RoomType, type PlayerState, type RoomInfo } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import { CaretDown } from 'phosphor-svelte'
  import VideoPlayer from './VideoPlayer.svelte'
  import { DropdownButton, Menu, MenuItem } from 'heliodor'

  interface Props {
    error: string | null
    roomInfo: RoomInfo
    playerState: PlayerState
    subtitles: Record<string, string | null>
    onPlayerStateChange: (newState: PlayerState) => void
    transientVideo: File | null
    fullscreenEl: Element
  }

  const id = page.params.id
  let {
    error,
    roomInfo,
    playerState,
    subtitles = $bindable(),
    onPlayerStateChange,
    transientVideo = $bindable(null),
    fullscreenEl,
  }: Props = $props()

  let video = $state<File | null>(null)
  const target = $derived(roomInfo.target.substring(roomInfo.target.indexOf(':') + 1))
  const src = $derived(
    roomInfo.type === RoomType.RemoteFile ? target : video ? URL.createObjectURL(video) : null,
  )
  const name = $derived(
    roomInfo.type === RoomType.RemoteFile
      ? decodeURIComponent(target.substring(target.lastIndexOf('/') + 1))
      : target,
  )
  let menuOpen = $state(false)
  // If transientVideo matches up with the target, play it, else discard it
  $effect(() => {
    if (roomInfo.type === RoomType.LocalFile && transientVideo !== null) {
      if (video === null && name === transientVideo.name) video = transientVideo
      transientVideo = null
    }
  })

  const handleSelectVideo = async () => {
    try {
      video =
        (await openFileOrFiles({
          types: [
            // .mkv is not supported by Firefox (so far, tested on Linux + Chrome / Firefox)
            { description: 'Videos', accept: { 'video/*': ['.mp4', '.webm', '.mkv', '.mov'] } },
          ],
        })) ?? null
    } catch (e: unknown) {
      console.error('Failed to select local file!', e)
    }
  }

  const handleStop = async () => {
    try {
      await ky.patch(`api/room/${id}`, { json: { type: RoomType.None, target: '' } })
    } catch (e: unknown) {
      alert('Failed to stop video!')
      console.error('Failed to stop video!', e)
    }
  }
</script>

<div class="video-container">
  {#if src === null}
    <div class="video-select">
      <h1>Select {name} to start playing</h1>
      <DropdownButton
        primary={{ onclick: handleSelectVideo }}
        secondary={{ onclick: () => (menuOpen = !menuOpen) }}
      >
        {#snippet primaryChild()}Select local file{/snippet}
        {#snippet secondaryChild()}<CaretDown weight="bold" size="1rem" />{/snippet}
        <Menu open={menuOpen} onClose={() => (menuOpen = false)}>
          <MenuItem onclick={handleStop}>Play another video</MenuItem>
        </Menu>
      </DropdownButton>
    </div>
  {:else}
    <VideoPlayer
      {src}
      {name}
      {playerState}
      {onPlayerStateChange}
      bind:subtitles
      onStop={handleStop}
      {fullscreenEl}
    />
  {/if}
  {#if error}
    <h3 class="error-banner">Error: {error}<br />Reconnecting in 10s...</h3>
  {/if}
</div>

<style lang="scss">
  .video-select {
    flex-grow: 1;
    display: flex;
    flex-direction: column;

    min-height: 280px;
    justify-content: center;
    align-items: center;
    text-align: center;
    padding: 1rem;
    gap: 1rem;

    h1 {
      word-break: break-word;
    }
  }

  .error-banner {
    padding: 1rem;
    text-align: center;
    background-color: var(--error-color);
  }

  .video-container {
    justify-content: center;

    background-color: #000000;
    width: 100%;
    display: flex;
    flex-direction: column;
    color: white;
    @media screen and (min-width: 768px) {
      flex: 1;
    }
  }
</style>
