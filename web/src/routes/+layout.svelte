<script>
  import '../app.css'
	import { onMount } from 'svelte';
  import { initializeApiEndpoint, isLoading, isAuthenticated } from '$lib/stores/configStore';
	import '@tabler/core/dist/css/tabler.min.css';
	import { theme } from '$lib/stores/themeStore';
	import { page } from '$app/stores';
  import { derived } from 'svelte/store';
  import { accessLevels } from '$lib/userStore';
	import User from '$lib/components/User.svelte';
	import Loader from '$lib/components/Loader.svelte';
	import NotificationDropdown from '$lib/components/Notifications/NotificationDropdown.svelte';
  import { goto } from '$app/navigation';
  import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { Toaster } from 'svelte-sonner';

  if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/service-worker.js')
    .then(registration => {
      console.log('Service Worker registered with scope:', registration.scope);
    })
    .catch(error => {
      console.log('Service Worker registration failed:', error);
    });
}

  function navigate(url) {
    return (event) => {
      event.preventDefault(); // Prevent the default anchor navigation
      goto(url);              // Use goto for navigation
    };
  }

	function toggleTheme() {
		$theme = $theme === 'light' ? 'dark' : 'light';
	}

	const isLoginPage = derived(page, $page => $page.url.pathname === '/login');
	const isAuthPage = derived(page, $page => $page.url.pathname === '/auth');

	let isInitialized = false;

  onMount(async () => {
    await initializeApiEndpoint();
  });

	$: isInitialized = !$isLoading;
</script>
{#if isInitialized}
{#if $isLoginPage || $isAuthPage}
  <slot/>
{:else if $isAuthenticated}
<Toaster position="top-center" visibleToasts={9} richColors/>

<div class="page">
	<header class="navbar navbar-expand-sm navbar-light navbar-overlap d-print-none">
		<div class="container-xl">
			<h1 class="navbar-brand navbar-brand-autodark d-none-navbar-horizontal pe-0 pe-md-3">
				<a href="/">
					<img src="/favicon.png" width="110" height="32" alt="PRISM" class="navbar-brand-image" />
				</a>
			</h1>

			<div class="navbar-nav flex-row order-md-last">
        <div class="d-none d-md-flex cursor-pointer">
        <a on:click|preventDefault={toggleTheme} class="nav-link px-0">
						<!-- Download SVG icon from http://tabler-icons.io/i/moon -->
						{#if $theme === 'dark'}
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
								><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
									d="M12 3c.132 0 .263 0 .393 0a7.5 7.5 0 0 0 7.92 12.446a9 9 0 1 1 -8.313 -12.454z"
								></path></svg
							>
						{:else}
							<!-- Download SVG icon from http://tabler-icons.io/i/sun -->
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
								><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
									d="M12 12m-4 0a4 4 0 1 0 8 0a4 4 0 1 0 -8 0"
								></path><path
									d="M3 12h1m8 -9v1m8 8h1m-9 8v1m-6.4 -15.4l.7 .7m12.1 -.7l-.7 .7m0 11.4l.7 .7m-12.1 -.7l-.7 .7"
								></path></svg
							>
						{/if}
					</a>
					<div class="nav-item dropdown d-none d-md-flex me-3">
						<NotificationDropdown />
						<div class="dropdown-menu dropdown-menu-arrow dropdown-menu-end dropdown-menu-card">
							<div class="card">
								<div class="card-header">
									<h3 class="card-title">Last updates</h3>
								</div>
								<div class="list-group list-group-flush list-group-hoverable">
									<div class="list-group-item">
										<div class="row align-items-center">
											<div class="col-auto">
												<span class="status-dot status-dot-animated bg-red d-block"></span>
											</div>
											<div class="col text-truncate">
												<a href="#" class="text-body d-block">Example 1</a>
												<div class="d-block text-secondary text-truncate mt-n1">
													Change deprecated html tags to text decoration classes (#29604)
												</div>
											</div>
											<div class="col-auto">
												<a href="#" class="list-group-item-actions">
													<!-- Download SVG icon from http://tabler-icons.io/i/star -->
													<svg
														xmlns="http://www.w3.org/2000/svg"
														class="icon text-muted"
														width="24"
														height="24"
														viewBox="0 0 24 24"
														stroke-width="2"
														stroke="currentColor"
														fill="none"
														stroke-linecap="round"
														stroke-linejoin="round"
														><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
															d="M12 17.75l-6.172 3.245l1.179 -6.873l-5 -4.867l6.9 -1l3.086 -6.253l3.086 6.253l6.9 1l-5 4.867l1.179 6.873z"
														></path></svg
													>
												</a>
											</div>
										</div>
									</div>
									<div class="list-group-item">
										<div class="row align-items-center">
											<div class="col-auto"><span class="status-dot d-block"></span></div>
											<div class="col text-truncate">
												<a href="#" class="text-body d-block">Example 2</a>
												<div class="d-block text-secondary text-truncate mt-n1">
													justify-content:between ⇒ justify-content:space-between (#29734)
												</div>
											</div>
											<div class="col-auto">
												<a href="#" class="list-group-item-actions show">
													<!-- Download SVG icon from http://tabler-icons.io/i/star -->
													<svg
														xmlns="http://www.w3.org/2000/svg"
														class="icon text-yellow"
														width="24"
														height="24"
														viewBox="0 0 24 24"
														stroke-width="2"
														stroke="currentColor"
														fill="none"
														stroke-linecap="round"
														stroke-linejoin="round"
														><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
															d="M12 17.75l-6.172 3.245l1.179 -6.873l-5 -4.867l6.9 -1l3.086 -6.253l3.086 6.253l6.9 1l-5 4.867l1.179 6.873z"
														></path></svg
													>
												</a>
											</div>
										</div>
									</div>
									<div class="list-group-item">
										<div class="row align-items-center">
											<div class="col-auto"><span class="status-dot d-block"></span></div>
											<div class="col text-truncate">
												<a href="#" class="text-body d-block">Example 3</a>
												<div class="d-block text-secondary text-truncate mt-n1">
													Update change-version.js (#29736)
												</div>
											</div>
											<div class="col-auto">
												<a href="#" class="list-group-item-actions">
													<!-- Download SVG icon from http://tabler-icons.io/i/star -->
													<svg
														xmlns="http://www.w3.org/2000/svg"
														class="icon text-muted"
														width="24"
														height="24"
														viewBox="0 0 24 24"
														stroke-width="2"
														stroke="currentColor"
														fill="none"
														stroke-linecap="round"
														stroke-linejoin="round"
														><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
															d="M12 17.75l-6.172 3.245l1.179 -6.873l-5 -4.867l6.9 -1l3.086 -6.253l3.086 6.253l6.9 1l-5 4.867l1.179 6.873z"
														></path></svg
													>
												</a>
											</div>
										</div>
									</div>
									<div class="list-group-item">
										<div class="row align-items-center">
											<div class="col-auto">
												<span class="status-dot status-dot-animated bg-green d-block"></span>
											</div>
											<div class="col text-truncate">
												<a href="#" class="text-body d-block">Example 4</a>
												<div class="d-block text-secondary text-truncate mt-n1">
													Regenerate package-lock.json (#29730)
												</div>
											</div>
											<div class="col-auto">
												<a href="#" class="list-group-item-actions">
													<!-- Download SVG icon from http://tabler-icons.io/i/star -->
													<svg
														xmlns="http://www.w3.org/2000/svg"
														class="icon text-muted"
														width="24"
														height="24"
														viewBox="0 0 24 24"
														stroke-width="2"
														stroke="currentColor"
														fill="none"
														stroke-linecap="round"
														stroke-linejoin="round"
														><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
															d="M12 17.75l-6.172 3.245l1.179 -6.873l-5 -4.867l6.9 -1l3.086 -6.253l3.086 6.253l6.9 1l-5 4.867l1.179 6.873z"
														></path></svg
													>
												</a>
											</div>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
				<div class="nav-item dropdown">
					<User />
				</div>
			</div>
			<div class="collapse navbar-collapse" id="navbar-menu">
				<div
					class="d-flex flex-column flex-md-row flex-fill align-items-stretch align-items-md-center"
				>
					<ul class="navbar-nav">
						<li class="nav-item">
							<a class="nav-link" href="/" on:click={navigate("/")}>
								<span class="nav-link-icon d-md-none d-lg-inline-block"
									><!-- Download SVG icon from http://tabler-icons.io/i/home -->
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
										><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path
											d="M5 12l-2 0l9 -9l9 9l-2 0"
										></path><path d="M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-7"></path><path
											d="M9 21v-6a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v6"
										></path></svg
									>
								</span>
								<span class="nav-link-title"> Home </span>
							</a>
						</li>
						<li class="nav-item">
							<a href="/vulnerability" class="nav-link" on:click={navigate("/vulnerability")}>
								<span class="nav-link-icon d-md-none d-lg-inline-block">
									<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-flag" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M5 5a5 5 0 0 1 7 0a5 5 0 0 0 7 0v9a5 5 0 0 1 -7 0a5 5 0 0 0 -7 0v-9z" /><path d="M5 21v-7" /></svg>
								</span>
								<span class="nav-link-title"> Vulnerabilities </span>
							</a>
						</li>
						<li class="nav-item">
							<a href="/project" class="nav-link" on:click={navigate("/project")}>
								<span class="nav-link-icon d-md-none d-lg-inline-block">
									<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-briefcase" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7m0 2a2 2 0 0 1 2 -2h14a2 2 0 0 1 2 2v9a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2z" /><path d="M8 7v-2a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v2" /><path d="M12 12l0 .01" /><path d="M3 13a20 20 0 0 0 18 0" /></svg>
								</span>
								<span class="nav-link-title"> Projects </span>
							</a>
						</li>
            {#if $accessLevels['/report']}
						<li class="nav-item">
							<a href="/report" class="nav-link" on:click={navigate("/report")}>
								<span class="nav-link-icon d-md-none d-lg-inline-block">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="icon icon-tabler icon-tabler-report-analytics"
										width="24"
										height="24"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										fill="none"
										stroke-linecap="round"
										stroke-linejoin="round"
										><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path
											d="M9 5h-2a2 2 0 0 0 -2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-12a2 2 0 0 0 -2 -2h-2"
										/><path
											d="M9 3m0 2a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v0a2 2 0 0 1 -2 2h-2a2 2 0 0 1 -2 -2z"
										/><path d="M9 17v-5" /><path d="M12 17v-1" /><path d="M15 17v-3" /></svg
									>
								</span>
								<span class="nav-link-title"> Report </span>
							</a>
						</li>
            {/if}

            {#if $accessLevels['/planning']}
						<li class="nav-item">
							<a href="/planning" class="nav-link" on:click={navigate("/planning")}>
								<span class="nav-link-icon d-md-none d-lg-inline-block">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="icon icon-tabler icon-tabler-brand-trello"
										width="24"
										height="24"
										viewBox="0 0 24 24"
										stroke-width="2"
										stroke="currentColor"
										fill="none"
										stroke-linecap="round"
										stroke-linejoin="round"
										><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path
											d="M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z"
										/><path d="M7 7h3v10h-3z" /><path d="M14 7h3v6h-3z" /></svg
									>
								</span>
								<span class="nav-link-title"> Planning </span>
							</a>
						</li>
            {/if}

					</ul>
				</div>
			</div>
		</div>
	</header>

	<div class="page-wrapper">
		<div class="page-header d-print-none">
			<div class="container-xl mb-5">
				<slot />
				<!-- This is where your page content will be injected -->
			</div>
		</div>
	</div>
</div>
{/if}
{:else}
  <Loader />
{/if}
<style>
  @import url("@tabler/icons-webfont/tabler-icons.min.css");

	[data-bs-theme='dark'] .navbar-brand-autodark .navbar-brand-image {
		filter: none !important;
	}

	.alert-container {
		position: fixed;
		top: 4em;
		right: 2em;
		z-index: 1050 !important;
	}

  .alert {
    background-color: var(--tblr-active-bg);
  }
  .cursor-pointer{
    cursor: pointer;
  }
</style>
