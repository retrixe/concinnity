<script lang="ts">
  import { goto, invalidate } from '$app/navigation'
  import ky from '$lib/api/ky'
  import { Button, Dialog, TextInput } from 'heliodor'

  const { open, onClose } = $props<{
    open: boolean
    onClose: () => void
  }>()

  let currentPassword = $state('')
  let disabled = $state(false)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function handleDeleteAccount() {
    disabled = true
    try {
      await ky.delete(`api/delete-account`, { json: { currentPassword } }).json()
      error = ''
      // Don't bother cleaning up the timeout, what if the user closes the dialog?
      localStorage.removeItem('concinnity:token')
      setTimeout(() => goto('/').then(() => invalidate('app:auth')), 3000)
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to delete account!`)
    }
    disabled = false
  }
</script>

<Dialog {open} {onClose}>
  <h2 class="gutter-bottom">Delete Account</h2>
  <p class="gutter-bottom">
    Enter your password below and press "Confirm" to delete your account.
    <span style:color="var(--error-color)">
      This will delete your Concinnity data permanently and cannot be undone.
    </span>
  </p>
  <label for="delete-account-password">Current Password</label>
  <TextInput
    id="delete-account-password"
    class="gutter-bottom"
    bind:value={currentPassword}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
    placeholder="Enter your password"
  />
  {#if error}
    <p class="gutter-bottom error">{error}</p>
  {:else if error === ''}
    <p class="gutter-bottom">
      Account deleted successfully! You will be redirected to the <a href="/">homepage</a> shortly...
    </p>
  {/if}
  <Button class="delete-account-btn" onclick={handleDeleteAccount}>Confirm</Button>
</Dialog>

<style lang="scss">
  .error {
    color: var(--error-color);
  }

  :global(.gutter-bottom) {
    margin-bottom: 1rem !important;
  }

  :global(.delete-account-btn) {
    align-self: flex-end;
  }

  label {
    padding-bottom: 0.5rem;
    font-weight: bold;
  }
</style>
