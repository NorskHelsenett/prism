<script>
	import { goto } from '$app/navigation';
	import { Fetch } from '$lib/fetchUtil.js';
  import AvatarList from '$lib/components/calendar/Avatarlist.svelte';
	import { onMount } from 'svelte';
	import ProjectList from '$lib/components/calendar/ProjectList.svelte';

  export let data;
  export let assessment = {}

  onMount(async () => {
    assessment = await Fetch(`/api/planning/${data.id}`)
  })

</script>

<div class="row align-items-center mb-3">
  <div class="col-auto">
    <a href="#" class="btn btn-dark w-100" on:click="{() => goto(`/planning/${data.id}/view`)}">Back</a>
  </div>
  <div class="col-auto">
    <a href="#" class="btn btn-primary w-100">Save</a>
  </div>
</div>

<div>{JSON.stringify(assessment)}</div>

<div class="row">
<div class="card">
  <!-- Photo -->
  <div class="ribbon bg-red">EDIT</div>
                  <!-- <div class="img-responsive img-responsive-21x9 card-img-top" style="background-image: url(/edit-banner.webp)"></div> -->
    <div class="card-body">
      <h3 class="card-title">Edit assessment</h3>
      <div class="card-body">

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Title</label>
          <div class="col">
            <input type="text" class="form-control" aria-describedby="emailHelp" placeholder="Enter email" bind:value={assessment.title}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Work order</label>
          <div class="col">
            <input type="text" class="form-control" aria-describedby="emailHelp" placeholder="Enter email" bind:value={assessment.workorder}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Estimate</label>
          <div class="col">
            <input type="number" class="form-control" aria-describedby="emailHelp" placeholder="Enter email" bind:value={assessment.estimate}>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Date</label>
          <div class="col">
            <input type="date" class="form-control" aria-describedby="emailHelp" placeholder="Enter email" bind:value={assessment.dateFrom}>
          </div>
          <div class="col">
            <input type="date" class="form-control" aria-describedby="emailHelp" placeholder="Enter email" bind:value={assessment.dateTo}>
          </div>
        </div>

<hr />

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Responsible</label>
          <div class="col">
            <AvatarList hackers={[{"email": assessment.responsible_hacker}]} on:updateHackers="{e => assessment.responsible_hacker = e.detail[0].email}"/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Projects</label>
          <div class="col">
            <ProjectList projects={assessment.projects} on:updateProjects="{e => assessment.projects = e.detail}"/>
          </div>
        </div>

        <div class="mb-3 row">
          <label class="col-3 col-form-label required">Hackers</label>
          <div class="col">
            <AvatarList hackers={assessment.hackers} on:updateHackers="{e => assessment.hackers = e.detail}"/>
          </div>
        </div>

<hr />
			<!-- <fieldset class="col-lg-12 form-fieldset"> -->
                <div class="mb-3 row">
          <div class="col">
            			<textarea
                    class="form-control"
                    rows=10
                    placeholder="Notes..."
                    bind:value={assessment.note} />
          </div>
        </div>
      <!-- </fieldset> -->

      </div>
    </div>
  </div>
</div>