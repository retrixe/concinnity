<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { Box } from 'heliodor'

  const { userId, username, email } = $derived(page.data)

  $effect(() => {
    if (!username) goto('/login', { replaceState: true }).catch(console.error)
  })

  // TODO: Add change username functionality
  // TODO: Add change email functionality
  // TODO: Add change password functionality
  // TODO: Add delete account functionality
</script>

<div class="container">
  <h1>Account Settings</h1>
  <Box class="content">
    <h4>Username</h4>
    <h2>{username}</h2>
    <br />
    <h4>Email</h4>
    <p>{email}</p>
    <br />
    <h4>Account ID</h4>
    <p>{userId}</p>
  </Box>
</div>

<style lang="scss">
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

  .container {
    margin: 0 16px;
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    align-items: center;
  }
</style>
