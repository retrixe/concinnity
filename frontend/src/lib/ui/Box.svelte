<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { SvelteHTMLElements } from 'svelte/elements'

  const { children, ...rest }: { children: Snippet } & SvelteHTMLElements['div'] = $props()
</script>

<div {...rest}>
  {@render children()}
</div>

<style lang="scss">
  div {
    background-color: var(--surface-color);
    border: 1px solid var(--divider-color);
    border-radius: 8px;
    box-shadow: 0 0 1rem rgba(0, 0, 0, 0.2);
  }
</style>
