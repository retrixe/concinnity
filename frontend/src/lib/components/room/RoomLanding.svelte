<script lang="ts">
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import { RoomType } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import Button from '../Button.svelte'
  import LinearProgress from '../LinearProgress.svelte'

  interface Props {
    error: string | null
    connecting: boolean
    transientVideo: File | null
  }

  const id = page.params.id
  let { error, connecting, transientVideo = $bindable(null) }: Props = $props()

  const onclick = async () => {
    try {
      const file = await openFileOrFiles({
        types: [
          // .mkv is not supported by Firefox (so far, tested on Linux + Chrome / Firefox)
          { description: 'Videos', accept: { 'video/*': ['.mp4', '.webm', '.mkv', '.mov'] } },
        ],
      })
      if (!file) return

      await ky.patch(`api/room/${id}`, {
        json: { type: RoomType.LocalFile, target: `${Date.now()}:${file.name}` },
      })
      transientVideo = file
    } catch (e: unknown) {
      alert('Failed to select local file!')
      console.error('Failed to select local file!', e)
    }
  }
</script>

<div class="video" class:error>
  {#if error}
    <h1>Error encountered! Reconnecting in 10s...</h1>
    <h2>{error}</h2>
  {:else if connecting}
    <h1>Connecting to room...</h1>
    <LinearProgress />
  {:else}
    <h1>No video playing</h1>
    <Button {onclick}>Select local file</Button>
  {/if}
</div>

<style lang="scss">
  // Linear progress
  :global(.loader) {
    max-width: 20rem;
  }

  .error {
    h1 {
      color: var(--error-color);
    }
    h2 {
      font-weight: 300;
    }
  }

  .video {
    min-height: 280px;
    justify-content: center;
    align-items: center;
    text-align: center;
    padding: 1rem;
    gap: 1rem;

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
