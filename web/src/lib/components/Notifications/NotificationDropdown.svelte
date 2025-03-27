<script>
  import { onDestroy, onMount } from 'svelte';
  import { clickOutside } from '../clickOutside.js';
  import { userStore } from '$lib/userStore.js';
  import { goto } from '$app/navigation';
  import { slide } from 'svelte/transition'
  import NotificationButton from "$lib/components/Notifications/NotificationButton.svelte";
	import NotificationsListItem from './NotificationsListItem.svelte';
	import { Fetch } from '$lib/fetchUtil.js';
	import { toast } from 'svelte-sonner';

/**
 * Notification structure as received from server.
 * @typedef {Object} Notification
 * @property {string} who - The identifier of the user who triggered the notification.
 * @property {string} what - A description of the notification event.
 * @property {boolean} read - Whether the notification has been read.
 * @property {string} where - The location associated with the notification event.
 * @property {string} when - The timestamp of when the notification was created.
 */

/**
 * Global array to store notifications.
 * @type {Notification[]}
 */
  export let notifications = [];
  let notificationPermission = "default";

  $: if (notifications) {
    sortNotifications()
  }

  // Function to sort notifications by 'when' in descending order
  function sortNotifications() {
    /** @type {Notification[]} */
    const newNotifications = notifications.filter(notification =>
      !notifications.some(existing => existing.when === notification.when)
    );

    // Display toasts for new notifications
    newNotifications.forEach(async newNotification => {
      if(newNotification.read == false){
        const userData = await Fetch(`/api/profile/${newNotification.who}`);
        toast.info(`${userData.name}- ${newNotification.what}`, {
                  action: {
                    label: 'Show',
                    onClick: () => openAction(newNotification)
                  }
                });
      }
    });
    notifications = notifications.sort((a, b) => new Date(b.when) - new Date(a.when));
  }

  onMount(async () => {
    await sortNotifications()
    if ('Notification' in window) {
      notificationPermission = Notification.permission;
    }
  });

  let isHidden = true;

  function toggleHidden() {
      isHidden = !isHidden;
  }

  function closeDropdown() {
      isHidden = true;
  }

  let user = {
      image: "",
      role: "visitor",
      name: ""
  }

  // Subscribe to the user store
  const unsubscribe = userStore.subscribe(storeUser => {
      if (!storeUser.loading) {
          user.image = storeUser.picture;
          user.role = storeUser.role;
          user.name = storeUser.name;
      }
  });

  function handlePermissionChange(event) {
    notificationPermission = event.detail.notificationPermission;
  }

  async function openAction(notification){
    await Fetch("/api/notification/" + notification.when + "/read", {
      method:"PUT"
    });
    notification.read = true;
    window.location.href = notification.where; //@todo force redirect, fix
    await goto(notification.where)
    closeDropdown();
  }

  async function markAllRead() {
    await Fetch("/api/notification", {method: "DELETE"});
    notifications = []
    setTimeout(() => {
      closeDropdown();
    }, 1000);
  }
</script>

  <a
  href="#"
  on:click|preventDefault={toggleHidden}
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
  {#if notifications?.some(f => f.read === false) || notificationPermission == 'default'}
    <span class="badge bg-red"></span>
  {/if}
</a>
{#if !isHidden}
<div
    use:clickOutside on:outsideClick={closeDropdown}
    class="dropdown-menu dropdown-menu-end dropdown-menu-arrow show mw-400 z-1"
    data-bs-popper="static"
    transition:slide
>

  <!-- <div class="dropdown-divider"></div> -->
  {#if notificationPermission == 'granted'}
  {#if notifications?.length > 0}
    <div class=" pl-3 pr-3 pt-1">
      <div class="row">
        <div class="col">
        <label class="form-label text-azure">Notifications</label>
      </div>
      <div class="col-auto">
        <a href="#" on:click={markAllRead}>Clear all</a>
      </div>
      </div>
    </div>
  {/if}
  <div class="d-flex justify-content-center flex-wrap">
    {#if notifications?.length > 0}
        {#each notifications as notification}
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <div class="dropdown-item w-100 cursor-pointer" class:bg-secondary-lt={notification.read == false} on:click={() => openAction(notification)} >
          <div class="d-flex align-items-center w-100">
            <NotificationsListItem notification={notification}/>
          </div>
        </div>
        {/each}
    {:else}
    <svg height="300px" viewBox="0 0 100 100" id="background">
      <path fill="CornerFlowerBlue" d="M0,0, 100,0, C50,20 50,80 0,100 Z"/>
    </svg>
    <img src="/inbox-zero.png" alt="..." class="opacity-3" />
    <div class="pb-2">
      <h3 class="text-info text-small text-center m-0">No unread messages</h3>
      <p class="text-secondary text-small text-center">Enjoy yourself a cup of coffee</p>
    </div>
    {/if}
  </div>
  {:else}
  <div class="container">
    <div class="row justify-content-center">
      <div class="col">
        <h4 class="text-secondary p-4 pb-0">Enable browser notifications to get the latest news from <strong>PRISM</strong></h4>
        <div class="text-center">
          <img src="/notification.png" class="w-70" alt="Notification Icon">
        </div>
        <div class="text-center">
          <NotificationButton {notificationPermission} on:permissionChange={handlePermissionChange}/>
        </div>
      </div>
    </div>
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
  .pb-0{
    padding-bottom: 0;
  }
  .w-70{
    width: 70%;
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