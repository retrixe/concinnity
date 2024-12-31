<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import Button from '$lib/components/Button.svelte'

  const { username } = $derived(page.data)

  let status: string | null = $state(null)

  // TODO: Convert the button to a text field for custom IDs
  async function handleCreateRoom() {
    status = ''
    try {
      const req = await fetch(`${PUBLIC_BACKEND_URL}/api/room`, {
        method: 'POST',
        headers: { authorization: localStorage.getItem('concinnity:token') ?? '' },
        body: JSON.stringify({}),
      })
      const data = (await req.json()) as { error?: string; id: string }
      if (!req.ok) {
        throw new Error(data.error ?? req.statusText)
      }
      goto(`/room/${data.id}`).catch(console.error)
      status = null
    } catch (e) {
      status = e instanceof Error ? e.message : typeof e === 'string' ? e : 'Failed to create room!'
    }
  }
</script>

<div class="container">
  <div class="content">
    <h1>Get started</h1>
    <br />
    <p>
      Watch videos together with your friends using concinnity, a FOSS, lightweight and easy to use
      website built by a developer looking for something better.
    </p>
    <br />
    {#if username}
      <Button onclick={handleCreateRoom} disabled={status === ''}>Create a new room</Button>
    {:else}
      <a href="/login">
        <Button>Login / Sign Up</Button>
      </a>
    {/if}
    {#if !!status}
      <h4 style:color="var(--error-color)">{status}</h4>
    {/if}
  </div>
  <!-- TODO: Replace this image everywhere -->
  <img
    class="content"
    alt="A screenshot of the concinnity website"
    src="https://media.discordapp.net/attachments/588340346841464835/1321795849571008572/image.png?ex=67708410&is=676f3290&hm=7d04e84e556d48740664a0b5368009b0c21e73a4037b896a7920bbbb6cc7a0bf&=&format=webp&quality=lossless&width=1536&height=844"
  />
</div>

<style lang="scss">
  .content {
    padding: 1rem;
    @media screen and (min-width: 768px) {
      width: 45%;
      max-width: 640px;
    }
    p {
      font-size: 1.2rem;
    }
    :global(button) {
      font-size: 1rem;
    }
    h4 {
      padding-top: 1rem;
    }
  }

  .container {
    flex-grow: 1;
    display: flex;
    @media screen and (width < 768px) {
      flex-direction: column;
    }
    @media screen and (width >= 768px) {
      justify-content: center;
      align-items: center;
    }
  }
</style>
