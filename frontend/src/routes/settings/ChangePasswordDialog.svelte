<script lang="ts">
  import ky from '$lib/api/ky'
  import { Button, Dialog, TextInput } from 'heliodor'

  const {
    open,
    onClose,
    onSuccess: handleSuccess,
  } = $props<{
    open: boolean
    onClose: () => void
    onSuccess: () => void
  }>()

  let currentPassword = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')

  let abortController: AbortController | null = $state(null)
  let disabled = $derived(!!abortController)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function handleChangePassword() {
    if (newPassword !== confirmPassword) {
      error = 'Both passwords do not match!'
      return
    }
    abortController = new AbortController()
    try {
      await ky.post(`api/change-password`, {
        json: { currentPassword, newPassword },
        signal: abortController.signal,
      })
      handleSuccess()
      handleClose()
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to change password!`)
    }
    abortController = null
  }

  function handleClose() {
    currentPassword = ''
    newPassword = ''
    confirmPassword = ''
    abortController?.abort()
    abortController = null
    error = null
    onClose()
  }
</script>

<Dialog {open} onClose={handleClose}>
  <h2 class="gutter-bottom">Change Password</h2>
  <p class="gutter-bottom">Enter your current password and new password below.</p>
  <label for="change-password-current">Current Password</label>
  <TextInput
    id="change-password-current"
    class="gutter-bottom"
    bind:value={currentPassword}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
    placeholder="Enter your password"
  />
  <label for="change-password-new">New Password</label>
  <TextInput
    id="change-password-new"
    class="gutter-bottom"
    bind:value={newPassword}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
    placeholder="Enter new password"
  />
  <label for="change-password-confirm">Confirm New Password</label>
  <TextInput
    id="change-password-confirm"
    class="gutter-bottom"
    bind:value={confirmPassword}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
    placeholder="Enter new password again"
  />
  {#if error}
    <p class="gutter-bottom error">{error}</p>
  {/if}
  <Button class="change-password-btn" {disabled} onclick={handleChangePassword}>Submit</Button>
</Dialog>

<style lang="scss">
  .error {
    color: var(--error-color);
  }

  :global(.gutter-bottom) {
    margin-bottom: 1rem !important;
  }

  :global(.change-password-btn) {
    align-self: flex-end;
  }

  label {
    padding-bottom: 0.5rem;
    font-weight: bold;
  }
</style>
