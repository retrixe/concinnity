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
  let newUsername = $state('')

  let abortController: AbortController | null = $state(null)
  let disabled = $derived(!!abortController)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function handleChangeUsername() {
    abortController = new AbortController()
    try {
      await ky.post(`api/change-username`, {
        json: { currentPassword, newUsername },
        signal: abortController.signal,
      })
      await invalidate('app:auth')
      handleSuccess()
      handleClose()
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to change username!`)
    }
    abortController = null
  }

  function handleClose() {
    // TODO: Change layout data and update username cache!
    currentPassword = ''
    newUsername = ''
    abortController?.abort()
    abortController = null
    error = null
    onClose()
  }
</script>

<Dialog {open} onClose={handleClose}>
  <h2 class="gutter-bottom">Change Username</h2>
  <p class="gutter-bottom">Enter your current password and new username below.</p>
  <label for="change-username-current">Current Password</label>
  <TextInput
    id="change-username-current"
    class="gutter-bottom"
    bind:value={currentPassword}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
    placeholder="Enter your password"
  />
  <label for="change-username-new">New Username</label>
  <TextInput
    id="change-username-new"
    class="gutter-bottom"
    bind:value={newUsername}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="email"
    placeholder="Enter new username"
  />
  {#if error}
    <p class="gutter-bottom error">{error}</p>
  {/if}
  <Button class="change-username-btn" onclick={handleChangeUsername}>Submit</Button>
</Dialog>

<style lang="scss">
  .error {
    color: var(--error-color);
  }

  :global(.gutter-bottom) {
    margin-bottom: 1rem !important;
  }

  :global(.change-username-btn) {
    align-self: flex-end;
  }

  label {
    padding-bottom: 0.5rem;
    font-weight: bold;
  }
</style>
