<script lang="ts">
  interface Props {
    typingUsers: string[]
  }

  const { typingUsers }: Props = $props()
</script>

<div class="typing-indicator">
  {#if typingUsers.length >= 3}
    <p>Multiple users are typing...</p>
  {:else if typingUsers.length === 2}
    <p>{`${typingUsers[0]} and ${typingUsers[1]} are typing...`}</p>
  {:else if typingUsers.length === 0}
    <p></p>
  {:else}
    <p>{`${typingUsers[0]} is typing...`}</p>
  {/if}
</div>

<style lang="scss">
  .typing-indicator {
    font-size: 1rem;
  }
</style>
