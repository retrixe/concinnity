<script lang="ts">
  import { Box, Button, Toast } from 'heliodor'
  import { Check } from 'phosphor-svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import DeleteAccountDialog from './DeleteAccountDialog.svelte'
  import ChangePasswordDialog from './ChangePasswordDialog.svelte'

  const { userId, username, email } = $derived(page.data)

  $effect(() => {
    if (!username) goto('/login', { replaceState: true }).catch(console.error)
  })

  let currentDialog: 'changeUsername' | 'changeEmail' | 'changePassword' | 'deleteAccount' | null =
    $state(null)

  let successMessage: string | null = $state(null)

  // TODO: Add change password functionality with dialog
  // TODO: Add change username functionality with dialog
  // TODO: Add change email functionality with dialog
</script>

<div class="container">
  <h1>Account Settings</h1>

  <Box class="content">
    <div class="space-between">
      <div>
        <h4>Username</h4>
        <h2>{username}</h2>
      </div>
      <!-- <Button>Edit</Button> -->
    </div>
    <hr />
    <div class="space-between">
      <div>
        <h4>Email</h4>
        <p>{email}</p>
      </div>
      <!-- <Button>Edit</Button> -->
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

{#if successMessage !== null}
  <Toast
    message={successMessage}
    duration={3000}
    onclose={() => (successMessage = null)}
    color="success"
  >
    {#snippet icon(color)}
      <Check {color} weight="bold" size="1.5rem" />
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
</style>
