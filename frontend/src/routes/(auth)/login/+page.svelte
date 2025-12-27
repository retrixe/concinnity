<script lang="ts">
  import { invalidate } from '$app/navigation'
  import { resolve } from '$app/paths'
  import ky from '$lib/api/ky'
  import { Button, TextInput } from 'heliodor'

  let login = $state({ username: '', password: '' })
  let disabled = $state(false)
  let error: string | null = $state(null)

  const clearError = () => (error = null)

  async function onLogin() {
    disabled = true
    try {
      const res = await ky
        .post(`api/login`, { json: login })
        .json<{ token: string; username: string }>()
      localStorage.setItem('concinnity:token', res.token)
      error = ''
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to login!`)
    }
    disabled = false
    if (!error) {
      invalidate('app:auth').catch(console.error)
    }
  }
</script>

<h2>Login</h2>
<div class="spacer"></div>
<label for="login-username">E-mail / Username</label>
<TextInput
  id="login-username"
  bind:value={login.username}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="email"
  placeholder="e.g. aelia@retrixe.xyz"
/>
<label for="login-password">Password</label>
<TextInput
  id="login-password"
  bind:value={login.password}
  oninput={clearError}
  error={!!error}
  {disabled}
  type="password"
  onkeypress={e => e.key === 'Enter' && onLogin()}
/>
{#if error === ''}
  <p class="center">Logged in successfully! You should be redirected shortly...</p>
{:else if !!error}
  <p class="center error">{error}</p>
{/if}
<div class="spacer"></div>
<Button {disabled} onclick={onLogin}>Login</Button>
<div class="spacer"></div>
<p class="center">Don't have an account? <a href={resolve('/register')}>Sign up</a></p>
<p class="center">
  Forgot your password? <a href={resolve('/forgot-password')}>Reset via e-mail</a>
</p>
