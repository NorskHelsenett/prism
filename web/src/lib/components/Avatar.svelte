<script>
  import { Fetch } from '$lib/fetchUtil'
  export let email = '';
  export let option = { showName: true}

  let user = null;

  async function getUser(email) {
    return await Fetch(`/api/userinfo/${email}`);
  }

  $: if (email) {
    getUser(email).then(u => user = u)
  }
</script>

{#if email}
  {#if user}
    <img src={user.Picture} alt={user.Name} title={user.Name} class="avatar avatar-xs me-2 rounded"/>
    {#if option.showName}
      {user.Name}
    {/if}
  {:else}
    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-user-square" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 10a3 3 0 1 0 6 0a3 3 0 0 0 -6 0" /><path d="M6 21v-1a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v1" /><path d="M3 5a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2v-14z" /></svg>
  {/if}
{:else}
  N/A
{/if}
