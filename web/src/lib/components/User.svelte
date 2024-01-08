<script>
    import { onMount } from 'svelte';
    import { clickOutside } from './clickOutside.js';
    import { Fetch } from '$lib/fetchUtil.js';

    let isHidden = true;
    let name = '';
    let image = '';

    function toggleHidden() {
        isHidden = !isHidden;
    }

    function closeDropdown() {
        isHidden = true;
    }

    onMount(async () => {
        const data = await Fetch(`/api/user`);
        name = data.name;
        image = data.picture;
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
    use:clickOutside on:outsideClick={closeDropdown}
    hidden={isHidden}
    class="dropdown-menu dropdown-menu-end dropdown-menu-arrow show"
    data-bs-popper="static"
>
    <a href="/" class="dropdown-item" on:click={closeDropdown}>Status</a>
    <a href="/settings" class="dropdown-item" on:click={closeDropdown}>Settings</a>
    <div class="dropdown-divider"></div>
    <a href="/api/logout" class="dropdown-item">Logout</a>
</div>
