<script lang="ts">
  import { goto } from '$app/navigation'
  import { onMount } from 'svelte'

  let error: string | null = $state(null)

  onMount(() => {
    try {
      // await ky.post(`api/register`, { json: register }).json<{ token: string; username: string }>()
      error = ''
      const timeout = setTimeout(() => goto('/login'), 5000)
      return () => clearTimeout(timeout)
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : (e?.toString() ?? `Failed to reset password!`)
    }
  })
</script>

<h2>Verify Account</h2>
<div class="spacer"></div>
{#if error === ''}
  <p class="result left-align">
    Verified your account <b>eslyfail</b> successfully! Redirecting you to the
    <a href="/login">login page</a> in 5s...
  </p>
{:else if !!error}
  <p class="result error left-align">{error}</p>
{:else}
  <p class="left-align">Please wait while we verify your account...</p>
{/if}
