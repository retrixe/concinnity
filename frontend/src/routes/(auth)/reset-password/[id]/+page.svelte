<script lang="ts">
  import Button from '$lib/lunaria/Button.svelte'
  import TextInput from '$lib/lunaria/TextInput.svelte'

  let password = $state('')
  let confirmPw = $state('')
  let disabled = $state(false)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  function onResetPassword() {
    disabled = true
    try {
      if (password !== confirmPw) {
        throw new Error(`Passwords do not match!`)
      }
      // await ky.post(`api/register`, { json: register }).json<{ token: string; username: string }>()
      error = ''
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to reset password!`)
    }
    disabled = false
  }
</script>

<h2>Reset Password</h2>
<div class="spacer"></div>
<p class="left-align">Enter your new account password for: <b>eslyfail</b></p>
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
  <p class="result">Reset password successfully!</p>
{:else if !!error}
  <p class="result error">{error}</p>
{/if}
<div class="spacer"></div>
<Button {disabled} onclick={onResetPassword}>Reset Password</Button>
<div class="spacer"></div>
<p>Want to try logging in? <a href="/login">Log in</a></p>
