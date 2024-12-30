<script lang="ts">
  import Button from '../Button.svelte'

  interface Props {
    error: string | null
    loading: boolean
  }

  let { error, loading }: Props = $props()

  // FIXME: Button logic
</script>

<div class="video">
  <div class="video-select" class:error>
    {#if error}
      <h1>Error encountered!</h1>
      <h2>{error}</h2>
    {:else if loading}
      <!-- TODO: Loading spinner? -->
      <h1>Connecting to room...</h1>
    {:else}
      <h1>No video playing</h1>
      <br />
      <Button>Select local file</Button>
    {/if}
  </div>
</div>

<style lang="scss">
  .error {
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
    min-height: 280px;
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
