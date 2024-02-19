<script>
	import Modal from '$lib/components/Modal.svelte';

  import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';
	import { slide } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { Fetch } from '$lib/fetchUtil';
	import Avatar from '$lib/components/Avatar.svelte';
	import NewAssessment from '$lib/components/calendar/newAssessment.svelte';

  let showModal = false

  onMount(async () => {
      pageMeta.set({ pretitle: 'Planning',title: 'Plan future world domination' });

      // calendarEvents = await Fetch("/api/planning")
  });

  let calendarEvents = []

  let months =["Jan", "Feb","Mar", "Apr","May","Jun", "Aug", "Sep", "Oct", "Nov", "Dec"]

function eventIn(month, dateFrom, dateTo) {
  // Parse the month string to get the month index
  const monthIndex = months.indexOf(month);

  // Parse the dateFrom and dateTo strings to Date objects
  const fromDate = new Date(dateFrom);
  const toDate = new Date(dateTo);

  // Check if the month of dateFrom or dateTo matches the specified month
  // Note: getMonth() returns a 0-based index, so January is 0, February is 1, etc.
  return fromDate.getMonth() === monthIndex || toDate.getMonth() === monthIndex;
}

// Define an async function to fetch the data
  async function fetchCalendarEvents() {
    calendarEvents = await Fetch("/api/planning")
  }

$: if(showModal == false) {
  fetchCalendarEvents()
}

</script>

<!-- svelte-ignore missing-declaration -->
<Modal bind:showModal on:close={() => showModal = false} large={false}>
    <div class="card-header" slot="title">
      <div class="card-title">New Assassment
      </div>
    </div>
  <NewAssessment bind:showModal on:close={() => showModal = false}/>
</Modal>

        <!-- Page header -->
        <div class="page-header d-print-none">
          <div class="container-xl">
            <div class="row g-2 align-items-center">
              <div class="col">
                <h2 class="page-title">
                  Planning
                </h2>
              </div>
              <!-- Page title actions -->
              <div class="col-auto ms-auto d-print-none">
                <a href="#" class="btn btn-primary" on:click={() => showModal = !showModal} transition:slide={{ delay: 250, duration: 300, easing: quintOut, axis: 'x' }}>
                  <!-- Download SVG icon from http://tabler-icons.io/i/plus -->
                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"></path><path d="M5 12l14 0"></path></svg>
                  Add
                </a>
              </div>
            </div>
          </div>
        </div>
        <!-- Page body -->
        <div class="page-body" style="margin-top: 17px;">
          <div class="container-xl">
            <ul class="nav nav-bordered mb-4">
              <li class="nav-item">
                <a class="nav-link" aria-current="page" href="#">Calendar</a>
              </li>
              <li class="nav-item">
                <a class="nav-link active" href="#">List</a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">Board</a>
              </li>
            </ul>
            <div class="row" hidden>
              <div class="col-12 col-md-6 col-lg">
                <h2 class="mb-3">To Do</h2>
                <div class="mb-4">
                  <div class="row row-cards">
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">Enable analytics tracking</h3>
                          <div class="ratio ratio-16x9">
                            <img src="./static/projects/dashboard-1.png" class="rounded object-cover" alt="Enable analytics tracking">
                          </div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded">EP</span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/002f.jpg)"></span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/003f.jpg)"></span>
                                  <span class="avatar avatar-xs rounded">HS</span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                7
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  2</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">User should receive a daily digest email</h3>
                          <div class="text-secondary">Dedicated form for a category of users that will perform actions.</div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/000f.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-warning">
                                  <!-- Download SVG icon from http://tabler-icons.io/i/calendar -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12z"></path><path d="M16 3v4"></path><path d="M8 3v4"></path><path d="M4 11h16"></path><path d="M11 15h1"></path><path d="M12 15v3"></path></svg>
                                  10 Sep
                                </a>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                6
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-status-top bg-yellow"></div>
                        <div class="card-body">
                          <h3 class="card-title">Change license and remove references to products</h3>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale active" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                34
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  4</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-6 col-lg">
                <h2 class="mb-3">In Progress</h2>
                <div class="mb-4">
                  <div class="row row-cards">
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-status-top bg-green"></div>
                        <div class="card-body">
                          <h3 class="card-title">Write a release note for Admin v1.0</h3>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted">
                                  <!-- Download SVG icon from http://tabler-icons.io/i/activity -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M3 12h4l3 8l4 -16l3 8h4"></path></svg>
                                  1/3
                                </a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  11</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                          <div class="divide-y-2 mt-4">
                            <div>
                              <!-- Download SVG icon from http://tabler-icons.io/i/check -->
                              <svg xmlns="http://www.w3.org/2000/svg" class="icon text-muted" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M5 12l5 5l10 -10"></path></svg>
                              <span class="text-secondary text-decoration-line-through">Implement new designs</span>
                            </div>
                            <div>
                              <!-- Download SVG icon from http://tabler-icons.io/i/check -->
                              <svg xmlns="http://www.w3.org/2000/svg" class="icon text-green" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M5 12l5 5l10 -10"></path></svg>
                              Usability testing
                            </div>
                            <div>
                              <!-- Download SVG icon from http://tabler-icons.io/i/check -->
                              <svg xmlns="http://www.w3.org/2000/svg" class="icon text-green" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M5 12l5 5l10 -10"></path></svg>
                              Design navigation changes
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="ribbon ribbon-top ribbon-bookmark bg-yellow">
                          <!-- Download SVG icon from http://tabler-icons.io/i/star -->
                          <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 17.75l-6.172 3.245l1.179 -6.873l-5 -4.867l6.9 -1l3.086 -6.253l3.086 6.253l6.9 1l-5 4.867l1.179 6.873z"></path></svg>
                        </div>
                        <div class="card-body">
                          <h3 class="card-title">Product Update - Q4 2021</h3>
                          <div class="text-secondary">Dedicated form for a category of users that will perform actions.</div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/002f.jpg)"></span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/003f.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                11
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  6</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">Code HTML email template for welcome email</h3>
                          <div class="ratio ratio-16x9">
                            <img src="./static/projects/dashboard-3.png" class="rounded object-cover" alt="Code HTML email template for welcome email">
                          </div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  11</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-6 col-lg">
                <h2 class="mb-3">Test</h2>
                <div class="mb-4">
                  <div class="row row-cards">
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-status-top bg-red"></div>
                        <div class="card-body">
                          <h3 class="card-title">Design new diagrams</h3>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded">HS</span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/006m.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                6
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  9</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">Improve animation loader</h3>
                          <div class="ratio ratio-16x9">
                            <img src="./static/projects/dashboard-2.png" class="rounded object-cover" alt="Improve animation loader">
                          </div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/004f.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale active" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                5
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  6</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">iOS App home page</h3>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/002m.jpg)"></span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/003m.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-status-top bg-blue"></div>
                        <div class="card-body">
                          <h3 class="card-title">Changelog 1.7</h3>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-warning">
                                  <!-- Download SVG icon from http://tabler-icons.io/i/calendar -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12z"></path><path d="M16 3v4"></path><path d="M8 3v4"></path><path d="M4 11h16"></path><path d="M11 15h1"></path><path d="M12 15v3"></path></svg>
                                  10 Jan
                                </a>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/message -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M8 9h8"></path><path d="M8 13h6"></path><path d="M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z"></path></svg>
                                  6</a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-12 col-md-6 col-lg">
                <h2 class="mb-3">Completed</h2>
                <div class="mb-4">
                  <div class="row row-cards">
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">Enable analytics tracking</h3>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/002f.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                1
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">Coordinate with business development</h3>
                          <div class="text-secondary">This content is a little bit longer.</div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/000m.jpg)"></span>
                                  <span class="avatar avatar-xs rounded">JL</span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/002m.jpg)"></span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/003m.jpg)"></span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/000f.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale active" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                7
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted">
                                  <!-- Download SVG icon from http://tabler-icons.io/i/activity -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M3 12h4l3 8l4 -16l3 8h4"></path></svg>
                                  1/3
                                </a>
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                          <div class="divide-y-2 mt-4">
                            <div>
                              <!-- Download SVG icon from http://tabler-icons.io/i/check -->
                              <svg xmlns="http://www.w3.org/2000/svg" class="icon text-muted" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M5 12l5 5l10 -10"></path></svg>
                              <span class="text-secondary text-decoration-line-through">Find out the old contract documents</span>
                            </div>
                            <div>
                              <!-- Download SVG icon from http://tabler-icons.io/i/check -->
                              <svg xmlns="http://www.w3.org/2000/svg" class="icon text-green" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M5 12l5 5l10 -10"></path></svg>
                              Organize meeting sales associates to understand need in detail
                            </div>
                            <div>
                              <!-- Download SVG icon from http://tabler-icons.io/i/check -->
                              <svg xmlns="http://www.w3.org/2000/svg" class="icon text-green" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M5 12l5 5l10 -10"></path></svg>
                              Make sure to cover every small details
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="col-12">
                      <div class="card card-sm">
                        <div class="card-body">
                          <h3 class="card-title">Managing teams</h3>
                          <div class="text-secondary">Publishing industries for previewing layouts and visual <a href="#">#family</a> 🔥</div>
                          <div class="mt-4">
                            <div class="row">
                              <div class="col">
                                <div class="avatar-list avatar-list-stacked">
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/006m.jpg)"></span>
                                  <span class="avatar avatar-xs rounded" style="background-image: url(./static/avatars/004f.jpg)"></span>
                                </div>
                              </div>
                              <div class="col-auto text-secondary">
                                <button class="switch-icon switch-icon-scale" data-bs-toggle="switch-icon">
                                  <span class="switch-icon-a text-muted">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                  <span class="switch-icon-b text-red">
                                    <!-- Download SVG icon from http://tabler-icons.io/i/heart -->
                                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-filled" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M19.5 12.572l-7.5 7.428l-7.5 -7.428a5 5 0 1 1 7.5 -6.566a5 5 0 1 1 7.5 6.572"></path></svg>
                                  </span>
                                </button>
                                4
                              </div>
                              <div class="col-auto">
                                <a href="#" class="link-muted"><!-- Download SVG icon from http://tabler-icons.io/i/share -->
                                  <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M6 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 6m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M18 18m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path><path d="M8.7 10.7l6.6 -3.4"></path><path d="M8.7 13.3l6.6 3.4"></path></svg>
                                </a>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="card">
              <div class="table-responsive small">
                <table class="table table-vcenter card-table">
                  <thead>
                    <tr>
                      <th class="sticky-col first-col">Title</th>
                      <th>Estimate</th>
                      <th>AO</th>
                      <th>Project</th>
                      <th>Ordered by</th>
                      <th>Responsible</th>
                      <th>Hackers</th>
                      <th>Status</th>
                      <th>From</th>
                      <th>To</th>
                      <th>Note</th>
                      {#each months as month}
                        <th>{month}</th>
                      {/each}
                    </tr>
                  </thead>
                  <tbody>
                    {#if calendarEvents?.length > 0}
                    {#each calendarEvents as event}
                    <tr>
                      <td class="sticky-col first-col">{event?.description}</td>
                      <td>40 h</td>
                      <td>4096</td>
                      <td>
                        {#each event.projects as project}
                          {project.name}
                        {/each}
                      </td>
                      <td>Ordered by</td>
                      <td>
                        <Avatar email="{event.responsible_hacker}" option={{ showName: false, size: "sm", emptyFields: false, circle: true}}/>
                      </td>
                      <td>
                        <i class="ti ti-circle-check-filled text-green"></i>
                        <i class="ti ti-clock-filled text-yellow"></i>
                        <i class="ti ti-calendar-time text-orange"></i>
                      </td>
                      <td>
                        <div class="avatar-list avatar-list-stacked" style="min-width:10em">
                          {#each event.hackers as hacker}
                            <Avatar email="{hacker?.email}" option={{ showName: false, size: "sm", emptyFields: false, circle: true}}/>
                            {/each}
                        </div>
                      </td>
                      <td>{event?.dateFrom}</td>
                      <td>{event?.dateTo}</td>
                      <td>
                        {#if event?.note}
                          <i class="ti ti-notes" title="{event?.note}"></i>
                        {/if}
                      </td>
                      {#each months as month}
                        {#if eventIn(month, event?.dateFrom, event?.dateTo)}
                          <td>x</td>
                        {:else}
                          <td></td>
                        {/if}
                      {/each}
                    </tr>
                    {/each}
                    {/if}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

<style>
  .sticky-col {
    position: -webkit-sticky; /* For Safari */
    position: sticky;
    background-color: var(--tblr-body-bg); /* Background color is necessary to avoid content overlap */
    left: 0;
    z-index: 100; /* Ensure the sticky column is above other elements */
}

/* Add this if you want a border separation */
.first-col {
    border-right: solid 1px var(--tblr-body-bg); /* Bootstrap's default border color */
}
</style>