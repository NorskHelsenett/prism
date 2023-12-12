<script>
	import { apiEndpoint } from '$lib/stores/configStore';
	import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

	let isHidden = true;

	function toggleHidden() {
		isHidden = !isHidden;
	}

	let name = '';
	let image = '';

	onMount(async () => {
		try {
			const response = await fetch(`${$apiEndpoint}/api/user`,{ credentials: 'include' });
			if (response.status === 401) {
				// Redirect to your login page or handle the 401 error as needed
				goto('/login');
				return;
			}

			if (!response.ok) {
        goto('/login');
				throw new Error(`Error: ${response.status}`);
			}

			const data = await response.json();
			name = data.name;
			image = data.picture;
		} catch (error) {
			console.error('Error fetching data:', error);
			goto('/login');
		}
	});
</script>

<a href="#" on:click|preventDefault={toggleHidden} class="nav-link d-flex lh-1 text-reset p-0">
	<span class="avatar avatar-sm" style="background-image: url({image})"></span>
	<div class="d-none d-xl-block ps-2">
		<div>{name}</div>
		<div class="mt-1 small text-secondary">Security Engineer/pen-tester</div>
	</div>
</a>
<div
	hidden={isHidden}
	class="dropdown-menu dropdown-menu-end dropdown-menu-arrow show"
	data-bs-theme="dark"
	data-bs-popper="static"
>
	<a href="#" class="dropdown-item">Status</a>
	<a href="./profile.html" class="dropdown-item">Profile</a>
	<div class="dropdown-divider"></div>
	<a href="./settings.html" class="dropdown-item">Settings</a>
	<a href="/api/logout" class="dropdown-item">Logout</a>
</div>
