<script lang="ts">
  interface Props {
    typingUsers: string[]
  }

  const { typingUsers }: Props = $props()
</script>

<p class="typing-indicator">
  {#if typingUsers.length >= 3}
    Multiple users are typing...
  {:else if typingUsers.length === 2}
    <b>{typingUsers[0]}</b> and <b>{typingUsers[1]}</b> are typing...
  {:else if typingUsers.length === 1}
    <b>{typingUsers[0]}</b> is typing...
  {/if}
</p>

<style lang="scss">
  .typing-indicator {
    /* Just in case */
    font-size: 1rem;
    line-height: 1.2;
    height: 1.2rem;
    margin: 0.4rem 0;
    white-space: nowrap;
    overflow: hidden;
  }
</style>
