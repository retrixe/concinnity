<script lang="ts">
  import { page } from '$app/state'
  import Button from '$lib/components/Button.svelte'

  const { username } = $derived(page.data)
</script>

<div class="container">
  <div class="content">
    <h1>Get started</h1>
    <br />
    <p>
      Watch videos with your friends together using concinnity, a FOSS, lightweight and easy to use
      website built by a developer looking for something better.
    </p>
    <br />
    <!-- FIXME: Create room -->
    {#if username}
      <Button>Create a new room</Button>
    {:else}
      <a href="/login">
        <Button>Login / Sign Up</Button>
      </a>
    {/if}
  </div>
  <!-- FIXME: Image -->
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
      width: 40%;
      max-width: 640px;
    }
    p {
      font-size: 1.2rem;
    }
    :global(button) {
      font-size: 1rem;
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
