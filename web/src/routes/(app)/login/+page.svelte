<script>
	import { apiEndpoint, providers } from '$lib/stores/configStore';
	let endpoint = $apiEndpoint;
</script>

<div class="page page-center">
	<div class="container container-normal py-4">
		<div class="row align-items-center g-4">
			<div class="col-lg">
				<div class="container-tight">
					<div class="text-center mb-4">
						<a href="." class="navbar-brand navbar-brand-autodark"
							><img src="/img/logo.png" style="height:36px" alt="" /></a
						>
					</div>
					<div class="card card-md">
						<div class="card-body">
							<div class="text-secondary small lh-base">
								Empower your security team with precision: <strong class="text-teal"
									>Pentest Report Information Security Management</strong
								> - your comprehensive, team-centric solution for efficient vulnerability identification,
								reporting, and follow-up. Ensuring maximum security and stability in your cybersecurity
								organization.
							</div>
							<div class="hr-text">login</div>
							{#each $providers as provider}
								<div class="row mb-3">
									<a
										class:btn-primary={provider.type == 'azure'}
										class:btn-warning={provider.type == 'gitlab'}
										class:btn-generic={provider.type != 'azure' && provider.type != 'gitlab'}
										href="{endpoint}/api/login?provider={provider.type}"
										type="submit"
										class="col btn w-100"
									>
										{#if provider.type == 'azure'}
											<svg
												xmlns="http://www.w3.org/2000/svg"
												class="icon icon-tabler icon-tabler-brand-windows"
												width="24"
												height="24"
												viewBox="0 0 24 24"
												stroke-width="1"
												stroke="currentColor"
												fill="none"
												stroke-linecap="round"
												stroke-linejoin="round"
												><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path
													d="M17.8 20l-12 -1.5c-1 -.1 -1.8 -.9 -1.8 -1.9v-9.2c0 -1 .8 -1.8 1.8 -1.9l12 -1.5c1.2 -.1 2.2 .8 2.2 1.9v12.1c0 1.2 -1.1 2.1 -2.2 1.9z"
												/><path d="M12 5l0 14" /><path d="M4 12l16 0" /></svg
											>
										{:else if provider.type == 'gitlab'}
											<svg
												xmlns="http://www.w3.org/2000/svg"
												class="icon icon-tabler icon-tabler-brand-gitlab"
												width="24"
												height="24"
												viewBox="0 0 24 24"
												stroke-width="1"
												stroke="currentColor"
												fill="none"
												stroke-linecap="round"
												stroke-linejoin="round"
												><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path
													d="M21 14l-9 7l-9 -7l3 -11l3 7h6l3 -7z"
												/></svg
											>
										{:else}
											<svg
												xmlns="http://www.w3.org/2000/svg"
												class="icon icon-tabler icon-tabler-apps"
												width="24"
												height="24"
												viewBox="0 0 24 24"
												stroke-width="1"
												stroke="currentColor"
												fill="none"
												stroke-linecap="round"
												stroke-linejoin="round"
												><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path
													d="M4 4m0 1a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v4a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z"
												/><path
													d="M4 14m0 1a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v4a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z"
												/><path
													d="M14 4m0 1a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v4a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z"
												/><path
													d="M14 14m0 1a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v4a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z"
												/></svg
											>
										{/if}
										Sign in with {provider.name}
									</a>
								</div>
							{/each}
						</div>
					</div>
					<div class="text-center text-secondary mt-3 hover">
						Made with <span class="pulse">❤️</span> from us at
						<a href="nhn.no">Norsk helsenett SF</a>
					</div>
				</div>
			</div>
			<div class="col-lg d-none d-lg-block">
				<img src="/img/login_landing_page.png" style="height:450px" class="d-block mx-auto" alt="" />
			</div>
		</div>
	</div>
</div>

<style>
	.btn-generic {
		background-color: rgb(72, 192, 127);
		color: white;
	}

	.btn-generic:hover {
		background-color: rgb(62, 172, 107);
	}

	.hover:hover .pulse {
		animation: pulseAnimation 1s infinite;
		display: inline-block;
	}

	@keyframes pulseAnimation {
		0% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.3);
		}
		100% {
			transform: scale(1);
		}
	}
</style>
