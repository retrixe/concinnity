<script lang="ts">
  import { Box, Button } from 'heliodor'
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import DeleteAccountDialog from './DeleteAccountDialog.svelte'

  const { userId, username, email } = $derived(page.data)

  $effect(() => {
    if (!username) goto('/login', { replaceState: true }).catch(console.error)
  })

  let currentDialog: 'changeUsername' | 'changeEmail' | 'changePassword' | 'deleteAccount' | null =
    $state(null)

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
    <br />
    <div class="space-between">
      <div>
        <h4>Email</h4>
        <p>{email}</p>
      </div>
      <!-- <Button>Edit</Button> -->
    </div>
    <br />
    <h4>Account ID</h4>
    <p>{userId}</p>
  </Box>

  <Box class="content">
    <Button class="gap-right" onclick={() => (currentDialog = 'changePassword')}>
      Change Password
    </Button>
    <Button class="error" onclick={() => (currentDialog = 'deleteAccount')}>Delete Account</Button>
  </Box>
</div>

<DeleteAccountDialog
  open={currentDialog === 'deleteAccount'}
  onClose={() => (currentDialog = null)}
/>

<style lang="scss">
  :global(.gap-right) {
    margin-right: 12px;
  }

  :global(button.error) {
    background-color: var(--error-color) !important;
  }

  .container > :global(.content) {
    padding: 16px;
  }

  .container > h1,
  :global(.content) {
    margin-top: 32px;
    width: 100%;
    max-width: 600px;
  }

  .space-between {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .container {
    margin: 0 16px;
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    align-items: center;
  }
</style>
