<script>
	import Export from "$lib/components/settings/export.svelte";
  import Profile from "$lib/components/settings/profile.svelte";
  import APIKeys from "$lib/components/settings/apikeys.svelte";
  import About from "$lib/components/settings/about.svelte";
  import Settings from "$lib/components/settings/settings.svelte";
  import Events from "$lib/components/settings/eventQueue.svelte";
  import ShareDocs from "$lib/components/settings/shareDocuments.svelte";
	import Auditlog from "$lib/components/settings/auditlog.svelte";
	import { accessLevels } from "$lib/userStore";
	import UserTable from "$lib/components/settings/UserTable.svelte";

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
            <a href="#" class:active="{activeComponent === APIKeys}" class="list-group-item list-group-item-action d-flex align-items-center" on:click|preventDefault={() => show(APIKeys)}>API keys</a>
            {#if $accessLevels["/vulnerability"].write}
            <a href="#" class:active="{activeComponent === ShareDocs}" class="list-group-item list-group-item-action" on:click|preventDefault={() => show(ShareDocs)}>Public Links</a>
            {/if}
            <a href="#" class:active="{activeComponent === About}" class="list-group-item list-group-item-action d-flex align-items-center" on:click|preventDefault={() => show(About)}>About</a>
          </div>
          {#if $accessLevels["/settings"]}
          <h4 class="subheader mt-4">Experience</h4>
          <div class="list-group list-group-transparent">
            <a href="#" class:active="{activeComponent === Settings}" class="list-group-item list-group-item-action" on:click|preventDefault={() => show(Settings)}>Settings</a>
            <a href="#" class:active="{activeComponent === UserTable}" class="list-group-item list-group-item-action" on:click|preventDefault={() => show(UserTable)}>Users</a>
            <a href="#" class:active="{activeComponent === Events}" class="list-group-item list-group-item-action" on:click|preventDefault={() => show(Events)}>Events</a>
            <a href="#" class:active="{activeComponent === Auditlog}" class="list-group-item list-group-item-action" on:click|preventDefault={() => show(Auditlog)}>Audit log</a>
          </div>
          {/if}
        </div>
      </div>
      <div class="col-12 col-md-9 d-flex flex-column">
        <svelte:component this={activeComponent} />
      </div>
    </div>
  </div>
</div>