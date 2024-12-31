<script lang="ts">
  import { assets } from '$app/paths'
  import { page } from '$app/state'
  import type { Snippet } from 'svelte'
  import type { LayoutData, PageData } from './$types'
  import { invalidate } from '$app/navigation'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'

  const { data, children }: { data: LayoutData; children: Snippet } = $props()
  const { username } = $derived(data)
  const { title, description, image, noIndex } = $derived(page.data) as Omit<PageData, 'image'> & {
    image?: string
    noIndex?: boolean
  }

  async function logout(event: Event) {
    event.preventDefault()
    try {
      const req = await fetch(`${PUBLIC_BACKEND_URL}/api/logout`, {
        method: 'POST',
        headers: { authorization: localStorage.getItem('concinnity:token') ?? '' },
      })
      if (!req.ok) {
        const error = (await req.json()) as { error?: string }
        throw new Error(error.error ?? req.statusText)
      }
      localStorage.removeItem('concinnity:token')
      invalidate('app:auth').catch(console.error)
    } catch (error) {
      console.error('Failed to logout!', error)
    }
  }
</script>

<svelte:head>
  <title>{title}</title>
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
  {#if username}
    <span>{username}</span>
    <div class="divider"></div>
    <a href="/" class="unstyled-link" onclick={logout}>Sign Out</a>
  {:else if page.url.pathname !== '/login'}
    <a href="/login" class="unstyled-link">Login</a>
    <div class="divider"></div>
    <a href="/login" class="unstyled-link">Sign Up</a>
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
