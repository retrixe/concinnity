<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'

  type Props = HTMLButtonAttributes & { children: Snippet }
  const { children, ...rest }: Props = $props()
</script>

<button {...rest}>{@render children()}</button>

<style lang="scss">
  button {
    display: inline-flex;
    justify-content: center;
    align-items: center;

    padding: 0.6rem 1.6rem;
    color: white;
    background-color: var(--primary-color);
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
    font-weight: bold;
    transition: filter 0.2s ease-in-out;
    &:enabled {
      &:hover {
        filter: brightness(1.2);
      }
      &:active {
        filter: brightness(0.8);
      }
    }
    &:disabled {
      background-color: var(--divider-color);
      cursor: not-allowed;
    }
  }
</style>
