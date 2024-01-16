<script>
  import { onMount } from 'svelte';
  import { Fetch } from '$lib/fetchUtil.js';
	import DebouncedInput from '../DebouncedInput.svelte';

  let user;
  let persisting = false

    onMount(async () => {
      user = await Fetch(`/api/user`);
    });

    async function handleTitleChange(newVal) {
      persisting = true
      user.title = newVal.detail
      const response = await Fetch("/api/user", {method: "PUT", body: JSON.stringify(user)})
      persisting = false
    }
</script>

{#if user}
<div class="card-body">
  <h2 class="mb-4">My Account</h2>
  <h3 class="card-title">Profile Details</h3>
  <div class="row align-items-center">
    <div class="col-auto">
      <span class="avatar avatar-xl" style="background-image: url({user.Picture})"></span>
    </div>
    <div class="col m-2">
      <strong class="row">Name</strong>
      <div class="row text-secondary">{user.Name}</div>
      <strong class="row">Email</strong>
      <a href="mailto:{user.Email}" class="row">{user.Email}</a>
    </div>
  </div>
  <h3 class="card-title mt-4">Profile</h3>
  <div class="row g-3">
    <div class="col-md">
      <DebouncedInput
          id="userTitle"
          placeholder="Title"
          bind:value={user.Title}
          on:change={handleTitleChange}
          persisting={persisting} />
    </div>
  </div>
  <h3 class="card-title mt-4">Public profile</h3>
  <p class="card-subtitle">Making your profile public means that anyone on the everyone will be able to find
    you.</p>
  <div>
    <label class="form-check form-switch form-switch-lg">
      <input class="form-check-input" type="checkbox">
      <span class="form-check-label form-check-label-on">You're currently visible</span>
      <span class="form-check-label form-check-label-off">You're
        currently invisible</span>
    </label>
  </div>
</div>
{/if}