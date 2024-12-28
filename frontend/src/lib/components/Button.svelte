<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'

  type Props = HTMLButtonAttributes & { children: Snippet }
  const { children, ...rest }: Props = $props()
</script>

<button {...rest}>{@render children()}</button>

<style lang="scss">
  button {
    padding: 0.6rem 1.6rem;
    color: white;
    background-color: var(--primary-color);
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 1rem;
    transition: filter 0.2s ease-in-out;
    &:hover {
      filter: brightness(1.2);
    }
    &:active {
      filter: brightness(0.8);
    }
  }
</style>
