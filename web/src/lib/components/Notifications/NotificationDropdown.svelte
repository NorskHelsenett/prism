<script>
  import { preventDefault } from 'svelte/legacy';

  import { onMount } from 'svelte';
  import { clickOutside } from '../clickOutside.js';
  import { slide } from 'svelte/transition'
  import NotificationButton from "$lib/components/Notifications/NotificationButton.svelte";
  import NotificationsListItem from './NotificationsListItem.svelte';
  import { notifications, unreadCount, markAllRead, openNotification } from '$lib/stores/notificationStore.js';

  // The list renders straight from the store — delivery, toasts and cross-tab
  // sync live in notificationStore.js (fed by the per-user SSE stream). The
  // in-app list is NOT gated on browser push permission: push is an optional
  // extra channel, offered below the list while the user hasn't decided yet.
  let notificationPermission = $state("default");

  onMount(() => {
    if ('Notification' in window) {
      notificationPermission = Notification.permission;
    }
  });

  let isHidden = $state(true);

  function toggleHidden() {
      isHidden = !isHidden;
  }

  function closeDropdown() {
      isHidden = true;
  }

  function handlePermissionChange(event) {
    notificationPermission = event.detail.notificationPermission;
  }

  async function openAction(notification){
    await openNotification(notification);
    closeDropdown();
  }

  async function markAllReadAction() {
    await markAllRead();
    setTimeout(closeDropdown, 1000);
  }
</script>

  <a
  href="#"
  onclick={preventDefault(toggleHidden)}
  class="nav-link px-0  d-flex lh-1 text-reset"
  data-bs-toggle="dropdown"
  tabindex="-1"
  aria-label="Show notifications"
>
  <!-- Download SVG icon from http://tabler-icons.io/i/bell -->
  <svg
    xmlns="http://www.w3.org/2000/svg"
    class="icon"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    stroke-width="2"
    stroke="currentColor"
    fill="none"
    stroke-linecap="round"
    stroke-linejoin="round"
    ><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
      d="M10 5a2 2 0 1 1 4 0a7 7 0 0 1 4 6v3a4 4 0 0 0 2 3h-16a4 4 0 0 0 2 -3v-3a7 7 0 0 1 4 -6"
    ></path><path d="M9 17v1a3 3 0 0 0 6 0v-1"></path></svg
  >
  {#if $unreadCount > 0}
    <span class="badge bg-red"></span>
  {/if}
</a>
{#if !isHidden}
<div
    use:clickOutside onoutsideClick={closeDropdown}
    class="dropdown-menu dropdown-menu-end dropdown-menu-arrow show mw-400 z-1"
    data-bs-popper="static"
    transition:slide
>

  {#if $notifications.length > 0}
    <div class=" pl-3 pr-3 pt-1">
      <div class="row">
        <div class="col">
        <label class="form-label text-azure">Notifications</label>
      </div>
      <div class="col-auto">
        <a href="#" onclick={markAllReadAction}>Mark all read</a>
      </div>
      </div>
    </div>
  {/if}
  <div class="d-flex justify-content-center flex-wrap">
    {#if $notifications.length > 0}
        {#each $notifications as notification (notification.id)}
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <div class="dropdown-item w-100 cursor-pointer" class:bg-secondary-lt={notification.read == false} onclick={() => openAction(notification)} >
          <div class="d-flex align-items-center w-100">
            <NotificationsListItem notification={notification}/>
          </div>
        </div>
        {/each}
    {:else}
    <svg height="300px" viewBox="0 0 100 100" id="background">
      <path fill="CornerFlowerBlue" d="M0,0, 100,0, C50,20 50,80 0,100 Z"/>
    </svg>
    <img src="/img/inbox-zero.png" alt="..." class="opacity-3" />
    <div class="pb-2">
      <h3 class="text-info text-small text-center m-0">No unread messages</h3>
      <p class="text-secondary text-small text-center">Enjoy yourself a cup of coffee</p>
    </div>
    {/if}
  </div>
  {#if notificationPermission === 'default'}
  <div class="pl-3 pr-3 pt-1 text-center border-top">
    <p class="text-secondary text-small mb-2 mt-2">Enable browser notifications to get the latest news from <strong>PRISM</strong></p>
    <NotificationButton {notificationPermission} on:permissionChange={handlePermissionChange}/>
  </div>
  {/if}
</div>
{/if}

<style>
  .dropdown-menu {
    max-height: 35em;
    overflow: auto;
    z-index: 1000000 !important;
    padding-bottom: 20px;
  }
  .mw-400{
    min-width: 400px;
  }
  .z-1{
    z-index: 1;
  }
  #background{
    position: absolute;
    z-index: -1;
    height: 100%;
    width: 100%;
    top: -9%;
    fill: color(srgb 0.2296 0.5287 0.82 / 0.12);
  }
  .opacity-3 {
      opacity:1;
  }
  .pl-3{
    padding-left: 1.5em;
  }
  .pr-3 {
    padding-right: 1.5em;
  }
  .pt-1 {
    padding-top:.75rem!important;
  }
  .pb-2{
    padding-bottom: 1.5em !important;
  }
</style>
