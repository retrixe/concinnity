<script lang="ts">
/**
 * Spacer.svelte
 * A simple spacer component for vertical or horizontal spacing.
 * Usage: <Spacer size="1rem" axis="vertical" />
 */
export let size: string = "1rem";
export let axis: "vertical" | "horizontal" = "vertical";
</script>

<div aria-hidden="true"
  style={axis === "vertical"
    ? `height: ${size}; width: 100%;`
    : `width: ${size}; height: 100%; display: inline-block;`}
></div>
