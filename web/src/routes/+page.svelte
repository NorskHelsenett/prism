<script>
	import BugBounty from '$lib/components/dashboard/BugBounty.svelte';
	import Criticality from '$lib/components/dashboard/Criticality.svelte';
	import CriticalityPie from '$lib/components/charts/CriticalityPie.svelte';
	import Projects from '$lib/components/dashboard/Projects.svelte';
	import Tasks from '$lib/components/dashboard/Tasks.svelte';
	import CountVulnerabilities from '$lib/components/dashboard/countVulnerabilities.svelte';
	import VulnerabilityReportForm from '$lib/components/vulnerabilityFinding.svelte';
	import { dashboardStore } from '$lib/stores/dashboardStore';

	function handleRefresh() {
			dashboardStore.refreshData(); // Force refresh data when button clicked
	}

    let showModal = false;

  $: if (!showModal) {
    handleRefresh()
  }

    function openVulnerabilityReportForm() {
        showModal = true;
    }

    function closeModal() {
        showModal = false;
    }

	import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';
	import Categories from '$lib/components/dashboard/Categories.svelte';
	import OwaspDonut from '$lib/components/charts/OwaspDonut.svelte';
	import OwaspTable from '$lib/components/dashboard/OwaspTable.svelte';
	let severityData;
	onMount(() => {
			pageMeta.set({ pretitle: 'Overview',title: 'Pentest Report Information Security Management' });

	});

  $: if ($dashboardStore && $dashboardStore.criticalities) {
    severityData = {
      critical: $dashboardStore.criticalities.critical || 0,
      high: $dashboardStore.criticalities.high || 0,
      medium: $dashboardStore.criticalities.medium || 0,
      low: $dashboardStore.criticalities.low || 0,
      information: $dashboardStore.criticalities.information || 0
    };
  }

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
	<Criticality {severityData}/>
</div>

		<div class="row row-deck row-cards">
			<div class="col12">
				<div class="row row-cards">
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<CountVulnerabilities data={$dashboardStore.total}/>
						</div>
					</div>
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<Projects />
						</div>
					</div>
					<div class="col-sm-6 col-lg-3">
						<div class="card card-sm">
							<BugBounty data={$dashboardStore.bugBounties}/>
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
						<CriticalityPie {severityData}/>
					</div>
				</div>
			</div>
			<div class="col-8">
				<div class="card">
					<div class="card-body">
						<OwaspDonut severityData={$dashboardStore.owasp}/>
					</div>
				</div>
			</div>
			<div class="col-4">
				<div class="card">
						<div class="card-body">
                    <h3 class="card-title">Top Pages</h3>
                    <table class="table table-sm table-borderless">
                      <thead>
                        <tr>
                          <th>Page</th>
                          <th class="text-end">Visitors</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 82.54%" role="progressbar" aria-valuenow="82.54" aria-valuemin="0" aria-valuemax="100" aria-label="82.54% Complete">
                                  <span class="visually-hidden">82.54% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">4896</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 76.29%" role="progressbar" aria-valuenow="76.29" aria-valuemin="0" aria-valuemax="100" aria-label="76.29% Complete">
                                  <span class="visually-hidden">76.29% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/form-elements.html</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">3652</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 72.65%" role="progressbar" aria-valuenow="72.65" aria-valuemin="0" aria-valuemax="100" aria-label="72.65% Complete">
                                  <span class="visually-hidden">72.65% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/index.html</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">3256</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 44.89%" role="progressbar" aria-valuenow="44.89" aria-valuemin="0" aria-valuemax="100" aria-label="44.89% Complete">
                                  <span class="visually-hidden">44.89% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/icons.html</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">986</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 41.12%" role="progressbar" aria-valuenow="41.12" aria-valuemin="0" aria-valuemax="100" aria-label="41.12% Complete">
                                  <span class="visually-hidden">41.12% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/docs/</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">912</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 32.65%" role="progressbar" aria-valuenow="32.65" aria-valuemin="0" aria-valuemax="100" aria-label="32.65% Complete">
                                  <span class="visually-hidden">32.65% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/accordion.html</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">855</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 16.22%" role="progressbar" aria-valuenow="16.22" aria-valuemin="0" aria-valuemax="100" aria-label="16.22% Complete">
                                  <span class="visually-hidden">16.22% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/datagrid.html</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">764</td>
                        </tr>
                        <tr>
                          <td>
                            <div class="progressbg">
                              <div class="progress progressbg-progress">
                                <div class="progress-bar bg-primary-lt" style="width: 8.69%" role="progressbar" aria-valuenow="8.69" aria-valuemin="0" aria-valuemax="100" aria-label="8.69% Complete">
                                  <span class="visually-hidden">8.69% Complete</span>
                                </div>
                              </div>
                              <div class="progressbg-text">/datatables.html</div>
                            </div>
                          </td>
                          <td class="w-1 fw-bold text-end">686</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
					</div>
			</div>
			<!-- <div class="col-8">
				<div class="card">
					<div class="card-body" style="height: 10rem"></div>
				</div>
			</div> -->
			<div class="col-8">
				<div class="card">
					<div class="card-body">
						<div class="table-responsive w-100">
							<OwaspTable owaspData={$dashboardStore.owaspCriticalities}/>
						</div>
					</div>
				</div>
			</div>
		</div>

{#if showModal}
	<VulnerabilityReportForm bind:showModal on:close={closeModal} />
{/if}
