<script lang="ts">
  import { untrack } from 'svelte'
  import { PUBLIC_BACKEND_URL } from '$env/static/public'
  import type { ChatMessage } from '$lib/api/room'
  import usernameCache from '$lib/state/usernameCache.svelte'
  import Textarea from '../Textarea.svelte'

  const systemUUID = '00000000-0000-0000-0000-000000000000'

  interface Props {
    disabled?: boolean
    messages: ChatMessage[]
    onSendMessage: (message: string) => void
  }

  const { messages, onSendMessage, disabled }: Props = $props()

  type ChatMessageGroup = Omit<Omit<ChatMessage, 'message'>, 'id'> & { messages: string[] }
  const messageGroups = $derived(
    messages.reduce<ChatMessageGroup[]>((acc, { userId, timestamp, message }) => {
      const lastGroup = acc[acc.length - 1] as ChatMessageGroup | undefined
      if (lastGroup?.userId === userId && userId !== systemUUID) {
        lastGroup.timestamp = timestamp
        lastGroup.messages.push(message)
      } else acc.push({ userId, timestamp, messages: [message] })
      return acc
    }, []),
  )

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
  const replaceLeadingUUID = (message: string) => {
    const uuid = message.slice(0, message.indexOf(' '))
    return message.replace(uuid, getUsername(uuid))
  }
  const parseTimestamp = (timestamp: string) =>
    new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

  let message = $state('')
  const handleSendMessage = () => {
    onSendMessage(message.trim())
    message = ''
  }

  // Scroll to the bottom when messages are added
  // TODO (low): This doesn't interact well with Chrome fullscreen. Maybe use flex column-reverse there?
  let messagesEl = null as HTMLDivElement | null
  let isScrolledToBottom = $state(true)
  $effect.pre(() => {
    if (messages.length && messagesEl)
      isScrolledToBottom =
        messagesEl.scrollHeight - messagesEl.clientHeight <= messagesEl.scrollTop + 16
  })
  $effect(() => {
    if (messages.length && messagesEl && isScrolledToBottom)
      messagesEl.scrollTop = messagesEl.scrollHeight - messagesEl.clientHeight
  })
</script>

<div class="chat">
  <div class="messages" bind:this={messagesEl}>
    {#each messageGroups as messageGroup}
      {#if messageGroup.userId === systemUUID}
        <h5 style:text-align="center">
          {replaceLeadingUUID(messageGroup.messages[0])} — {parseTimestamp(messageGroup.timestamp)}
        </h5>
      {:else}
        <div>
          <h4>{getUsername(messageGroup.userId)} — {parseTimestamp(messageGroup.timestamp)}</h4>
          {#each messageGroup.messages as message}
            <p>{message.trim()}</p>
          {/each}
        </div>
      {/if}
    {/each}
  </div>
  <!-- prettier-ignore -->
  <Textarea
    {disabled}
    maxlength={2000}
    placeholder="Type message here..."
    bind:value={message}
    onkeypress={(e: KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        handleSendMessage()
      }
    }}
  />
</div>

<style lang="scss">
  .chat {
    min-height: 0; // Fixes chat overflowing out of parent
    padding: 1rem;
    display: flex;
    flex-direction: column;
    :global(textarea) {
      font-family: inherit;
      resize: none;
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
    word-wrap: break-word;
    white-space: pre-line;
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
