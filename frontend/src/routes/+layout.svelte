<script lang="ts">
  import { assets } from '$app/paths'
  import { page } from '$app/state'
  import type { Snippet } from 'svelte'

  const { title, description, image, noIndex } = page.data
  const { children }: { children: Snippet } = $props()
</script>

<svelte:head>
  <title>{title ?? 'concinnity'}</title>
  <meta property="og:type" content="website" />
  <meta property="og:title" content={title} />
  <meta property="og:url" content={page.url.origin + page.url.pathname} />
  <meta property="og:image" content={image ?? assets + '/favicon.png'} />
  <meta property="og:description" content={description} />
  <meta name="Description" content={description} />
  {#if noIndex}
    <meta name="robots" content="noindex,nofollow" />
  {/if}
</svelte:head>

<div class="top-bar">
  <h1>
    <a class="unstyled-link" href="/">concinnity</a>
  </h1>
  <!-- FIXME: Improve the design of this -->
  <a href="/login">Sign In</a>
</div>

<div style:margin-top="4rem"></div>

{@render children()}

<style lang="scss">
  :global {
    * {
      margin: 0;
    }

    :root {
      --primary-color: #8f00ff;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        color-scheme: dark;
        --link-color: #df73ff;
        --background-color: #111;
        --color: #ffffff;
        --divider-color: #666;
      }
    }

    @media (prefers-color-scheme: light) {
      :root {
        --link-color: #8f00ff;
        --background-color: #eee;
        --color: #000000;
        --divider-color: #bbb;
      }
    }

    body {
      font-family: system-ui, 'Segoe UI', Roboto, Oxygen-Sans, Ubuntu, Cantarell, 'Helvetica Neue',
        Helvetica, Arial, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol';
      background-color: var(--background-color);
      color: var(--color);
      width: 100vw;
      max-width: 100%;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      a {
        color: var(--link-color);
      }
    }
  }

  .top-bar {
    position: fixed;
    display: flex;
    width: 100%;
    box-sizing: border-box;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    backdrop-filter: blur(20px);
    border-bottom: 1px solid var(--divider-color);
    h1 {
      font-size: 1.5rem;
    }
    a {
      font-weight: bold;
      text-decoration: none;
    }
  }

  .unstyled-link {
    color: inherit;
    text-decoration: none;
  }
</style>
