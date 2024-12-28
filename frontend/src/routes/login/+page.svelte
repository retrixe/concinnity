<script lang="ts">
  import Box from '$lib/components/Box.svelte'
  import Button from '$lib/components/Button.svelte'
  import TextInput from '$lib/components/TextInput.svelte'
</script>

<div class="container">
  <Box>
    <div>
      <h2>Login</h2>
      <br />
      <TextInput type="email" placeholder="E-mail / Username" />
      <br />
      <TextInput type="password" placeholder="Password" />
      <br />
      <Button>Login</Button>
    </div>
    <div class="divider"></div>
    <div>
      <h2>Don't have an account?</h2>
      <br />
      <TextInput type="email" placeholder="Username" />
      <br />
      <TextInput type="email" placeholder="E-mail" />
      <br />
      <TextInput type="password" placeholder="Password" />
      <br />
      <Button>Sign Up</Button>
    </div>
  </Box>
</div>

<style lang="scss">
  .divider {
    background-color: var(--divider-color);
    @media screen and (width < 768px) {
      height: 1px;
      margin: 2rem 0;
    }
    @media screen and (width >= 768px) {
      width: 1px;
      margin: 0 2rem;
    }
  }

  .container > :global(div) {
    display: flex;
    padding: 2rem;
    @media screen and (width < 768px) {
      flex-direction: column;
      margin-top: 2rem;
      padding: 1.5rem;
    }
  }

  .container {
    flex-grow: 1;
    display: flex;
    justify-content: center;
    @media screen and (width >= 768px) {
      align-items: center;
    }
    @media screen and (width < 768px) {
      align-items: flex-start;
    }
  }
</style>
