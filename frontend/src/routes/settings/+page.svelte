<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { Box, Button, TextInput } from 'heliodor'

  const { userId, username, email } = $derived(page.data)

  $effect(() => {
    if (!username) goto('/login', { replaceState: true }).catch(console.error)
  })

  // TODO: Add change username functionality with dialog
  // TODO: Add change email functionality with dialog
  // TODO: Add change password functionality
  // TODO: Add delete account functionality with dialog
</script>

<div class="container">
  <h1>Account Settings</h1>

  <Box class="content">
    <div class="space-between">
      <div>
        <h4>Username</h4>
        <h2>{username}</h2>
      </div>
      <Button>Edit</Button>
    </div>
    <br />
    <div class="space-between">
      <div>
        <h4>Email</h4>
        <p>{email}</p>
      </div>
      <Button>Edit</Button>
    </div>
    <br />
    <h4>Account ID</h4>
    <p>{userId}</p>
  </Box>

  <Box class="content">
    <h3>Change Password</h3>
    <br />
    <TextInput class="changepw" type="password" placeholder="Enter your current password" />
    <TextInput class="changepw" type="password" placeholder="Enter new password" />
    <TextInput class="changepw" type="password" placeholder="Confirm password" />
    <div class="changepw-button">
      <Button>Change Password</Button>
    </div>
  </Box>

  <Box class="content">
    <Button class="error">Delete Account</Button>
  </Box>
</div>

<style lang="scss">
  :global(.changepw) {
    display: block;
  }

  .changepw-button {
    margin-top: 24px;
  }

  :global(.error) {
    background-color: var(--error-color) !important;
  }

  @mixin content {
    margin-top: 32px;
    width: 100%;
    max-width: 600px;
  }

  .container > :global(.content) {
    padding: 16px;
    @include content();
  }

  .container > h1 {
    @include content();
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
