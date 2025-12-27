<script lang="ts">
  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { onMount } from 'svelte'

  let error: string | null = $state(null)

  onMount(() => {
    try {
      // await ky.post(`api/register`, { json: register }).json<{ token: string; username: string }>()
      error = ''
      const timeout = setTimeout(() => goto(resolve('/login')), 5000)
      return () => clearTimeout(timeout)
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to reset password!`)
    }
  })
</script>

<h2>Verify Account</h2>
<div class="spacer"></div>
{#if error === ''}
  <p>
    Verified your account <b>eslyfail</b> successfully! Redirecting you to the
    <a href={resolve('/login')}>login page</a> in 5s...
  </p>
{:else if !!error}
  <p class="error">{error}</p>
{:else}
  <p>Please wait while we verify your account...</p>
{/if}
