<script>
	import BugBounty from '$lib/components/dashboard/BugBounty.svelte';
	import Criticality from '$lib/components/dashboard/Criticality.svelte';
	import Criticalities from '$lib/components/dashboard/Criticalities.svelte';
	import Projects from '$lib/components/dashboard/Projects.svelte';
	import Tasks from '$lib/components/dashboard/Tasks.svelte';
	import CountVulnerabilities from '$lib/components/dashboard/countVulnerabilities.svelte';
	import VulnerabilityReportForm from '$lib/components/vulnerabilityFinding.svelte';
	import { dashboardStore } from '$lib/stores/dashboardStore';

	function handleRefresh() {
			dashboardStore.refreshData(); // Force refresh data when button clicked
	}

    let showModal = false;

    function openVulnerabilityReportForm() {
        showModal = true;
    }

    function closeModal() {
        showModal = false;
    }

	import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';
	import Categories from '$lib/components/dashboard/Categories.svelte';

	onMount(() => {
			pageMeta.set({ pretitle: 'Overview',title: 'Pentest Report Information Security Management' });
	});

</script>

				<div class="row g-2 align-items-center">
					<div class="col">
						<div class="page-pretitle">{$pageMeta.pretitle}</div>
						<h2 class="page-title">
							{$pageMeta.title}
						</h2>
					</div>
					<div class="col-auto ms-auto d-print-non">
						<div class="btn-list">
							<span class="d-none d-sm-inline">
								<a href="#" class="btn" on:click={handleRefresh}>refresh</a>
							</span>
							<a
								href="#"
								class="btn btn-primary d-none d-sm-inline-block"
								on:click={openVulnerabilityReportForm}
							>
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
									><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M12 5l0 14"
									></path><path d="M5 12l14 0"></path></svg
                  >
                  Create new vulnerability</a
                  >
                </div>
              </div>
              <div class="page-body">
						<div class="container-xl">
						</div>
					</div>
				</div>

<div class="row-deck row-cards" >
	<Criticality />
</div>

		<div class="row row-deck row-cards">
			<div class="col12">
				<div class="row row-cards">
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<CountVulnerabilities />
						</div>
					</div>
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<Projects />
						</div>
					</div>
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<BugBounty />
						</div>
					</div>
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<Tasks />
						</div>
					</div>
				</div>
			</div>
			<div class="col-4">
				<div class="card">
					<div class="card-body">
						<!-- <Categories /> -->
					</div>
				</div>
			</div>
			<div class="col-4">
				<div class="card">
					<div class="card-body" style="height: 10rem">
					</div>
				</div>
			</div>
			<div class="col-4">
				<div class="card">
					<div class="card-body"></div>
				</div>
			</div>
			<div class="col-12">
				<div class="card">
					<div class="card-body" style="height: 10rem"></div>
				</div>
			</div>
		</div>


<VulnerabilityReportForm bind:showModal on:close={closeModal} />
