<script>
  import { page } from '$app/state'
</script>

<div>
  <h1>{page.status}</h1>
  <p>{page.error?.message}</p>
</div>

<style lang="scss">
  div {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex: 1;
  }
</style>
