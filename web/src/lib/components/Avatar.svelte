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

  let tooltipVisible = false;

	function showTooltip(event) {
		tooltipVisible = true;
	}

	function hideTooltip() {
		tooltipVisible = false;
	}
</script>

{#if email}
<span on:mouseleave={hideTooltip}>
  {#if user}
    <img src={user.Picture} alt={user.Name} class="avatar avatar-xs me-2 rounded" on:mousemove={showTooltip} />
    {#if option.showName}
      {user.Name}
    {/if}
  {:else}
    <svg on:mousemove={showTooltip}  xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-user-square" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 10a3 3 0 1 0 6 0a3 3 0 0 0 -6 0" /><path d="M6 21v-1a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v1" /><path d="M3 5a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2v-14z" /></svg>
  {/if}

  {#if tooltipVisible}
	<div class="user-info-card">
    <div class="card">
      <div class="row row-0">
        <div class="col-2">
          {#if user}
          <img src={user.Picture} class="avatar avatar-md object-cover card-img-start m-1" alt={user.Name}>
          {:else}
              <svg xmlns="http://www.w3.org/2000/svg" class="w-100 h-100 object-cover card-img-start icon icon-tabler icon-tabler-user-square" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 10a3 3 0 1 0 6 0a3 3 0 0 0 -6 0" /><path d="M6 21v-1a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v1" /><path d="M3 5a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2v-14z" /></svg>
          {/if}
        </div>
        <div class="col">
          <div class="card-body d-none d-xl-block ps-2">
            <div class="">{user?.Name}</div>
            <div class="mt-1 small text-secondary">{email}</div>
          </div>
        </div>
      </div>
    </div>
</div>
  {/if}
</span>
{:else}
  N/A
{/if}

<style>
	.user-info-card {
    position: absolute;
		border-radius: 5px;
		z-index: 1000;
    margin-top: 2px;
	}
  .card-body {
    padding: 3px;
    margin-left: 0.5em;
    margin-right: 0.5em;
    width: 100%
  }
</style>