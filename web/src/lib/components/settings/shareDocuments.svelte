<script>
  import { preventDefault } from 'svelte/legacy';

  import { Fetch } from "$lib/fetchUtil";
  import { onMount } from 'svelte';
	import { toast } from "svelte-sonner";
	import DeleteModal from "../DeleteModal.svelte";
	import { accessLevels } from "$lib/userStore";

  let loading = $state(true)
  let showDeleteModal = $state(false)
  let deleteDialogButton ="Remove public link"
  const deleteDialogText = "Are you sure that you would like to remove this public link?"
  let deletePublicLinkID = -1

  async function deleteUserPrompted() {
    try {
      const result = await Fetch(`/api/vulnerability/share/${deletePublicLinkID}`, {
        method: "DELETE"
      });
      if(!result){
        toast.error("Unable to delete public link")
      }else {
        await fetchEvents(); // Re-fetch events after retrying
        toast.success("Public link is deleted")
      }
    } catch (error) {
      toast.error("Unable to delete public link")
      console.error("Error retrying event:", error);
    }
    deletePublicLinkID = -1
  }

/** @type {import('$lib/models/shareDialog').ShareInput[]}*/
  let publiclinks = $state([]);

  async function fetchEvents() {
    try {
      publiclinks = await Fetch("/api/vulnerability/share/all");
    } catch (error) {
      console.error("Error fetching events:", error);
    }
  }

  async function deleteEvent(id) {
    deletePublicLinkID = id
    showDeleteModal = true
  }

  onMount(async() => {
    await fetchEvents(); // Initial fetch when component mounts
    loading = false
  });

  function formatDate(dateString) {
    const options = { year: 'numeric', month: '2-digit', day: '2-digit', hour12: false };
    return new Date(dateString).toLocaleDateString('en-US', options).replace(/\//g, '.').replace(',', '');
  }
</script>
{#if !loading}
<div class="card-body">
  <h2 class="card-title">Public links</h2>
  <div class="text-secondary">
    <p>Public links are URLs that provide access to shared resources or content, sometimes without requiring authentication. They are useful for quickly sharing information with a wide audience. However, it's important to manage and monitor these links to ensure they do not expose sensitive information. Always set appropriate permissions and expiration dates to mitigate unauthorized access. Regularly review and audit shared links to maintain security and compliance with organizational policies.</p>
  </div>
{#if publiclinks.length > 0}
<div class="table-responsive">
  <table class="table table-vcenter card-table table-striped">
    <thead>
      <tr>
        <th>Created</th>
        <th>Document ID</th>
        <th>Shared by</th>
        <th>Accesstype</th>
        <th>invitedEmails</th>
        <th>Share token</th>
        <th></th>
      </tr>
    </thead>
    <tbody>

      {#each publiclinks as link}
      <tr>
        <td>{formatDate(link.createdAt)}</td>
        <td> <a href="/vulnerability/{link.documentId}/view">{link.documentId}</a></td>
        <td>{link.sharedByEmail}</td>
        <td class="text-azure">{link.accessType}
        </td>
        <td>
          {link.invitedEmails}
        </td>
        <td class="text-center">
          <a href="/s/{link.shareToken}" target="_blank" >{link.shareToken}</a>
        </td>
        {#if $accessLevels["/vulnerability"]?.write}
        <td class="text-center">
          <a class="btn btn-sm btn-pill text-red" onclick={preventDefault(deleteEvent(link.documentId))}>delete</a>
        </td>
        {/if}
      </tr>
      {/each}

    </tbody>
  </table>
</div>
{:else}
<div class="card-body text-center py-4 p-sm-5">
  <img src="/img/public-link.png" class="img mb-n2" alt="apikey factory pipeline"/>
  <h1 class="mt-5 text-azure">Public links</h1>
  <div class="row">
<div class="col-2"></div>
<div class="col-8">
  <p class="text-secondary m-0">Wow, look at that! No one has bothered to create a public link. </p>
  <p class="text-secondary m-0">Guess what? You finally have a chance to be first at something!</p>
  <p class="text-secondary m-0">Don't blow it.</p>
</div>
  </div>
</div>
{/if}
</div>
{/if}

<DeleteModal bind:showDeleteModal onDelete={deleteUserPrompted} deleteButtonText={deleteDialogButton} text={deleteDialogText}/>

<style>
  .hover-alert {
    transform: translateX(-10%);
    position: absolute;
  }
  .img{
    height: 256px;
    width: 256px;
  }
</style>