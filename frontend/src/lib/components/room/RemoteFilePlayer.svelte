<script lang="ts">
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import { RoomType, type PlayerState, type RoomInfo } from '$lib/api/room'
  import VideoPlayer from './VideoPlayer.svelte'

  interface Props {
    error: string | null
    roomInfo: RoomInfo
    playerState: PlayerState
    subtitles: Record<string, string | null>
    onPlayerStateChange: (newState: PlayerState) => void
    fullscreenEl: Element
  }

  const id = page.params.id
  let {
    error,
    roomInfo,
    playerState,
    subtitles = $bindable(),
    onPlayerStateChange,
    fullscreenEl,
  }: Props = $props()
  const src = roomInfo.target.substring(roomInfo.target.indexOf(':') + 1)
  const name = $derived(decodeURIComponent(src.substring(src.lastIndexOf('/') + 1)))

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
  <VideoPlayer
    {src}
    {name}
    {playerState}
    {onPlayerStateChange}
    bind:subtitles
    onStop={handleStop}
    {fullscreenEl}
  />
  {#if error}
    <h3 class="error-banner">Error: {error}<br />Reconnecting in 10s...</h3>
  {/if}
</div>

<style lang="scss">
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
