<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'

  type Props = HTMLButtonAttributes & { children: Snippet }
  const { children, ...rest }: Props = $props()
</script>

<button {...rest}>{@render children()}</button>

<style lang="scss">
  button {
    /* Inherits considerable styling from Button */
    display: inline-flex;
    justify-content: flex-start;
    align-items: center;

    padding: 8px;
    background-color: inherit;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: filter 0.2s ease-in-out;
    &:enabled {
      &:hover {
        filter: brightness(0.8);
      }
      &:active {
        filter: brightness(0.6);
      }
    }
    &:disabled {
      background-color: var(--divider-color);
      cursor: not-allowed;
    }
  }
</style>
