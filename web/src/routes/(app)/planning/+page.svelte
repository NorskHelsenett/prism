<script>
	import Modal from '$lib/components/Modal.svelte';
  import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';
	import NewAssessment from '$lib/components/calendar/newAssessment.svelte';
	import List from '$lib/components/calendar/views/List.svelte';
	import Calendar from '$lib/components/calendar/views/Calendar.svelte';
	import Swimlane from '$lib/components/calendar/views/Swimlane.svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  
  let activeComponent = List;

  function show(component) {
    activeComponent = component;
    // Update URL based on component
    let path = '/planning';
    if (component === Calendar) path = '/planning/calendar';
    if (component === Swimlane) path = '/planning/board';
    if (component === List) path = '/planning';
    goto(path, { replaceState: true });
  }

  let showModal = false

  onMount(async () => {
      pageMeta.set({ pretitle: 'Planning', title: 'Plan future world domination' });
  });

  // Reactive statement to update active component based on URL
  $: {
    const path = $page.url.pathname;
    console.log("Current path:", path); // Debug logging
    if (path.includes('/planning/calendar')) {
      activeComponent = Calendar;
    } else if (path.includes('/planning/board')) {
      activeComponent = Swimlane;
    } else if (path === '/planning') {
      activeComponent = List;
    }
  }

$: componentsToShow = [{ id: 1, component: activeComponent }];
</script>

<!-- svelte-ignore missing-declaration -->
<Modal bind:showModal on:close={() => showModal = false} large={false}>
    <div class="card-header" slot="title">
      <div class="card-title">New Assessment
      </div>
    </div>
  <NewAssessment bind:showModal on:close={() => showModal = false}/>
</Modal>

<!-- Page header - Static part, not animated -->
<div class="d-print-none">
    <div class="row g-2 align-items-center">
      <div class="col">
        <div class="page-pretitle">
          Overview
        </div>
        <h2 class="page-title">
          Planning
        </h2>
      </div>
      <!-- Page title actions -->
      <div class="col-auto ms-auto d-print-none">
        <a href="#" class="btn btn-primary" on:click={() => showModal = !showModal} >
          <!-- Download SVG icon from http://tabler-icons.io/i/plus -->
          <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"></path><path d="M5 12l14 0"></path></svg>
          Add
        </a>
      </div>
    </div>
</div>

<!-- Page body - Header tabs are static, content is animated -->
<div class="page-body" style="margin-top: 37px;">
    <ul class="nav nav-bordered mb-4">
      <li class="nav-item">
        <a class="nav-link" class:active="{activeComponent === Calendar}" on:click|preventDefault={() => show(Calendar)} aria-current="page" href="#">Calendar</a>
      </li>
      <li class="nav-item">
        <a class="nav-link" class:active="{activeComponent === List}" on:click|preventDefault={() => show(List)} href="#">List</a>
      </li>
      <li class="nav-item">
        <a class="nav-link" class:active="{activeComponent === Swimlane}" on:click|preventDefault={() => show(Swimlane)} href="#">Board</a>
      </li>
    </ul>

    <!-- Only the component content is animated now -->
    <div class="component-container">
      {#each componentsToShow as { component } (component)}
          <div>
            <svelte:component this={component} reload={showModal} />
          </div>
      {/each}
    </div>
</div>

<style>
  .component-container {
    position: relative;
    min-height: 300px; /* Adjust this based on the minimum height of your components */
  }
</style>