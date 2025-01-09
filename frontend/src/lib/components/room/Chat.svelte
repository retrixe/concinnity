<script lang="ts">
  import { untrack } from 'svelte'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import type { ChatMessage } from '$lib/api/room'
  import TextInput from '$lib/components/TextInput.svelte'
  import usernameCache from '$lib/state/usernameCache.svelte'

  const systemUUID = '00000000-0000-0000-0000-000000000000'

  // TODO (high): Timestamp design needs improvement to account for latest messages
  interface Props {
    disabled?: boolean
    messages: ChatMessage[]
    onSendMessage: (message: string) => void
  }

  const { messages, onSendMessage, disabled }: Props = $props()

  // Fetch usernames for user IDs
  let prevId = 0
  $effect(() => {
    const userIds = messages.slice(prevId).reduce((set, message) => {
      const userId = message.userId === systemUUID ? message.message.split(' ')[0] : message.userId
      if (untrack(() => !usernameCache.has(userId)) /* Ignore changes to usernameCache */) {
        usernameCache.set(userId, null)
        set.add(userId)
      }
      return set
    }, new Set<string>())
    prevId = messages.length
    if (!userIds.size) return

    const authorization = localStorage.getItem('concinnity:token') ?? ''
    const query = userIds
      .values()
      .map(id => `id=${id}`)
      .reduce((acc, val) => `${acc}&${val}`)
    fetch(`${PUBLIC_BACKEND_URL}/api/usernames?${query}`, { headers: { authorization } })
      .then(res => res.json())
      .then((data: Record<string, string>) => {
        for (const [userId, username] of Object.entries(data)) usernameCache.set(userId, username)
      })
      .catch((e: unknown) => console.error('Failed to retrieve usernames!', e))
  })
  const getUsername = (userId: string) => usernameCache.get(userId) ?? userId.split('-')[0] // UUID
  const parseTimestamp = (timestamp: string) =>
    new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

  let message = $state('')
  const handleSendMessage = () => {
    onSendMessage(message)
    message = ''
  }

  // Scroll to the bottom when messages are added
  let messagesEl: HTMLDivElement | null = null
  let isScrolledToBottom = $state(true)
  $effect.pre(() => {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- false positive
    if (messages.length && messagesEl)
      isScrolledToBottom =
        messagesEl.scrollHeight - messagesEl.clientHeight <= (messagesEl.scrollTop as number) + 1
  })
  $effect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- false positive
    if (messages.length && messagesEl && isScrolledToBottom)
      messagesEl.scrollTop = messagesEl.scrollHeight - messagesEl.clientHeight
  })
</script>

<div class="chat">
  <div class="messages" bind:this={messagesEl}>
    {#each messages as message, i}
      {#if message.userId === systemUUID}
        <h5 style:text-align="center">
          {message.message.replace(
            message.message.split(' ')[0],
            getUsername(message.message.split(' ')[0]),
          )} — {parseTimestamp(message.timestamp)}
        </h5>
      {:else}
        <div>
          {#if i === 0 || messages[i - 1].userId !== message.userId}
            <h4>{getUsername(message.userId)} — {parseTimestamp(message.timestamp)}</h4>
          {/if}
          <p>{message.message}</p>
        </div>
      {/if}
    {/each}
  </div>
  <!-- prettier-ignore -->
  <TextInput
    {disabled}
    placeholder="Type message here..."
    bind:value={message}
    onkeypress={e => e.key === 'Enter' && handleSendMessage() /* eslint-disable-line */}
  />
</div>

<style lang="scss">
  .chat {
    min-height: 0; // Fixes chat overflowing out of parent
    padding: 1rem;
    display: flex;
    flex-direction: column;
    :global(input) {
      width: 100%;
    }
    @media screen and (width < 768px) {
      flex: 1;
    }
    @media screen and (min-width: 768px) {
      width: 320px;
    }
  }

  .messages {
    flex: 1;
    overflow-y: scroll;
    margin-bottom: 1rem;
    h4,
    h5 {
      margin-top: 0.5rem;
    }
    p {
      margin-top: 0.3rem;
    }
  }
</style>
