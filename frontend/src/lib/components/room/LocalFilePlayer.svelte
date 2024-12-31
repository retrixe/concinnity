<script lang="ts">
  import type { PlayerState, RoomInfo } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import Button from '../Button.svelte'

  interface Props {
    error: string | null
    roomInfo: RoomInfo
    playerState: PlayerState
    transientVideo: File | null
  }

  let { error, roomInfo, playerState, transientVideo = $bindable(null) }: Props = $props()
  // FIXME: Implement VideoPlayer with synced video controls and a way to go back to landing state
  // FIXME: Autoplay may not work on browsers, so a manual play button may be needed
  $inspect(playerState)

  // FIXME: Clear this if room_info changes (how do we know? modifiedAt + target change? we don't get modifiedAt right now)
  let video = $state<File | null>(null)
  // If transientVideo matches up with the target, play it, else discard it
  $effect(() => {
    if (transientVideo !== null) {
      if (video === null && roomInfo.target === transientVideo.name) video = transientVideo
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
</script>

<div class="video-container">
  {#if video === null}
    <div class="video-select">
      <h1>Select {roomInfo.target} to start playing</h1>
      <br />
      <Button onclick={handleSelectVideo}>Select local file</Button>
    </div>
  {:else}
    <h1 style:flex-grow="1">Video: {video.name}</h1>
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
