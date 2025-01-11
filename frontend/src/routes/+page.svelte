<script lang="ts">
  import type { FormEventHandler } from 'svelte/elements'
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import Button from '$lib/components/Button.svelte'
  import TextInput from '$lib/components/TextInput.svelte'

  const { username } = $derived(page.data)

  let status: string | null = $state(null)
  let roomId = $state('')

  const onRoomIdChange: FormEventHandler<HTMLInputElement> = e => {
    const inputEl = e.target as HTMLInputElement
    if (/^[a-zA-Z0-9_-]{0,24}$/.test(inputEl.value)) roomId = inputEl.value
    else inputEl.value = roomId
  }

  async function handleCreateRoom() {
    status = ''
    try {
      const req = await fetch(`${PUBLIC_BACKEND_URL}/api/room`, {
        method: 'POST',
        headers: { authorization: localStorage.getItem('concinnity:token') ?? '' },
        body: JSON.stringify(roomId ? { id: roomId } : {}),
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
      <TextInput
        value={roomId}
        oninput={onRoomIdChange}
        autocapitalize="off"
        autocomplete="off"
        placeholder="Enter custom name (optional)"
      />
      <Button onclick={handleCreateRoom} disabled={status === ''}>Create room</Button>
    {:else}
      <a href="/login">
        <Button>Login / Sign Up</Button>
      </a>
    {/if}
    {#if !!status}
      <h4 style:color="var(--error-color)">{status}</h4>
    {/if}
  </div>
  <picture>
    <source
      type="image/webp"
      srcset="https://f002.backblazeb2.com/file/mythic-storage-public/demo-dark.webp"
      media="(prefers-color-scheme: dark)"
    />
    <source
      type="image/webp"
      srcset="https://f002.backblazeb2.com/file/mythic-storage-public/demo-light.webp"
    />
    <img
      class="content"
      alt="A screenshot of the concinnity website"
      src="https://f002.backblazeb2.com/file/mythic-storage-public/demo-dark.webp"
    />
  </picture>
</div>

<style lang="scss">
  img {
    border-radius: 0.5rem;
    filter: drop-shadow(0 0 1rem var(--primary-color));
  }

  picture {
    display: contents;
  }

  .content {
    margin: 1rem;
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
