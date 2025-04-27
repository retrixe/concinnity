<script lang="ts">
  import ky from '$lib/api/ky'
  import { Button, TextInput } from 'heliodor'

  let usernameEmail = $state('')
  let disabled = $state(false)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function onForgotPassword() {
    disabled = true
    try {
      const { success } = await ky
        .post(`api/login?user=${encodeURIComponent(usernameEmail)}`)
        .json<{ success: boolean }>()
      error = success ? '' : 'Failed to send reset password e-mail!'
    } catch (e: unknown) {
      error =
        e instanceof Error ? e.message : (e?.toString() ?? `Failed to send reset password e-mail!`)
    }
    disabled = false
  }
</script>

<h2>Forgot your password?</h2>
<div class="spacer"></div>
<p class="left-align">
  No worries! Enter your email address and we will send you a link to reset your password.
</p>
<div class="spacer"></div>
<label for="forgot-password-username-email">Username / E-mail</label>
<TextInput
  id="forgot-password-username-email"
  bind:value={usernameEmail}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="email"
  placeholder="e.g. aelia@retrixe.xyz"
  onkeypress={e => e.key === 'Enter' && onForgotPassword() /* eslint-disable-line */}
/>
{#if error === ''}
  <p class="result">Sent reset link successfully! Keep an eye on your e-mail...</p>
{:else if !!error}
  <p class="result error">{error}</p>
{/if}
<div class="spacer"></div>
<Button {disabled} onclick={onForgotPassword}>Send E-mail</Button>
<div class="spacer"></div>
<p>Want to try logging in? <a href="/login">Log in</a></p>
