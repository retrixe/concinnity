<script lang="ts">
  import { invalidate } from '$app/navigation'
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
  let newEmail = $state('')

  let abortController: AbortController | null = $state(null)
  let disabled = $derived(!!abortController)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function handleChangeEmail() {
    abortController = new AbortController()
    try {
      await ky.post(`api/change-email`, {
        json: { currentPassword, newEmail },
        signal: abortController.signal,
      })
      await invalidate('app:auth')
      handleSuccess()
      handleClose()
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to change e-mail!`)
    }
    abortController = null
  }

  function handleClose() {
    currentPassword = ''
    newEmail = ''
    abortController?.abort()
    abortController = null
    error = null
    onClose()
  }
</script>

<Dialog {open} onClose={handleClose}>
  <h2 class="gutter-bottom">Change Email</h2>
  <p class="gutter-bottom">Enter your current password and new email below.</p>
  <label for="change-email-current">Current Password</label>
  <TextInput
    id="change-email-current"
    class="gutter-bottom"
    bind:value={currentPassword}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
    placeholder="Enter your password"
  />
  <label for="change-email-new">New Email</label>
  <TextInput
    id="change-email-new"
    class="gutter-bottom"
    bind:value={newEmail}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="email"
    placeholder="Enter new email"
  />
  {#if error}
    <p class="gutter-bottom error">{error}</p>
  {/if}
  <Button class="change-email-btn" onclick={handleChangeEmail}>Submit</Button>
</Dialog>

<style lang="scss">
  .error {
    color: var(--error-color);
  }

  :global(.gutter-bottom) {
    margin-bottom: 1rem !important;
  }

  :global(.change-email-btn) {
    align-self: flex-end;
  }

  label {
    padding-bottom: 0.5rem;
    font-weight: bold;
  }
</style>
