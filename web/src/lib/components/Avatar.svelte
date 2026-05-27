<script>
  import { run } from 'svelte/legacy';

  import { onMount } from 'svelte';
  import { userStore } from '$lib/stores/userStore';

  /**
   * @typedef {Object} Props
   * @property {string} [email]
   * @property {any} [option]
   */

  /** @type {Props} */
  let { email = '', option = $bindable({}) } = $props();

  const defaultOption = {
    showName: true,
    size: "xs",
    emptyFields: false,
    circle: false,
    tooltipEnabled: true
  };

  onMount(() => {
    option = { ...defaultOption, ...option };
  });

  let user = $state(null);

  run(() => {
    if (email) {
      userStore.getUser(email);
      userStore.subscribe(store => {
        user = store[email];
      });
    }
  });

  let tooltipVisible = $state(false);

  function showTooltip(event) {
    if (option.tooltipEnabled) {
      tooltipVisible = true;
    }
  }

  function hideTooltip() {
    tooltipVisible = false;
  }

  function getSize() {
    if (option.size === "sm") return "avatar-sm";
    if (option.size === "md") return "avatar-md";
    if (option.size === "lg") return "avatar-lg";
    return "avatar-xs";
  }

  function calculateInitials(fullname) {
    const words = fullname.split(' ');
    return `${words[0][0]}${words[words.length - 1][0]}`;
  }

  function calculateInitialsFromEmail(email) {
    const mail = email.split("@")[0];
    const name = mail.split(".");
    return calculateInitials(`${name[0]} ${name[name.length-1]}`);
  }
</script>

{#if email}
<!-- svelte-ignore a11y_no_static_element_interactions -->
<span onmouseleave={hideTooltip} class="flex items-center">
  {#if user}
    {#if user.picture}
      <img src={user.picture} alt={user.name} class:rounded-circle="{option.circle}" class="avatar {getSize()} me-2 rounded" onmousemove={showTooltip} />
    {:else}
      <span class="avatar text-uppercase {getSize()}" class:rounded-circle="{option.circle}" onmousemove={showTooltip}>{calculateInitials(user.name)}</span>
    {/if}
    {#if option.showName}
      <span class="align-middle">{user.name}</span>
    {/if}
  {:else}
    <span class="avatar text-uppercase {getSize()}" class:rounded-circle="{option.circle}" onmousemove={showTooltip}>{calculateInitialsFromEmail(email)}</span>
  {/if}

  {#if tooltipVisible}
	<div class="user-info-card">
    <div class="card">
      <div class="row row-0">
        <div class="col-auto d-flex align-items-center ps-1">
          {#if user}
          <img loading="lazy" src={user.picture} class="avatar avatar-sm rounded-circle" alt={user.name}>
          {:else}
              <svg xmlns="http://www.w3.org/2000/svg" class="w-100 h-100 object-cover card-img-start icon icon-tabler icon-tabler-user-square" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 10a3 3 0 1 0 6 0a3 3 0 0 0 -6 0" /><path d="M6 21v-1a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v1" /><path d="M3 5a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2v-14z" /></svg>
          {/if}
        </div>
        <div class="col">
          <div class="card-body d-none d-xl-block ps-2">
            <div class="">{user?.name}</div>
            <div class="mt-1 small text-secondary">{email}</div>
          </div>
        </div>
      </div>
    </div>
</div>
  {/if}
</span>
{:else}
  {#if !option.emptyFields}
  N/A
  {/if}
{/if}

<style>

  .flex {
    display: inline-flex;
    gap:12px;
    vertical-align: middle;
  }
  .items-center {
    align-items: center;
  }

  img {
    margin: 0 !important;
  }
	.user-info-card {
    position: absolute;
		border-radius: 9999px;
		z-index: 1000;
    bottom: 100%;
    margin-bottom: 4px;
    margin-left: -20px;
    overflow: hidden;
	}
  .user-info-card .card {
    border-radius: 9999px;
  }
  .card-body {
    padding: 3px;
    margin-left: 0.5em;
    margin-right: 1.5em;
    width: 100%
  }
</style>