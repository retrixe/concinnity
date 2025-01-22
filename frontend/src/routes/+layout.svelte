<script lang="ts">
  import { assets } from '$app/paths'
  import { page } from '$app/state'
  import type { Snippet } from 'svelte'
  import type { LayoutData } from './$types'
  import { invalidate, onNavigate } from '$app/navigation'
  import ky from '$lib/api/ky'
  import GitHubImage from '$lib/assets/GitHubImage.svelte'

  const { data, children }: { data: LayoutData; children: Snippet } = $props()
  const { username } = $derived(data)
  const { title, description, image, imageLarge, noIndex } = $derived(page.data)

  async function logout(event: Event) {
    event.preventDefault()
    try {
      await ky.post('api/logout')
      localStorage.removeItem('concinnity:token')
      invalidate('app:auth').catch(console.error)
    } catch (error) {
      console.error('Failed to logout!', error)
    }
  }

  onNavigate(navigation => {
    if (!('startViewTransition' in document)) return

    return new Promise(resolve => {
      document.startViewTransition(async () => {
        resolve()
        await navigation.complete
      })
    })
  })
</script>

<svelte:head>
  <title>{title}</title>
  <meta property="og:type" content="website" />
  <meta property="og:title" content={title} />
  <meta property="og:url" content={page.url.origin + page.url.pathname} />
  <meta property="og:image" content={image ?? assets + '/favicon.png'} />
  <meta property="og:description" content={description} />
  <meta name="twitter:title" content={title} />
  <meta name="twitter:card" content={imageLarge ? 'summary_large_image' : 'summary'} />
  <meta name="twitter:image:src" content={image ?? assets + '/favicon.png'} />
  <meta name="twitter:description" content={description} />
  <meta name="Description" content={description} />
  {#if noIndex}
    <meta name="robots" content="noindex,nofollow" />
  {/if}
</svelte:head>

<div class="top-bar">
  <h1>
    <a class="unstyled-link" href="/">concinnity</a>
  </h1>
  {#if username}
    <span>{username}</span>
    <div class="divider"></div>
    <a href="/" class="unstyled-link" onclick={logout}>Sign Out</a>
  {:else if page.url.pathname !== '/login'}
    <a href="/login" class="unstyled-link">Login</a>
    <div class="divider"></div>
    <a href="/login" class="unstyled-link">Sign Up</a>
    {#if page.url.pathname === '/'}
      <div class="divider"></div>
      <a
        href="https://github.com/retrixe/concinnity"
        target="_blank"
        rel="noopener noreferrer"
        class="unstyled-link"
      >
        <GitHubImage className="github-image" viewBox="0 0 98 96" height="28" width="28" />
      </a>
    {/if}
  {:else}
    <a href="/" class="unstyled-link">Home</a>
  {/if}
</div>

<div style:margin-top="4rem"></div>

{@render children()}

<style lang="scss">
  :global {
    * {
      margin: 0;
      box-sizing: border-box;
    }

    :root {
      --primary-color: #8f00ff;
      --error-color: #ff0042;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        color-scheme: dark;
        --link-color: #df73ff;
        --background-color: #111;
        --surface-color: #0e0e10; /* Jet black */
        --color: #ffffff;
        --divider-color: #666;
      }
      .github-image {
        filter: brightness(0) invert(1);
      }
    }

    @media (prefers-color-scheme: light) {
      :root {
        --link-color: #8f00ff;
        --background-color: #f5f5f5; /* White smoke */
        --surface-color: #f4f5fa; /* White solid */
        --color: #000000;
        --divider-color: #bbb;
      }
    }

    input {
      font: inherit;
    }
    select {
      font: inherit;
    }
    button {
      font: inherit;
    }
    textarea {
      font: inherit;
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
    z-index: 100;
    height: 4rem;
    position: fixed;
    display: flex;
    width: 100%;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    backdrop-filter: blur(20px);
    border-bottom: 1px solid var(--divider-color);
    h1 {
      font-size: 1.5rem;
      flex: 1;
    }
    span {
      font-weight: bold;
      color: var(--link-color);
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

  .divider {
    border-left: 1px solid var(--divider-color);
    height: 28px;
    margin: 0px 0.8rem;
  }
</style>
