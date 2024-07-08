<script >
	import { Fetch } from '$lib/fetchUtil';
	import { onMount } from 'svelte';
  import { fade } from 'svelte/transition'
	import { toast } from 'svelte-sonner';
	import { formatDateToYYYYMMDD } from '$lib/utils';
// @ts-check
  /**
   * @typedef {import('$lib/models/apikey').APIKey} APIKey
   */

  /** @type {APIKey[]} */
  let apiKeys = [];
	let showModal = false;
  let isPersisting = false
  let newApiKeyName = ""
  let newApiKeySecret = ""
  let loading = true

  async function closeModal() {
		let success = true
		if (success){
      showModal = false;
		}
	}

  async function handleSubmit() {
    if (!newApiKeyName || newApiKeyName == "") {
      return
    }
    isPersisting = true
    const result = await Fetch("/api/profile/apikey", {method:"POST", body: JSON.stringify({"name": newApiKeyName})})
    if (result.Error){
      console.log("Error sending")
    } else {
      newApiKeyName = ""
      if(result.apikey != null){
        newApiKeySecret = result.apikey
      }
      apiKeys.push(result)
      isPersisting = false
    }
  }

  async function handleCopyContent() {
    try {
      await navigator.clipboard.writeText(newApiKeySecret);
      showModal = false;
      newApiKeyName = ""
      newApiKeySecret = ""
      toast.success('API key copied to clipboard');
    } catch (err) {
      console.error('Failed to copy: ', err);
      toast.error('Failed to copy API key');
    }
  }

  onMount(async () => {
    apiKeys = await Fetch("/api/profile/apikey")
    loading = false
  })

  async function deleteAPIKey(id) {
    const result = await Fetch(`/api/profile/apikey/${id}`, {method: "DELETE"})
    if (result?.Error) {
        toast.error("Unable to delete apikey")
    } else {
      apiKeys = apiKeys.filter(apiKey => apiKey.ID !== id);
      toast.success("API Key deleted")
    }
  }

  let isHovering = false;
  function mouseOver() {
    isHovering = true;
  }
  function mouseOut() {
    isHovering = false;
  }
</script>

{#if !loading}
<div class="card-body">
  <h2 class="card-title">API Key Management</h2>
  <div class="text-secondary">
    <p>API keys are unique identifiers for authenticating users or applications to an API, meant to be kept secure and not exposed publicly. They help control API usage and prevent abuse. Keys should be monitored, rotated for security, and stored safely. Exposed or compromised keys must be revoked and replaced immediately. Audit API key usage regularly to ensure compliance with security policies.</p>
  </div>
  <div class="row">
    <div class="col-9"></div>
    <div class="col-3 d-flex flex-row justify-content-end">
      <a class="btn btn-link text-azure bg-transparent mb-2 text-decoration-none" on:click={()=> {showModal = true}} on:mouseover={mouseOver}
        on:mouseout={mouseOut}
        class:bg-azure-lt={isHovering}>
      Create new API Key</a></div>
  </div>
  {#if apiKeys.length > 0}
  <div class="row">
    <table class="table table-vcenter card-table table-striped">
      <thead>
        <tr>
          <th>Name</th>
          <!-- <th>Created</th> -->
          <th class="w-1">Last Used</th>
          <th class="w-1">Expire On</th>
          <th class="w-1"></th>
        </tr>
      </thead>
      <tbody>
        {#each apiKeys as apikey}
          <tr transition:fade={{ delay: 10, duration: 300 }}>
            <td>{apikey.name}</td>
            <!-- <td>{formatDateToYYYYMMDD(apikey.CreatedAt)}</td> -->
            <td>{formatDateToYYYYMMDD(apikey.UpdatedAt)}</td>
            <td>{formatDateToYYYYMMDD(apikey.expire)}</td>
            <td>
              <a class="btn btn-sm btn-pill text-red s-oeZ2237gM52l" on:click={deleteAPIKey(apikey.ID)}>
              delete</a>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  {:else}
  <div class="card-body text-center py-4 p-sm-5">
    <img src="/images/apikey_factory.png" class="img mb-n2" alt="apikey factory pipeline"/>
    <h1 class="mt-5 text-azure">Your api keys lives here!</h1>
    <p class="text-secondary m-0">You have not created any API keys, yet!</p>
  </div>
  {/if}
</div>
{/if}

{#if showModal}
<div class="modal modal-blur fade show" id="modal-small" tabindex="-1" role="dialog" aria-modal="true" style="display: block;">
  <div class="modal-dialog modal-sm modal-dialog-centered" role="document">
    <div class="modal-content">
      <div class="modal-body">
        {#if newApiKeySecret == ""}
        <div class="modal-title">Create new API key</div>
        <div>
          <div class="col">
            <label class="form-label" for="new-apikey-name">Name</label>
            <input id="new-apikey-name" type="text" bind:value={newApiKeyName} class="form-control">
          </div>
        </div>
          {:else}
          <div class="col">
            <div class="input-group input-group-flat">

              <input type="text" bind:value={newApiKeySecret} class="form-control disabled text-truncate" autocomplete="off">
              <span class="input-group-text">
                <a title="Copy apikey to clipboard" class="link-secondary cursor-pointer" data-bs-toggle="tooltip" aria-label="Copy api key" data-bs-original-title="Copy api key" on:click|preventDefault={handleCopyContent}>
                  <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-copy" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="#ffffff" fill="none" stroke-linecap="round" stroke-linejoin="round">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                    <path d="M7 7m0 2.667a2.667 2.667 0 0 1 2.667 -2.667h8.666a2.667 2.667 0 0 1 2.667 2.667v8.666a2.667 2.667 0 0 1 -2.667 2.667h-8.666a2.667 2.667 0 0 1 -2.667 -2.667z" />
                    <path d="M4.012 16.737a2.005 2.005 0 0 1 -1.012 -1.737v-10c0 -1.1 .9 -2 2 -2h10c.75 0 1.158 .385 1.5 1" />
                  </svg>
                </a>
              </span>
            </div>
          </div>

          {/if}
      </div>
      {#if newApiKeySecret == ""}
      <div class="modal-footer">
        <button type="button" class="btn btn-link link-secondary me-auto" data-bs-dismiss="modal" on:click|preventDefault={() => showModal = false}>Cancel</button>
        <button type="button" disabled={isPersisting} class="btn btn-primary" data-bs-dismiss="modal" on:click|preventDefault={handleSubmit}>Create API key</button>
      </div>
      {/if}
    </div>
  </div>
</div>
{/if}


<style>
  .img{
    height: 256px;
    width: 256px;
  }
</style>