<script lang="ts">
  import { invalidate } from '$app/navigation'
  import ky from '$lib/api/ky'
  import { Button, Dialog } from 'heliodor'

  const {
    open,
    onClose,
    onSuccess: handleSuccess,
  } = $props<{
    open: boolean
    onClose: () => void
    onSuccess: () => void
  }>()

  let abortController: AbortController | null = $state(null)
  let disabled = $derived(!!abortController)
  let error: string | null = $state(null)

  async function handleClearAvatar() {
    abortController = new AbortController()
    try {
      await ky.post(`api/avatar`, { signal: abortController.signal }).json()
      await invalidate('app:auth')
      handleSuccess()
      handleClose()
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to clear avatar!`)
    }
    abortController = null
  }

  function handleClose() {
    abortController?.abort()
    abortController = null
    error = null
    onClose()
  }
</script>

<Dialog {open} onClose={handleClose}>
  <h2 class="gutter-bottom">Clear Avatar</h2>
  <p class="gutter-bottom">Are you sure you want to remove your avatar?</p>
  {#if error}
    <p class="gutter-bottom error">{error}</p>
  {/if}
  <Button class="clear-avatar-btn" {disabled} onclick={handleClearAvatar}>Confirm</Button>
</Dialog>

<style lang="scss">
  .error {
    color: var(--error-color);
  }

  :global(.gutter-bottom) {
    margin-bottom: 1rem !important;
  }

  :global(.clear-avatar-btn) {
    align-self: flex-end;
  }
</style>
