<script lang="ts">
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import { RoomType, type PlayerState, type RoomInfo } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import Button from '../Button.svelte'
  import VideoPlayer from './VideoPlayer.svelte'

  interface Props {
    error: string | null
    roomInfo: RoomInfo
    playerState: PlayerState
    transientVideo: File | null
    fullscreenEl: Element
  }

  const id = page.params.id
  let {
    error,
    roomInfo,
    playerState,
    transientVideo = $bindable(null),
    fullscreenEl,
  }: Props = $props()
  const targetName = $derived(roomInfo.target.substring(roomInfo.target.indexOf(':') + 1))

  let video = $state<File | null>(null)
  // If transientVideo matches up with the target, play it, else discard it
  $effect(() => {
    if (transientVideo !== null) {
      if (video === null && targetName === transientVideo.name) video = transientVideo
      transientVideo = null
    }
  })

  const handleSelectVideo = async () => {
    try {
      video = (await openFileOrFiles()) ?? null
    } catch (e: unknown) {
      console.error('Failed to select local file!', e)
    }
  }

  const handleStop = async () => {
    try {
      const req = await fetch(`${PUBLIC_BACKEND_URL}/api/room/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ type: RoomType.None, target: '' }),
        headers: { authorization: localStorage.getItem('concinnity:token') ?? '' },
      })
      if (!req.ok) {
        // TODO: Better error handling? Maybe send it as a system message.
        console.error('Failed to remove video from room!', req)
      }
    } catch (e: unknown) {
      console.error('Failed to remove video from room!', e)
    }
  }
</script>

<div class="video-container">
  {#if video === null}
    <div class="video-select">
      <h1>Select {targetName} to start playing</h1>
      <br />
      <Button onclick={handleSelectVideo}>Select local file</Button>
    </div>
  {:else}
    <VideoPlayer {video} {playerState} onStop={handleStop} {fullscreenEl} />
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
