<script>
	import Export from "$lib/components/settings/export.svelte";
  import Profile from "$lib/components/settings/profile.svelte";
  import About from "$lib/components/settings/about.svelte";

  let activeComponent = Profile;

  function show(component) {
    activeComponent = component;
  }
</script>
<div class="row g-2 align-items-center">
  <div class="col">
    <div class="page-pretitle">System Settings</div>
    <h2 class="page-title">
      Settings
    </h2>
  </div>
</div>

<div class="page-body">
  <div class="card">
    <div class="row g-0">
      <div class="col-12 col-md-3 border-end">
        <div class="card-body">
          <h4 class="subheader">Business settings</h4>
          <div class="list-group list-group-transparent">
            <a href="#" class:active="{activeComponent === Profile}" class="list-group-item list-group-item-action d-flex align-items-center" on:click|preventDefault={() => show(Profile)}>My Account</a>
            <a href="#" class:active="{activeComponent === About}" class="list-group-item list-group-item-action d-flex align-items-center" on:click|preventDefault={() => show(About)}>About</a>
          </div>
          <h4 class="subheader mt-4">Experience</h4>
          <div class="list-group list-group-transparent">
            <a href="#" class:active="{activeComponent === Export}" class="list-group-item list-group-item-action" on:click|preventDefault={() => show(Export)}>Export</a>
          </div>
        </div>
      </div>
      <div class="col-12 col-md-9 d-flex flex-column">
        <svelte:component this={activeComponent} />
      </div>
    </div>
  </div>
</div>