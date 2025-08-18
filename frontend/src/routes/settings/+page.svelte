<script lang="ts">
  import { Box, Button, IconButton, Toast } from 'heliodor'
  import { Check, User, X } from 'phosphor-svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import DeleteAccountDialog from './DeleteAccountDialog.svelte'
  import ChangePasswordDialog from './ChangePasswordDialog.svelte'
  import ChangeUsernameDialog from './ChangeUsernameDialog.svelte'
  import ChangeEmailDialog from './ChangeEmailDialog.svelte'

  const { userId, username, email, avatar } = $derived(page.data)

  $effect(() => {
    if (!username) goto('/login', { replaceState: true }).catch(console.error)
  })

  let currentDialog: 'changeUsername' | 'changeEmail' | 'changePassword' | 'deleteAccount' | null =
    $state(null)

  let successMessage: string | null = $state(null)

  const handleDismissToast = () => (successMessage = null)
</script>

<div class="container">
  <h1>Account Settings</h1>

  <Box class="content">
    <div class="profile-container">
      {#if typeof avatar === 'string'}
        <img src={`${PUBLIC_BACKEND_URL}/api/avatar/${avatar}`} alt="User Avatar" class="avatar" />
      {:else}
        <User size="192px" />
      {/if}
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
    display: flex;
    justify-content: center;
    margin-bottom: 32px;
    > :global(svg) {
      border: 1px solid var(--divider-color);
      border-radius: 50%;
    }
  }

  .avatar {
    border-radius: 50%;
    width: 192px;
    height: 192px;
  }
</style>
