<script lang="ts">
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import { RoomType } from '$lib/api/room'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import { CaretDown } from 'phosphor-svelte'
  import LinearProgress from '$lib/lunaria/LinearProgress.svelte'
  import Button from '$lib/lunaria/Button.svelte'
  import Dialog from '$lib/lunaria/Dialog.svelte'
  import DropdownButton from '$lib/lunaria/DropdownButton.svelte'
  import Menu from '$lib/lunaria/Menu.svelte'
  import MenuItem from '$lib/lunaria/MenuItem.svelte'
  import TextInput from '$lib/lunaria/TextInput.svelte'

  interface Props {
    error: string | null
    connecting: boolean
    transientVideo: File | null
  }

  const id = page.params.id
  let { error, connecting, transientVideo = $bindable(null) }: Props = $props()

  let menuOpen = $state(false)
  let remoteFileUrl = $state<string | null>(null)

  const handleOpenMenu = () => (menuOpen = !menuOpen)

  const handleOpenRemoteFileDialog = () => {
    menuOpen = false
    remoteFileUrl = ''
  }

  const handleSelectLocalFile = async () => {
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

  const handlePlayRemoteFile = async () => {
    if (!remoteFileUrl) return

    try {
      await ky.patch(`api/room/${id}`, {
        json: { type: RoomType.RemoteFile, target: `${Date.now()}:${remoteFileUrl}` },
      })
    } catch (e: unknown) {
      alert('Failed to play remote file!')
      console.error('Failed to play remote file!', e)
    }
  }
</script>

<div class="video" class:error>
  {#if error}
    <h1>
      Error encountered!
      {#if error !== 'You are not authenticated to access this resource!'}
        Reconnecting in 10s...
      {/if}
    </h1>
    <h2>{error}</h2>
  {:else if connecting}
    <h1>Connecting to room...</h1>
    <LinearProgress />
  {:else}
    <h1>No video playing</h1>
    <DropdownButton
      primary={{ onclick: handleSelectLocalFile }}
      secondary={{ onclick: handleOpenMenu }}
    >
      {#snippet primaryChild()}Select local file{/snippet}
      {#snippet secondaryChild()}<CaretDown weight="bold" size="16px" />{/snippet}
      <Menu open={menuOpen} onClose={handleOpenMenu}>
        <MenuItem onclick={handleOpenRemoteFileDialog}>Remote file (HTTP/S)</MenuItem>
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
  {/if}
</div>

<style lang="scss">
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

    // Linear progress
    :global(.loader) {
      max-width: 20rem;
    }

    :global(.dialog-content) {
      color: var(--color);
      gap: 1rem;
    }
  }
</style>
