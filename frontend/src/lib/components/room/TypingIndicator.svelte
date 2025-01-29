<script lang="ts">
  interface Props {
    showTypingUsers: string[]
  }

  const { showTypingUsers }: Props = $props()
</script>

<div class="typing-indicator">
  {#if showTypingUsers.length >= 3}
    <p>Multiple users are typing...</p>
  {:else if showTypingUsers.length === 2}
    <p>{`${showTypingUsers[0]} and ${showTypingUsers[1]} are typing...`}</p>
  {:else if showTypingUsers.length === 0}
    <p></p>
  {:else}
    <p>{`${showTypingUsers[0]} is typing...`}</p>
  {/if}
</div>

<style lang="scss">
  .typing-indicator {
    font-size: 1rem;
  }
</style>
