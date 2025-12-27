<script lang="ts">
  import { Box, Button, IconButton, Toast } from 'heliodor'
  import { Check, Pencil, Trash, User, X, XCircle } from 'phosphor-svelte'
  import { goto, invalidate } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import { openFileOrFiles } from '$lib/utils/openFile'
  import ky from '$lib/api/ky'
  import DeleteAccountDialog from './DeleteAccountDialog.svelte'
  import ChangePasswordDialog from './ChangePasswordDialog.svelte'
  import ChangeUsernameDialog from './ChangeUsernameDialog.svelte'
  import ChangeEmailDialog from './ChangeEmailDialog.svelte'
  import ClearAvatarDialog from './ClearAvatarDialog.svelte'

  const { userId, username, email, avatar } = $derived(page.data)

  $effect(() => {
    if (!username) goto(resolve('/login'), { replaceState: true }).catch(console.error)
  })

  let currentDialog:
    | 'clearAvatar'
    | 'changeUsername'
    | 'changeEmail'
    | 'changePassword'
    | 'deleteAccount'
    | null = $state(null)

  let toastMessage: [boolean, string] | null = $state(null)

  const handleDismissToast = () => (toastMessage = null)

  let avatarAbortController: AbortController | null = $state(null)

  const handleChangeAvatar = async () => {
    const file = await openFileOrFiles({
      multiple: false,
      types: [
        {
          description: 'Images',
          accept: {
            'image/png': ['.png'],
            'image/jpeg': ['.jpeg', '.jpg'],
            'image/jpg': ['.jpg', '.jpeg'],
            'image/gif': ['.gif'],
            'image/webp': ['.webp'],
            'image/avif': ['.avif'],
            'image/bmp': ['.bmp'],
            'image/tiff': ['.tiff', '.tif'],
          },
        },
      ],
    })
    if (!file) return
    // Post the request
    avatarAbortController = new AbortController()
    try {
      await ky.post(`api/avatar`, { body: file, signal: avatarAbortController.signal }).json()
      // If successful, show a toast
      await invalidate('app:auth')
      toastMessage = [true, 'Avatar changed successfully!']
    } catch (e: unknown) {
      toastMessage = [
        false,
        e instanceof Error ? e.message : (e?.toString() ?? `Failed to clear avatar!`),
      ]
    }
    avatarAbortController = null
  }
</script>

<div class="container">
  <h1>Account Settings</h1>

  <Box class="content">
    <div class="profile-container">
      {#if typeof avatar === 'string'}
        <img src={`${PUBLIC_BACKEND_URL}/api/avatar/${avatar}`} alt="User Avatar" class="avatar" />
      {:else}
        <User size="15rem" />
      {/if}
      <div class="profile-buttons">
        <IconButton onclick={handleChangeAvatar} disabled={!!avatarAbortController}>
          <Pencil size="1.5rem" />
        </IconButton>
        {#if avatar}
          <IconButton
            onclick={() => (currentDialog = 'clearAvatar')}
            disabled={!!avatarAbortController}
          >
            <Trash color="var(--error-color)" size="1.5rem" />
          </IconButton>
        {/if}
      </div>
    </div>
    <div class="space-between">
      <div>
        <h4>Username</h4>
        <h2>{username}</h2>
      </div>
      <Button onclick={() => (currentDialog = 'changeUsername')}>Edit</Button>
    </div>
    <hr />
    <div class="space-between">
      <div>
        <h4>Email</h4>
        <p>{email}</p>
      </div>
      <Button onclick={() => (currentDialog = 'changeEmail')}>Edit</Button>
    </div>
    <hr />
    <h4>Account ID</h4>
    <p>{userId}</p>
  </Box>

  <Box class="content row-buttons">
    <Button onclick={() => (currentDialog = 'changePassword')}>Change Password</Button>
    <Button color="error" onclick={() => (currentDialog = 'deleteAccount')}>Delete Account</Button>
  </Box>
</div>

<ClearAvatarDialog
  open={currentDialog === 'clearAvatar'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (toastMessage = [true, 'Avatar cleared successfully!'])}
/>

<DeleteAccountDialog
  open={currentDialog === 'deleteAccount'}
  onClose={() => (currentDialog = null)}
/>

<ChangePasswordDialog
  open={currentDialog === 'changePassword'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (toastMessage = [true, 'Password changed successfully!'])}
/>

<ChangeEmailDialog
  open={currentDialog === 'changeEmail'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (toastMessage = [true, 'E-mail changed successfully!'])}
/>

<ChangeUsernameDialog
  open={currentDialog === 'changeUsername'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (toastMessage = [true, 'Username changed successfully!'])}
/>

{#if toastMessage !== null}
  <Toast
    message={toastMessage[1]}
    duration={3000}
    onclose={handleDismissToast}
    color={toastMessage[0] ? 'success' : 'error'}
  >
    {#snippet icon()}
      {#if toastMessage?.[0]}
        <Check weight="bold" size="1.5rem" />
      {:else}
        <XCircle weight="bold" size="1.5rem" />
      {/if}
    {/snippet}
    {#snippet footer()}
      <IconButton onclick={handleDismissToast} aria-label="Close">
        <X weight="thin" size="1.5rem" />
      </IconButton>
    {/snippet}
  </Toast>
{/if}

<style lang="scss">
  hr {
    margin: 16px 0;
  }

  .space-between {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .container > :global(.content) {
    padding: 1rem;
  }

  .container > :global(.row-buttons) {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 16px;
  }

  .container > :global(*) {
    width: 100%;
    max-width: 600px;
  }

  .container {
    margin: 2rem 1rem;
    gap: 32px;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .profile-container {
    display: grid;
    justify-content: center;
    margin: 32px 0;
    gap: 1rem;
    > :global(svg) {
      border: 1px solid var(--divider-color);
    }
  }

  .profile-container > :global(svg),
  .avatar {
    grid-area: 1 / 1;
    border-radius: 50%;
    width: 15rem;
    height: 15rem;
  }

  .profile-buttons {
    grid-area: 1 / 1;
    place-self: end;

    display: flex;
    gap: 4px;
    border: 1px solid var(--divider-color);
    border-radius: 0.5rem;
    background-color: var(--surface-color);
    box-shadow: 0 0 1rem rgba(0, 0, 0, 0.2);
  }
</style>
