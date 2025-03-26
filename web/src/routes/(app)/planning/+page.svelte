<script>
	import Modal from '$lib/components/Modal.svelte';
  import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';
	import NewAssessment from '$lib/components/calendar/newAssessment.svelte';
	import List from '$lib/components/calendar/views/List.svelte';
	import Calendar from '$lib/components/calendar/views/Calendar.svelte';
	import Swimlane from '$lib/components/calendar/views/Swimlane.svelte';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
  let activeComponent = List;

  function show(component) {
    activeComponent = component;
  }

  let showModal = false

  onMount(async () => {
      pageMeta.set({ pretitle: 'Planning',title: 'Plan future world domination' });
  });

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

        <!-- Page header -->
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
        <!-- Page body -->
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

        {#each componentsToShow as { component } (component)}
            <div transition:fly={{ delay: 0, duration: 90, x: 3000, y: 0, opacity: 0.5, easing: quintOut }}>
              <svelte:component this={component} reload={showModal} />
            </div>
        {/each}
        </div>

<style>

</style>