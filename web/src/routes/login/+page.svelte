<script>
	import {apiEndpoint, providers} from '$lib/stores/configStore';
  let endpoint = $apiEndpoint
</script>

<div class="page page-center">
	<div class="container container-normal py-4">
		<div class="row align-items-center g-4">
			<div class="col-lg">
				<div class="container-tight">
					<div class="text-center mb-4">
						<a href="." class="navbar-brand navbar-brand-autodark"
							><img src="./logo.png" height="36" alt="" /></a
						>
					</div>
					<div class="card card-md">
						<div class="card-body">
							<div class="text-secondary small lh-base">
								Empower your security team with precision: Pentest Report Information Security
								Management - your comprehensive, team-centric solution for efficient vulnerability
								identification, reporting, and follow-up. Ensuring maximum security and stability in
								your cybersecurity operations.
							</div>
							<div class="hr-text">login</div>
              {#each $providers as provider}
							<div class="row mb-3">

                <a class:btn-primary={provider.type == "azure"} class:btn-warning={provider.type == "gitlab"} href="{endpoint}/api/login?provider={provider.type}" type="submit" class="col btn w-100">
                  {#if provider.type == "azure"}
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-brand-windows" width="24" height="24" viewBox="0 0 24 24" stroke-width="1" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M17.8 20l-12 -1.5c-1 -.1 -1.8 -.9 -1.8 -1.9v-9.2c0 -1 .8 -1.8 1.8 -1.9l12 -1.5c1.2 -.1 2.2 .8 2.2 1.9v12.1c0 1.2 -1.1 2.1 -2.2 1.9z" /><path d="M12 5l0 14" /><path d="M4 12l16 0" /></svg>
                  {:else if provider.type == "gitlab"}
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-brand-gitlab" width="24" height="24" viewBox="0 0 24 24" stroke-width="1" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M21 14l-9 7l-9 -7l3 -11l3 7h6l3 -7z" /></svg>
                  {/if}
									Sign in with {provider.name}
								</a>
							</div>
              {/each}
						</div>
					</div>
					<div class="text-center text-secondary mt-3">
						Made with ❤️ from us at Norsk helsenett SF
					</div>
				</div>
			</div>
			<div class="col-lg d-none d-lg-block">
				<img src="./login_landing_page.png" height="450" class="d-block mx-auto" alt="" />
			</div>
		</div>
	</div>
</div>
