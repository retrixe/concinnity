<script lang="ts">
  import { Box, Button, IconButton, Toast } from 'heliodor'
  import { Check, Pencil, Trash, User, X } from 'phosphor-svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import DeleteAccountDialog from './DeleteAccountDialog.svelte'
  import ChangePasswordDialog from './ChangePasswordDialog.svelte'
  import ChangeUsernameDialog from './ChangeUsernameDialog.svelte'
  import ChangeEmailDialog from './ChangeEmailDialog.svelte'
  import ClearAvatarDialog from './ClearAvatarDialog.svelte'
  import { openFileOrFiles } from '$lib/utils/openFile'

  const { userId, username, email, avatar } = $derived(page.data)

  $effect(() => {
    if (!username) goto('/login', { replaceState: true }).catch(console.error)
  })

  let currentDialog:
    | 'clearAvatar'
    | 'changeUsername'
    | 'changeEmail'
    | 'changePassword'
    | 'deleteAccount'
    | null = $state(null)

  let successMessage: string | null = $state(null)

  const handleDismissToast = () => (successMessage = null)

  const handleChangeAvatar = async () => {
    const file = await openFileOrFiles({
      multiple: false,
      types: [
        {
          description: 'JPEG/PNG images',
          accept: {
            'image/png': ['.png'],
            'image/jpeg': ['.jpeg', '.jpg'],
            'image/jpg': ['.jpg', '.jpeg'],
          },
        },
      ],
    })
    // TODO: Post the request
    // TODO: Show a toast if error or success
    console.log(file)
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
        <IconButton onclick={handleChangeAvatar}>
          <Pencil size="1.5rem" />
        </IconButton>
        {#if avatar}
          <IconButton onclick={() => (currentDialog = 'clearAvatar')}>
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
  onSuccess={() => (successMessage = 'Avatar cleared successfully!')}
/>

<DeleteAccountDialog
  open={currentDialog === 'deleteAccount'}
  onClose={() => (currentDialog = null)}
/>

<ChangePasswordDialog
  open={currentDialog === 'changePassword'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (successMessage = 'Password changed successfully!')}
/>

<ChangeEmailDialog
  open={currentDialog === 'changeEmail'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (successMessage = 'E-mail changed successfully!')}
/>

<ChangeUsernameDialog
  open={currentDialog === 'changeUsername'}
  onClose={() => (currentDialog = null)}
  onSuccess={() => (successMessage = 'Username changed successfully!')}
/>

{#if successMessage !== null}
  <Toast message={successMessage} duration={3000} onclose={handleDismissToast} color="success">
    {#snippet icon()}<Check weight="bold" size="1.5rem" />{/snippet}
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
