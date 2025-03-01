<script lang="ts">
  import type { Snippet } from 'svelte'

  const { children, text }: { children: Snippet; text: string } = $props()
</script>

<div class="tooltip">
  <div class="tooltip-text">
    {text}
  </div>
  {@render children()}
</div>

<style lang="scss">
  .tooltip {
    position: relative;
    display: inline-block;
    &:hover .tooltip-text {
      opacity: 100;
      visibility: visible;
    }
  }

  .tooltip-text {
    position: absolute;
    /* Positioning */
    bottom: 80%;
    left: 50%;
    transform: translate(-50%, 0);
    z-index: 50;

    color: var(--color);
    background-color: var(--background-color);
    padding: 0.5rem;
    border-radius: 8px;
    box-shadow: 0 0 16px var(--primary-color);
    pointer-events: none;
    text-align: center;

    opacity: 0;
    visibility: hidden;
    transition:
      opacity 0.2s ease-in-out,
      visibility 0.2s ease-in-out;
  }
</style>
