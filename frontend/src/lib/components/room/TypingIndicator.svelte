<script lang="ts">
  import userProfileCache from '$lib/state/userProfileCache.svelte'

  const { typingUsers }: { typingUsers: string[] } = $props()

  const getUsername = (userId: string) =>
    userProfileCache.get(userId)?.username ?? userId.split('-')[0] // UUID
</script>

<p class="typing-indicator">
  {#if typingUsers.length >= 3}
    Multiple users are typing...
  {:else if typingUsers.length === 2}
    <b>{getUsername(typingUsers[0])}</b> and <b>{getUsername(typingUsers[1])}</b> are typing...
  {:else if typingUsers.length === 1}
    <b>{getUsername(typingUsers[0])}</b> is typing...
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
    flex-shrink: 0;
  }
</style>
