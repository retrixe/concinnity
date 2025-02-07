<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import ky from '$lib/api/ky'
  import Box from '$lib/components/Box.svelte'
  import Button from '$lib/components/Button.svelte'
  import TextInput from '$lib/components/TextInput.svelte'

  let register = $state({ username: '', password: '', email: '' })
  let disabled = $state(false)
  let error: string | null = $state(null)

  const { username } = $derived(page.data)
  $effect(() => {
    if (username) goto('/').catch(console.error)
  })

  async function onRegister() {
    disabled = true
    try {
      await ky.post(`api/register`, { json: register }).json<{ token: string; username: string }>()
      error = ''
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to register!`)
    }
    disabled = false
  }
</script>

<div class="container">
  <Box>
    <h2>Register</h2>
    <br />
    <label for="register-username">Username</label>
    <TextInput
      id="register-username"
      bind:value={register.username}
      error={!!error}
      {disabled}
      type="email"
      placeholder="e.g. retrixe"
    />
    <label for="register-email">E-mail</label>
    <TextInput
      id="register-email"
      bind:value={register.email}
      error={!!error}
      {disabled}
      type="email"
      placeholder="e.g. aelia@retrixe.xyz"
    />
    <label for="register-password">Password</label>
    <TextInput
      id="register-password"
      bind:value={register.password}
      error={!!error}
      {disabled}
      type="password"
      onkeypress={e => e.key === 'Enter' && onRegister() /* eslint-disable-line */}
    />
    {#if error === ''}
      <p class="result">Registered successfully! Wait for your account to be verified.</p>
    {:else if !!error}
      <p class="result error">{error}</p>
    {/if}
    <br />
    <Button {disabled} onclick={onRegister}>Sign Up</Button>
    <br />
    <p>Already have an account? <a href="/login">Log in</a></p>
  </Box>
</div>

<style lang="scss">
  .error {
    color: var(--error-color);
  }

  label {
    padding: 0.5rem 0rem;
    font-weight: bold;
  }

  p {
    align-self: center;
  }

  .container > :global(div) {
    display: flex;
    flex-direction: column;
    padding: 1.5rem;
    margin: 1.5rem;
    width: 100%;
    max-width: 400px;
  }

  .container {
    flex-grow: 1;
    display: flex;
    justify-content: center;
    align-items: center;
  }
</style>
