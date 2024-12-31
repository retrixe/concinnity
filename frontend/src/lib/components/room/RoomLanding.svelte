<script lang="ts">
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import { RoomType } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import Button from '../Button.svelte'

  interface Props {
    error: string | null
    connecting: boolean
    transientVideo: File | null
  }

  let { error, connecting, transientVideo = $bindable(null) }: Props = $props()
  const id = page.params.id

  const onclick = async () => {
    const file = await openFileOrFiles()
    if (!file) return

    try {
      const req = await fetch(`${PUBLIC_BACKEND_URL}/api/room/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ type: RoomType.LocalFile, target: file.name }),
        headers: { authorization: localStorage.getItem('concinnity:token') ?? '' },
      })
      if (!req.ok) {
        // TODO: Better error handling? Maybe send it as a system message.
        console.error('Failed to select local file!', req)
      } else {
        transientVideo = file
      }
    } catch (e: unknown) {
      console.error('Failed to select local file!', e)
    }
  }
</script>

<div class="video" class:error>
  {#if error}
    <h1>Error encountered! Reconnecting in 10s...</h1>
    <h2>{error}</h2>
  {:else if connecting}
    <!-- TODO: Loading spinner? -->
    <h1>Connecting to room...</h1>
  {:else}
    <h1>No video playing</h1>
    <br />
    <Button {onclick}>Select local file</Button>
  {/if}
</div>

<style lang="scss">
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
    min-height: 280px;
    justify-content: center;
    align-items: center;
    text-align: center;
    padding: 1rem;

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
