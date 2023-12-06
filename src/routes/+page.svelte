<script>
    import VulnerabilityReportForm from '$lib/components/vulnerabilityFinding.svelte';

    let showModal = false;

    function openVulnerabilityReportForm() {
        showModal = true;
    }

    function closeModal() {
        showModal = false;
    }

    import { pageMeta } from '$lib/stores/pageMeta';
  import { onMount } from 'svelte';

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
								<a href="#" class="btn">refresh</a>
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

		<div class="row row-deck row-cards">
			<div class="col-4">
				<div class="card">
					<div class="card-body" style="height: 10rem"></div>
				</div>
			</div>
			<div class="col-4">
				<div class="card">
					<div class="card-body" style="height: 10rem"></div>
				</div>
			</div>
			<div class="col-4">
				<div class="card">
					<div class="card-body" style="height: 10rem"></div>
				</div>
			</div>
			<div class="col-12">
				<div class="card">
					<div class="card-body" style="height: 10rem"></div>
				</div>
			</div>
		</div>


<VulnerabilityReportForm bind:showModal on:close={closeModal} />
