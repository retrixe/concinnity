<script lang="ts">
  import { goto, invalidate } from '$app/navigation'
  import { resolve } from '$app/paths'
  import ky from '$lib/api/ky'
  import { Button, Dialog, TextInput } from 'heliodor'

  const { open, onClose } = $props<{
    open: boolean
    onClose: () => void
  }>()

  let currentPassword = $state('')
  let abortController: AbortController | null = $state(null)
  let disabled = $derived(!!abortController)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function handleDeleteAccount() {
    abortController = new AbortController()
    try {
      await ky
        .delete(`api/delete-account`, { json: { currentPassword }, signal: abortController.signal })
        .json()
      error = ''
      // Don't bother cleaning up the timeout, what if the user closes the dialog?
      localStorage.removeItem('concinnity:token')
      setTimeout(() => goto(resolve('/')).then(() => invalidate('app:auth')), 3000)
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to delete account!`)
    }
    abortController = null
  }

  function handleClose() {
    currentPassword = ''
    abortController?.abort()
    abortController = null
    error = null
    onClose()
  }
</script>

<Dialog {open} onClose={handleClose}>
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
      Account deleted successfully! You will be redirected to the
      <a href={resolve('/')}>homepage</a> shortly...
    </p>
  {/if}
  <Button class="delete-account-btn" {disabled} onclick={handleDeleteAccount}>Confirm</Button>
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
