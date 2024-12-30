<script lang="ts">
  import Button from '../Button.svelte'

  interface Props {
    error: string | null
  }

  const { error }: Props = $props()
  // FIXME: Warning below the video if error or connecting
</script>

<!-- FIXME: This entire UI and props need changing up -->
<div class="video">
  {#if error}
    <div class="error">
      <h1>Error encountered!</h1>
      <h2>{error}</h2>
    </div>
  {:else}
    <div class="video-select">
      <h1>No video playing</h1>
      <br />
      <Button>Select local file</Button>
    </div>
  {/if}
</div>

<style lang="scss">
  .error {
    display: flex;
    flex-direction: column;
    text-align: center;
    gap: 0.5rem;

    h1 {
      color: var(--error-color);
    }

    h2 {
      font-weight: 300;
    }
  }

  .video {
    background-color: #000000;
    min-height: 280px; // FIXME: Remove this
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    color: white;
    @media screen and (min-width: 768px) {
      flex: 1;
    }
  }

  .video-select {
    display: flex;
    flex-direction: column;
    align-items: center;
  }
</style>
