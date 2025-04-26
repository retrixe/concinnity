<script lang="ts">
  import ky from '$lib/api/ky'
  import { Button, TextInput } from 'heliodor'

  let register = $state({ username: '', password: '', confirmPw: '', email: '' })
  let disabled = $state(false)
  let error: string | null = $state(null)
  let verified = $state(false)

  const clearError = () => (error = null)

  async function onRegister() {
    disabled = true
    try {
      if (register.password !== register.confirmPw) {
        throw new Error(`Passwords do not match!`)
      }
      const res = await ky.post(`api/register`, { json: register }).json<{ verified?: boolean }>()
      if (res.verified) {
        verified = true
      }
      error = ''
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to register!`)
    }
    disabled = false
  }
</script>

<h2>Register</h2>
<div class="spacer"></div>
<label for="register-username">Username</label>
<TextInput
  id="register-username"
  bind:value={register.username}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="email"
  placeholder="e.g. retrixe"
/>
<label for="register-email">E-mail</label>
<TextInput
  id="register-email"
  bind:value={register.email}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="email"
  placeholder="e.g. aelia@retrixe.xyz"
/>
<label for="register-password">Password</label>
<TextInput
  id="register-password"
  bind:value={register.password}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="password"
/>
<label for="register-confirm-pw">Confirm Password</label>
<TextInput
  id="register-confirm-pw"
  bind:value={register.confirmPw}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="password"
  onkeypress={e => e.key === 'Enter' && onRegister() /* eslint-disable-line */}
/>
{#if error === '' && !verified}
  <p class="result">Registered successfully! Check your e-mail to verify your account.</p>
{:else if error === ''}
  <p class="result">
    Registered successfully! Go to the <a href="/login">login page</a> to get started.
  </p>
{:else if !!error}
  <p class="result error">{error}</p>
{/if}
<div class="spacer"></div>
<Button {disabled} onclick={onRegister}>Sign Up</Button>
<div class="spacer"></div>
<p>Already have an account? <a href="/login">Log in</a></p>
