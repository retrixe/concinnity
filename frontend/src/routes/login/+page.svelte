<script lang="ts">
  import { goto, invalidate } from '$app/navigation'
  import { page } from '$app/state'
  import { PUBLIC_CONCINNITY_URL } from '$env/static/public'
  import Box from '$lib/components/Box.svelte'
  import Button from '$lib/components/Button.svelte'
  import TextInput from '$lib/components/TextInput.svelte'

  let progress: 'login' | 'register' | 'none' = $state('none')
  let login = $state({ username: '', password: '' })
  let loginError: string | false = $state('')
  let register = $state({ username: '', password: '', email: '' })
  let registerError: string | false = $state('')

  const { username } = $derived(page.data)
  $effect(() => {
    if (username) goto('/').catch(console.error)
  })

  async function onLogin() {
    progress = 'login'
    const result = await handleLoginRegister('login')
    progress = 'none'
    loginError = result || false
    if (!loginError) {
      invalidate('app:auth').catch(console.error)
    }
  }

  async function onRegister() {
    progress = 'register'
    const result = await handleLoginRegister('register')
    progress = 'none'
    registerError = result || false
  }

  async function handleLoginRegister(keyword: 'login' | 'register'): Promise<string> {
    try {
      const req = await fetch(`${PUBLIC_CONCINNITY_URL}/api/${keyword}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(keyword === 'login' ? login : register),
      })
      const res = (await req.json()) as { token: string; username: string; error?: string }
      if (!req.ok) {
        throw new Error(res.error ?? `Failed to ${keyword}! Error: ${req.statusText}`)
      }
      if (keyword === 'login') {
        localStorage.setItem('concinnity:token', res.token)
      }
      return ''
    } catch (e: unknown) {
      return e instanceof Error ? e.message : typeof e === 'string' ? e : `Failed to ${keyword}!`
    }
  }
</script>

<div class="container">
  <Box>
    <!-- TODO: Improve the design? -->
    <div>
      <h2>Login</h2>
      <br />
      <label for="login-username">E-mail / Username</label>
      <TextInput
        id="login-username"
        bind:value={login.username}
        error={!!loginError}
        disabled={progress !== 'none'}
        type="email"
        placeholder="e.g. aelia@retrixe.xyz"
      />
      <label for="login-password">Password</label>
      <TextInput
        id="login-password"
        bind:value={login.password}
        error={!!loginError}
        disabled={progress !== 'none'}
        type="password"
        onkeypress={e => e.key === 'Enter' && onLogin() /* eslint-disable-line */}
      />
      <br />
      {#if loginError === false}
        <p class="result">Logged in successfully! You should be redirected shortly...</p>
      {:else if !!loginError}
        <p class="result error">{loginError}</p>
      {/if}
      <br />
      <Button disabled={progress !== 'none'} onclick={onLogin}>Login</Button>
    </div>
    <div class="divider"></div>
    <div>
      <h2>Don't have an account?</h2>
      <br />
      <label for="register-username">Username</label>
      <TextInput
        id="register-username"
        bind:value={register.username}
        error={!!registerError}
        disabled={progress !== 'none'}
        type="email"
        placeholder="e.g. retrixe"
      />
      <label for="register-email">E-mail</label>
      <TextInput
        id="register-email"
        bind:value={register.email}
        error={!!registerError}
        disabled={progress !== 'none'}
        type="email"
        placeholder="e.g. aelia@retrixe.xyz"
      />
      <label for="register-password">Password</label>
      <TextInput
        id="register-password"
        bind:value={register.password}
        error={!!registerError}
        disabled={progress !== 'none'}
        type="password"
        onkeypress={e => e.key === 'Enter' && onRegister() /* eslint-disable-line */}
      />
      <br />
      {#if registerError === false}
        <p class="result">Registered successfully! Wait for your account to be verified.</p>
      {:else if !!registerError}
        <p class="result error">{registerError}</p>
      {/if}
      <br />
      <Button disabled={progress !== 'none'} onclick={onRegister}>Sign Up</Button>
    </div>
  </Box>
</div>

<style lang="scss">
  .result {
    max-width: 250px;
  }
  .error {
    color: var(--error-color);
  }

  label {
    padding: 0.5rem 0rem;
    display: block;
  }

  .divider {
    background-color: var(--divider-color);
    @media screen and (width < 768px) {
      height: 1px;
      margin: 2rem 0;
    }
    @media screen and (width >= 768px) {
      width: 1px;
      margin: 0 2rem;
    }
  }

  .container > :global(div) {
    display: flex;
    padding: 2rem;
    @media screen and (width < 768px) {
      flex-direction: column;
      margin-top: 2rem;
      padding: 1.5rem;
    }
  }

  .container {
    flex-grow: 1;
    display: flex;
    justify-content: center;
    @media screen and (width >= 768px) {
      align-items: center;
    }
    @media screen and (width < 768px) {
      align-items: flex-start;
    }
  }
</style>
