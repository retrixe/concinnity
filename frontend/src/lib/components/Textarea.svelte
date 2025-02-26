<script lang="ts">
  import type { HTMLTextareaAttributes } from 'svelte/elements'

  interface Props extends HTMLTextareaAttributes {
    value?: string
    error?: boolean
  }
  let { value = $bindable(''), error, ...rest }: Props = $props()
</script>

<textarea bind:value class:error {...rest}></textarea>

<style lang="scss">
  textarea {
    width: 250px;
    padding: 0.6rem 1rem;
    margin-bottom: 0.6rem;
    border: 1px solid var(--divider-color);
    border-radius: 0.5rem;
    font-size: 1rem;
    transition: border 0.2s ease-in-out;
    &:focus {
      outline: 1px solid var(--primary-color);
    }
  }

  textarea.error {
    border: 1px solid var(--error-color);
  }
</style>
