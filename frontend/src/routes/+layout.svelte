<script lang="ts">
  import { assets } from '$app/paths'
  import { page } from '$app/state'
  import type { Snippet } from 'svelte'
  import type { LayoutData } from './$types'
  import { invalidate, onNavigate } from '$app/navigation'
  import ky from '$lib/api/ky'
  import GitHubImage from '$lib/assets/GitHubImage.svelte'
  import 'heliodor/Baseline.scss'
  import { TopBar, TopBarDivider, TopBarTitle, TopBarLink } from 'heliodor'

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
      alert('Failed to logout!')
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

<TopBar>
  <TopBarTitle>
    <TopBarLink href="/">concinnity</TopBarLink>
  </TopBarTitle>
  {#if username}
    <TopBarLink highlighted href="/settings">{username}</TopBarLink>
    <TopBarDivider />
    <TopBarLink href="/" onclick={logout}>Sign Out</TopBarLink>
  {:else if page.url.pathname !== '/login' && page.url.pathname !== '/register'}
    <TopBarLink href="/login">Login</TopBarLink>
    <TopBarDivider />
    <TopBarLink href="/register">Sign Up</TopBarLink>
    {#if page.url.pathname === '/'}
      <TopBarDivider />
      <TopBarLink
        href="https://github.com/retrixe/concinnity"
        target="_blank"
        rel="noopener noreferrer"
      >
        <GitHubImage
          className="github-image"
          viewBox="0 0 98 96"
          height="1.75rem"
          width="1.75rem"
        />
      </TopBarLink>
    {/if}
  {:else}
    <TopBarLink href="/">Home</TopBarLink>
  {/if}
</TopBar>

{@render children()}

<style lang="scss">
  :global {
    :root {
      --primary-color: #8f00ff;
      --error-color: #ff0042;

      --link-color: #8f00ff;
      --background-color: #f5f5f5; /* White smoke */
      --surface-color: #fcfcfc; /* White smoke but brighter */
      --color: #000000;
      --divider-color: #bbb;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        --link-color: #df73ff;
        --background-color: #0e0e10; /* Jet black */
        --surface-color: #1b1b1b; /* Eerie black */
        --color: #ffffff;
        --divider-color: #666;
      }
      .github-image {
        filter: brightness(0) invert(1);
      }
    }
  }
</style>
