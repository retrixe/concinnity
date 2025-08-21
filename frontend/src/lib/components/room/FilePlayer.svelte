<script lang="ts">
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import { RoomType, type PlayerState, type RoomInfo } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import { CaretDown } from 'phosphor-svelte'
  import VideoPlayer from './VideoPlayer.svelte'
  import { Button, Dialog, DropdownButton, Menu, MenuItem, TextInput } from 'heliodor'

  interface Props {
    error: string | null
    reconnecting: number
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
    reconnecting,
    roomInfo,
    playerState,
    subtitles = $bindable(),
    onPlayerStateChange,
    transientVideo = $bindable(null),
    fullscreenEl,
  }: Props = $props()

  let currentVideo = $state<File | string | null>(
    roomInfo.type === RoomType.RemoteFile
      ? roomInfo.target.substring(roomInfo.target.indexOf(':') + 1)
      : null,
  )
  const src = $derived(
    currentVideo && typeof currentVideo !== 'string'
      ? URL.createObjectURL(currentVideo)
      : currentVideo,
  )
  const target = $derived(roomInfo.target.substring(roomInfo.target.indexOf(':') + 1))
  const name = $derived(
    roomInfo.type === RoomType.RemoteFile
      ? decodeURIComponent(target.substring(target.lastIndexOf('/') + 1))
      : target,
  )
  // If transientVideo matches up with the target, play it, else discard it
  $effect(() => {
    if (roomInfo.type === RoomType.LocalFile && transientVideo !== null) {
      if (currentVideo === null && name === transientVideo.name) currentVideo = transientVideo
      transientVideo = null
    }
  })

  let menuOpen = $state(false)
  let remoteFileUrl: string | null = $state(null)

  const handleSelectVideo = async () => {
    try {
      currentVideo =
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

  const handlePlayRemoteFile = () => {
    if (remoteFileUrl) {
      currentVideo = remoteFileUrl
      remoteFileUrl = null
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
          <MenuItem onclick={() => (remoteFileUrl = '')}>Play from remote URL</MenuItem>
          <MenuItem onclick={handleStop}>Stop playing this video</MenuItem>
        </Menu>
      </DropdownButton>
      <Dialog open={remoteFileUrl !== null} onClose={() => (remoteFileUrl = null)}>
        <h2>Enter URL of remote file</h2>
        <!-- eslint-disable @typescript-eslint/no-non-null-assertion -->
        <TextInput
          bind:value={remoteFileUrl!}
          type="url"
          placeholder="e.g. https://retrixe.xyz/example.mp4"
        />
        <!-- eslint-enable @typescript-eslint/no-non-null-assertion -->
        <Button onclick={handlePlayRemoteFile}>Play</Button>
      </Dialog>
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
      customActions={{ 'Change video file': () => (currentVideo = null) }}
    />
  {/if}
  {#if error}
    <h3 class="error-banner">
      Error: {error}
      {#if reconnecting === 0}
        <br />Reconnecting...
      {:else if reconnecting > 0}
        <br />Reconnecting in {reconnecting}s...
      {/if}
    </h3>
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

  :global(.dialog-content) {
    color: var(--color);
    gap: 1rem;
  }
</style>
