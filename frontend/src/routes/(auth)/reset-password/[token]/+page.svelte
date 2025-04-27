<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import { Button, LinearProgress, TextInput } from 'heliodor'
  import { onDestroy } from 'svelte'

  const token = $derived(page.params.token)
  let password = $state('')
  let confirmPw = $state('')
  let disabled = $state(false)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  const tokenInfoRequest = $derived(
    ky.get(`api/forgot-password/${token}`).json<{ username: string }>(),
  )
  let redirectTimeoutId: number | undefined
  onDestroy(() => clearTimeout(redirectTimeoutId))

  async function onResetPassword() {
    disabled = true
    try {
      if (password !== confirmPw) {
        throw new Error(`Passwords do not match!`)
      }
      const { success } = await ky
        .post(`api/reset-password`, { json: { token, password } })
        .json<{ success: boolean }>()
      error = success ? '' : 'Failed to reset password!'
      redirectTimeoutId = setTimeout(() => goto('/login'), 5000)
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to reset password!`)
      disabled = false
    }
  }
</script>

{#await tokenInfoRequest}
  <LinearProgress />
{:then tokenInfo}
  <h2>Reset Password</h2>
  <div class="spacer"></div>
  <p class="left-align">Enter your new account password for: <b>{tokenInfo.username}</b></p>
  <div class="spacer"></div>
  <label for="reset-password">Password</label>
  <TextInput
    id="reset-password"
    bind:value={password}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
  />
  <label for="reset-confirm-pw">Confirm Password</label>
  <TextInput
    id="reset-confirm-pw"
    bind:value={confirmPw}
    oninput={clearError}
    error={!!error}
    {disabled}
    type="password"
  onkeypress={e => e.key === 'Enter' && onResetPassword() /* eslint-disable-line */}
  />
  {#if error === ''}
    <p class="result">
      Reset password successfully! Redirecting you to the
      <a href="/login">login page</a> in 5s...
    </p>
  {:else if !!error}
    <p class="result error">{error}</p>
  {/if}
  <div class="spacer"></div>
  <Button {disabled} onclick={onResetPassword}>Reset Password</Button>
{:catch e}
  <h2>An error occurred.</h2>
  <p class="left-align spacer">
    {e instanceof Error
      ? e.message
      : ((e as unknown)?.toString() ?? `Failed to fetch reset password token info!`)}
  </p>
  <hr class="spacer" />
{/await}
<div class="spacer"></div>
<p>Want to try logging in? <a href="/login">Log in</a></p>
