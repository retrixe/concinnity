<script lang="ts">
  import TextInput from '$lib/components/TextInput.svelte'

  // TODO: Timestamp design needs improvement to account for latest messages
  // FIXME: User IDs need to be replaced with usernames fetched from the server
  interface Props {
    disabled?: boolean
    messages: { userId: string; message: string; timestamp: string }[]
    onSendMessage: (message: string) => void
  }

  const { messages, onSendMessage, disabled }: Props = $props()

  const parseTimestamp = (timestamp: string) =>
    new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

  let message = $state('')
  const handleSendMessage = () => {
    onSendMessage(message)
    message = ''
  }
</script>

<div class="chat">
  <div class="messages">
    {#each messages as message, i}
      <div>
        {#if i === 0 || messages[i - 1].userId !== message.userId}
          <h4>{message.userId} — {parseTimestamp(message.timestamp)}</h4>
        {/if}
        <p>{message.message}</p>
      </div>
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
    padding: 1rem;
    display: flex;
    flex-direction: column;
    @media screen and (width < 768px) {
      flex: 1;
    }
    @media screen and (min-width: 768px) {
      width: 280px;
    }
  }

  .messages {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: end;
    overflow-y: auto;
    margin-bottom: 1rem;
    h4 {
      margin-top: 0.5rem;
    }
    p {
      margin-top: 0.3rem;
    }
  }
</style>
