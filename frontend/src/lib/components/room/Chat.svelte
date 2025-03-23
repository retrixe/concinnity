<script lang="ts">
  import { untrack } from 'svelte'
  import { Remarkable } from 'remarkable'
  import { linkify } from 'remarkable/linkify'
  import ky from '$lib/api/ky'
  import type { ChatMessage } from '$lib/api/room'
  import usernameCache from '$lib/state/usernameCache.svelte'
  import Textarea from '$lib/lunaria/Textarea.svelte'
  import TypingIndicator from './TypingIndicator.svelte'
  import type { SvelteMap } from 'svelte/reactivity'

  import joinWebm from '$lib/assets/join.webm'
  import leaveWebm from '$lib/assets/leave.webm'
  import messageWebm from '$lib/assets/message.webm'

  const systemUUID = '00000000-0000-0000-0000-000000000000'

  const soundEffects =
    typeof Audio === 'undefined'
      ? null
      : {
          join: new Audio(joinWebm),
          leave: new Audio(leaveWebm),
          message: new Audio(messageWebm),
        }

  const playNotificationSound = (chatMessage: ChatMessage, chatCooldown = false) => {
    if (chatMessage.userId === systemUUID) {
      const { message } = chatMessage

      if (message.endsWith('joined') || message.endsWith('reconnected')) {
        soundEffects?.join.play().catch(console.warn)
      } else if (message.endsWith('left') || message.endsWith('disconnected')) {
        soundEffects?.leave.play().catch(console.warn)
      }
    } else if (!chatCooldown) {
      soundEffects?.message.play().catch(console.warn)
    }
  }

  interface Props {
    typingIndicators: SvelteMap<string, [number, number]>
    disabled?: boolean
    messages: ChatMessage[]
    onSendMessage: (message: string) => void
    onTyping: () => void
  }

  const { typingIndicators, messages, disabled, onSendMessage, onTyping }: Props = $props()
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

  const typingUsers = $derived(Array.from(typingIndicators.keys()))

  // Fetch usernames for user IDs
  let prevId = 0
  $effect(() => {
    const userIds = messages
      .slice(prevId)
      .map(({ userId, message }) => (userId === systemUUID ? message.split(' ')[0] : userId))
      .concat(typingUsers)
      .reduce((set, userId) => {
        if (untrack(() => !usernameCache.has(userId)) /* Ignore changes to usernameCache */) {
          usernameCache.set(userId, null)
          set.add(userId)
        }
        return set
      }, new Set<string>())
    prevId = messages.length
    if (!userIds.size) return

    const query = userIds
      .values()
      .map(id => `id=${id}`)
      .reduce((acc, val) => `${acc}&${val}`)
    ky(`api/usernames?${query}`)
      .json<Record<string, string>>()
      .then(data => {
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
    if (!message.trim()) return
    onSendMessage(message.trim())
    message = ''
  }

  const handleTyping = () => onTyping()

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

  // Notification sounds
  let previousTimestamp = Number.NEGATIVE_INFINITY
  $effect(() => {
    if (!messages.length) return

    const currentTimestamp = new Date().getTime()
    const shouldNotify =
      messages.length && !document.hasFocus() && currentTimestamp - previousTimestamp > 5000

    playNotificationSound(messages[messages.length - 1], !shouldNotify)
    if (shouldNotify) previousTimestamp = currentTimestamp
  })

  const remarkable = new Remarkable('commonmark', { linkTarget: '_blank' }).use(linkify)
  remarkable.inline.ruler.enable('del')
</script>

<div class="chat">
  <div class="messages" bind:this={messagesEl}>
    <div class="spacer"></div>
    {#each messageGroups as messageGroup}
      {#if messageGroup.userId === systemUUID}
        <h5 style:text-align="center">
          {replaceLeadingUUID(messageGroup.messages[0])} — {parseTimestamp(messageGroup.timestamp)}
        </h5>
      {:else}
        <div>
          <h4>{getUsername(messageGroup.userId)} — {parseTimestamp(messageGroup.timestamp)}</h4>
          {#each messageGroup.messages as message}
            <div class="message-content">
              <!-- eslint-disable-next-line svelte/no-at-html-tags -->
              {@html remarkable.render(message)}
            </div>
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
    oninput={handleTyping}
    onkeypress={(e: KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        handleSendMessage()
      }
    }}
  />
  <TypingIndicator {typingUsers} />
</div>

<style lang="scss">
  .chat {
    min-height: 0; // Fixes chat overflowing out of parent
    padding: 0 1rem;
    display: flex;
    flex-direction: column;
    :global(textarea) {
      flex-shrink: 0;
      margin-bottom: 0rem;
      font-family: inherit;
      resize: none;
      width: 100%;
    }
    .spacer {
      height: 1rem;
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
    white-space: pre-line;
    overflow-y: scroll;
    margin-bottom: 1rem;
    h4,
    h5 {
      margin-top: 0.5rem;
    }
    .message-content {
      margin-top: 0.3rem;
      > :global(*) {
        display: inline;
      }
      :global(blockquote) {
        padding-left: 1rem;
        border-left: 4px solid gray;
      }
    }
  }
</style>
