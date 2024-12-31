<script lang="ts">
  import type { PlayerState, RoomInfo } from '$lib/api/room'
  import Button from '../Button.svelte'

  interface Props {
    error: string | null
    roomInfo: RoomInfo
    playerState: PlayerState
    transientVideo: File | null
  }

  let { error, roomInfo, playerState, transientVideo = $bindable(null) }: Props = $props()
  // FIXME: Implement VideoPlayer with controls (video + go back to landing + sync up)
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

  // FIXME: Warning below the video if error
  // FIXME: Implement a way to select a video when a video is already requested to play in room info
  // FIXME: Autoplay may not work on browsers, so a manual play button may be needed
</script>

<div class="video" class:contents={video === null || error} class:error>
  {#if error && video === null}
    <h1>Error encountered! Reconnecting in 10s...</h1>
    <h2>{error}</h2>
  {:else if video === null}
    <h1>Select {roomInfo.target} to start playing</h1>
    <br />
    <Button>Select local file</Button>
  {:else}
    <h1>Video: {video.name}</h1>
  {/if}
</div>

<style lang="scss">
  .contents {
    min-height: 280px;
    text-align: center;
    padding: 1rem;
  }

  .error {
    gap: 0.5rem;
    h1 {
      color: var(--error-color);
    }
    h2 {
      font-weight: 300;
    }
  }

  .video {
    background-color: #000000;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    color: white;
    @media screen and (min-width: 768px) {
      flex: 1;
    }
  }
</style>
