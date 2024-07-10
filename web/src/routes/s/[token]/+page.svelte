<script>
	import SeverityBucket from '$lib/components/severityBucket.svelte';
	import ImpactBucket from '$lib/components/ImpactBucket.svelte';
	import Markdown from '$lib/components/Markdown.svelte';
	import CriticalityDonoutSvg from '$lib/components/charts/CriticalityDonoutSvg.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import { Fetch } from '$lib/fetchUtil';
  import { onMount } from 'svelte';
	import ImageCarousel from '$lib/components/ImageCarousel.svelte';
  import { fade } from 'svelte/transition';

  /**
   * @typedef {import('$lib/models/vulnerabilityDetail').VulnerabilityData} VulnerabilityData
   */
  /** @type {VulnerabilityData} */
  let vulnerability

  /** @type {import('./$types').PageData} */
  export let data;
  let token = data.token
  let showExploitationTooltip = false

  let requirePassphrase = false
  let showExpiredMessage = false
  let showUnauthorizedMessage = false
  let showNotFoundMessage = false
  let showRateLimitMessage = false

  onMount(async() => {
    localStorage.removeItem("redirectToAfterLogin")
    const result = await Fetch(`/api/share/${token}`, {method: "POST", body: "{}"})
    console.log(result)
      if(result?.error == "Passphrase is required"){
        requirePassphrase = true
      } else if (result?.error == "expired"){
        showExpiredMessage = true
      } else if (result?.error == "unauthorized"){
        showUnauthorizedMessage = true
      } else if (result?.error == "notFound"){
        showNotFoundMessage = true
      } else if (result?.error == "Rate limit exceeded"){
        showRateLimitMessage = true
      } else {
      vulnerability = result
    }
  });

  let showModal = false;
  let currentImageIndex = 0;

  function openModal(index) {
    currentImageIndex = index;
    showModal = true;
  }

  let passphrase = ""
  async function submitPassphrase() {
    const url = `/api/share/${token}`
    const result = await Fetch(url, {method: "POST", body: `{"passphrase":"${passphrase}"}`})
    if(result?.error == "Passphrase is required"){
      requirePassphrase = true
    }else {
      vulnerability = result
      requirePassphrase = false
    }
  }
  async function handleKeydown(event) {
    if (event.key === 'Enter') {
      await submitPassphrase();
    }
  }

  function replaceUrlWithSlackPrefix(text) {
    const urlRegex = /https?:\/\/[^\s]+/g;
    return text.replace(urlRegex, match => `slack://${match}`);
  }
</script>

{#if showRateLimitMessage}
<div class="d-flex justify-content-center align-items-center vh-100" transition:fade={{ delay: 250, duration: 300 }}>
  <div class="card">
    <div class="card-status-top bg-danger"></div>
    <div class="card-body text-center pb-0" style="min-width: 30em;">
      <div class="display-2 fw-bold my-3 text-danger">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon-tabler icon-tabler-alarm-filled" width="96" height="96" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M16 6.072a8 8 0 1 1 -11.995 7.213l-.005 -.285l.005 -.285a8 8 0 0 1 11.995 -6.643zm-4 2.928a1 1 0 0 0 -1 1v3l.007 .117a1 1 0 0 0 .993 .883h2l.117 -.007a1 1 0 0 0 .883 -.993l-.007 -.117a1 1 0 0 0 -.993 -.883h-1v-2l-.007 -.117a1 1 0 0 0 -.993 -.883z" stroke-width="0" fill="currentColor" />
          <path d="M6.412 3.191a1 1 0 0 1 1.273 1.539l-.097 .08l-2.75 2a1 1 0 0 1 -1.273 -1.54l.097 -.08l2.75 -2z" stroke-width="0" fill="currentColor" />
          <path d="M16.191 3.412a1 1 0 0 1 1.291 -.288l.106 .067l2.75 2a1 1 0 0 1 -1.07 1.685l-.106 -.067l-2.75 -2a1 1 0 0 1 -.22 -1.397z" stroke-width="0" fill="currentColor" />
        </svg>
      </div>
    </div>
    <div class="empty pt-0">
      <p class="empty-title text-danger">Rate Limit Exceeded</p>
      <p class="empty-subtitle text-secondary">
        Take a deep breath, relax. And come back in 20min.
      </p>
    </div>
  </div>
</div>
{:else if showNotFoundMessage}
<div class="d-flex justify-content-center align-items-center vh-100" transition:fade={{ delay: 250, duration: 300 }}>
  <div class="card">
    <div class="card-status-top bg-danger"></div>
    <div class="card-body text-center pb-0" style="min-width: 30em;">
      <div class="display-2 fw-bold my-3 text-danger">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon-tabler icon-tabler-sign-right-filled" width="96" height="96" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M10 2a1 1 0 0 1 .993 .883l.007 .117v2h5a1 1 0 0 1 .694 .28l.087 .095l2 2.5a1 1 0 0 1 .072 1.147l-.072 .103l-2 2.5a1 1 0 0 1 -.652 .367l-.129 .008h-5v8h1a1 1 0 0 1 .117 1.993l-.117 .007h-4a1 1 0 0 1 -.117 -1.993l.117 -.007h1v-8h-3a1 1 0 0 1 -.993 -.883l-.007 -.117v-5a1 1 0 0 1 .883 -.993l.117 -.007h3v-2a1 1 0 0 1 1 -1z" stroke-width="0" fill="currentColor" />
        </svg>
      </div>
    </div>
    <div class="empty pt-0">
      <p class="empty-title text-danger">404</p>
      <p class="empty-subtitle text-secondary">
        You are lost! ... Sooo lost...
      </p>
    </div>
  </div>
</div>
{:else if showUnauthorizedMessage}
<div class="d-flex justify-content-center align-items-center vh-100" transition:fade={{ delay: 250, duration: 300 }}>
  <div class="card">
    <div class="card-status-top bg-danger"></div>
    <div class="card-body text-center pb-0" style="min-width: 30em;">
      <div class="display-2 fw-bold my-3 text-danger">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon-tabler icon-tabler-ghost-2-filled" width="96" height="96" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M12 1.999l.041 .002l.208 .003a8 8 0 0 1 7.747 7.747l.003 .248l.177 .006a3 3 0 0 1 2.819 2.819l.005 .176a3 3 0 0 1 -3 3l-.001 1.696l1.833 2.75a1 1 0 0 1 -.72 1.548l-.112 .006h-10c-3.445 .002 -6.327 -2.49 -6.901 -5.824l-.028 -.178l-.071 .001a3 3 0 0 1 -2.995 -2.824l-.005 -.175a3 3 0 0 1 3 -3l.004 -.25a8 8 0 0 1 7.996 -7.75zm0 10.001a2 2 0 0 0 -2 2a1 1 0 0 0 1 1h2a1 1 0 0 0 1 -1a2 2 0 0 0 -2 -2zm-1.99 -4l-.127 .007a1 1 0 0 0 .117 1.993l.127 -.007a1 1 0 0 0 -.117 -1.993zm4 0l-.127 .007a1 1 0 0 0 .117 1.993l.127 -.007a1 1 0 0 0 -.117 -1.993z" stroke-width="0" fill="currentColor" />
        </svg>
      </div>
    </div>
    <div class="empty pt-0">
      <p class="empty-title text-danger">NOT AUTHORIZED</p>
      <p class="empty-subtitle text-secondary">
        What are you trying to do? Whatever it is, its not allowed!
      </p>
    </div>
  </div>
</div>

{:else if showExpiredMessage}
<div class="d-flex justify-content-center align-items-center vh-100" transition:fade={{ delay: 250, duration: 300 }}>
  <div class="card">
    <div class="card-status-top bg-danger"></div>
    <div class="card-body text-center pb-0" style="min-width: 30em;">
      <div class="display-2 fw-bold my-3 text-danger">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon-tabler icon-tabler-balloon-filled" width="96" height="96" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M12 1a7 7 0 0 1 7 7c0 5.457 -3.028 10 -7 10c-3.9 0 -6.89 -4.379 -6.997 -9.703l-.003 -.297l.004 -.24a7 7 0 0 1 6.996 -6.76zm0 4a1 1 0 0 0 0 2l.117 .007a1 1 0 0 1 .883 .993l.007 .117a1 1 0 0 0 1.993 -.117a3 3 0 0 0 -3 -3z" stroke-width="0" fill="currentColor" />
          <path d="M12 16a1 1 0 0 1 .993 .883l.007 .117v1a3 3 0 0 1 -2.824 2.995l-.176 .005h-3a1 1 0 0 0 -.993 .883l-.007 .117a1 1 0 0 1 -2 0a3 3 0 0 1 2.824 -2.995l.176 -.005h3a1 1 0 0 0 .993 -.883l.007 -.117v-1a1 1 0 0 1 1 -1z" stroke-width="0" fill="currentColor" />
        </svg>
      </div>
    </div>
    <div class="empty pt-0">
      <p class="empty-title text-danger">Expired</p>
      <p class="empty-subtitle text-secondary">
        Too late, sorry.
      </p>
    </div>
  </div>
</div>

{:else if requirePassphrase}
<div class="d-flex justify-content-center align-items-center vh-100" transition:fade={{ delay: 250, duration: 300 }}>
  <div class="card">
    <div class="card-status-top bg-danger"></div>
    <div class="card-body text-center pb-0">
      <div class="display-2 fw-bold my-3 text-danger">
        <svg xmlns="http://www.w3.org/2000/svg" class="icon-tabler icon-tabler-circle-key-filled" width="96" height="96" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
          <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
          <path d="M12 2c5.523 0 10 4.477 10 10a10 10 0 0 1 -20 0c0 -5.523 4.477 -10 10 -10zm2 5a3 3 0 0 0 -2.98 2.65l-.015 .174l-.005 .176l.005 .176c.019 .319 .087 .624 .197 .908l.09 .209l-3.5 3.5l-.082 .094a1 1 0 0 0 0 1.226l.083 .094l1.5 1.5l.094 .083a1 1 0 0 0 1.226 0l.094 -.083l.083 -.094a1 1 0 0 0 0 -1.226l-.083 -.094l-.792 -.793l.585 -.585l.793 .792l.094 .083a1 1 0 0 0 1.403 -1.403l-.083 -.094l-.792 -.793l.792 -.792a3 3 0 1 0 1.293 -5.708zm0 2a1 1 0 1 1 0 2a1 1 0 0 1 0 -2z" stroke-width="0" fill="currentColor" />
        </svg>
      </div>
    </div>
    <div class="empty pt-0">
      <p class="empty-title text-danger">Passphrase required</p>
      <p class="empty-subtitle text-secondary">
        Enter the magic word.
      </p>
      <input autofocus type="text" class="form-control w-100 text-teal display-3" bind:value={passphrase} on:keydown={handleKeydown}>
      <div class="empty-action">
        <a class="btn btn-danger" on:click|preventDefault={submitPassphrase}>
          <!-- Download SVG icon from http://tabler-icons.io/i/search -->
          <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-shield-lock" width="44" height="44" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
            <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
            <path d="M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1 -8.5 15a12 12 0 0 1 -8.5 -15a12 12 0 0 0 8.5 -3" />
            <path d="M12 11m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" />
            <path d="M12 12l0 2.5" />
          </svg>
          Unlock
        </a>
      </div>
    </div>
  </div>
</div>


{:else if vulnerability}
<div class="page container mt-4 mb-4">
  <div class="d-flex g-2 align-items-center">
    <div class="h-100 fs-1 br-2 justify-center">{vulnerability?.ID}</div>
    <div class="col">
      <div class="page-pretitle" style="position: relative">Title
  </div>
      <h2 class="page-title">
        {vulnerability?.Vulnerability.title}
      </h2>
    </div>
  </div>
  <div class="row row-deck row-cards mt-2">
    <div class="col12">
      <div class="row row-cards">
        <div class="col-sm-4 col-lg-4">
          <div class="card card-sm">
            <div class="card-body ">
              <div class="row align-items-center">
                <div class="col-auto">
                  <CriticalityDonoutSvg severity={vulnerability?.Vulnerability.criticality}/>
                </div>
                <div class="col">
                  <div class="font-weight-bold text-capitalize">
                    {vulnerability?.Vulnerability.criticality}
                  </div>
                  <div class="text-secondary">Criticality</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="col-sm-4 col-lg-4">
          <div class="card card-sm">
            <div class="card-body ">
              <div class="row align-items-center">
                <div class="col-auto">
                  <span class="bg-teal text-white avatar">
                    <Icon icon="versions-filled" />
                  </span>
                </div>
                <div class="col">
                  <div class="font-weight-medium text-upper text-truncate">
                    {vulnerability?.Vulnerability.category || ""}
                  </div>
                  <div class="text-secondary">Category</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="col-sm-4 col-lg-4">
          <div class="card card-sm">
            <div class="card-body ">
              <div class="row align-items-center">
                <div class="col-auto">
                  <span class="bg-{vulnerability?.Vulnerability.isPublicFacing ? 'pink' : 'muted'} text-white avatar">
                    {#if vulnerability?.Vulnerability.isPublicFacing}
                        <Icon icon="world" />
                        {:else}
                        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-eye-closed" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M21 9c-2.4 2.667 -5.4 4 -9 4c-3.6 0 -6.6 -1.333 -9 -4" /><path d="M3 15l2.5 -3.8" /><path d="M21 14.976l-2.492 -3.776" /><path d="M9 17l.5 -4" /><path d="M15 17l-.5 -4" /></svg>
                        {/if}
                  </span>
                </div>
                <div class="col">
                  <div class="font-weight-medium text-upper">
                    {#if vulnerability?.Vulnerability.isPublicFacing}
                        Internet Facing
                        {:else}
                        Restricted Access
                        {/if}
                  </div>
                  <div class="text-secondary">Network Exposure</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div class="card mt-3">
    <div class="card-body">
      <div class="row">
        <div class="datagrid col-6">

        <div class="datagrid-item">
          <div class="datagrid-title">Date</div>
          <div class="datagrid-content">{vulnerability?.Vulnerability.date}</div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Visibility</div>
          <div class="datagrid-content">
            {#if vulnerability?.Vulnerability.visibility == "public"}
              <Icon icon="world" />
            {:else if vulnerability?.Vulnerability.visibility == "private"}
              <Icon icon="eye-closed" />
            {:else if vulnerability?.Vulnerability.visibility == "hidden"}
              <Icon icon="eye-off" />
            {/if}
            {vulnerability?.Vulnerability.visibility}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Criticality</div>
          <div class="datagrid-content">
            <SeverityBucket severity={vulnerability?.Vulnerability.criticality} />
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Category</div>
          <div class="datagrid-content">{vulnerability?.Vulnerability.category || 'N/A'}</div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Impact</div>
          <div class="datagrid-content">
            <ImpactBucket impact={vulnerability?.Vulnerability.impact} />
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">
            Ease of Exploitation
            <svg on:mouseover={() => showExploitationTooltip = true} on:mouseout={() => showExploitationTooltip = false}  xmlns="http://www.w3.org/2000/svg" class="stroke-normal cursor-pointer icon icon-tabler icon-tabler-info-circle" width="12" height="12" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0" /><path d="M12 9h.01" /><path d="M11 12h1v4h1" /></svg>
          </div>
          <div class="datagrid-content" >
            {vulnerability?.Vulnerability.easeOfExploitation || 'N/A'}
            {#if showExploitationTooltip}
            <div class="dropdown-menu dropdown-menu-demo dropdown-menu-arrow show" style="max-width: 50em">
                <div class="m-2">
                  <div class="row mb-2">
                    <div class="col-2">Trivial</div>
                    <div class="col-10 text-secondary" class:text-info={vulnerability?.Vulnerability.easeOfExploitation === "Trivial"}>This vulnerability can be exploited effortlessly. No specific skills or tools are required.</div>
                  </div>
                  <div class="row mb-2">
                    <div class="col-2">Easy</div>
                    <div class="col-10 text-secondary" class:text-info={vulnerability?.Vulnerability.easeOfExploitation === "Easy"}>Exploitation is straightforward and requires minimal technical knowledge or standard tools.</div>
                  </div>
                  <div class="row mb-2">
                    <div class="col-2">Moderate</div>
                    <div class="col-10 text-secondary" class:text-info={vulnerability?.Vulnerability.easeOfExploitation === "Moderate"}>Exploitation requires a certain level of skill or specialized knowledge. Some custom tools might be needed.</div>
                  </div>
                  <div class="row mb-2">
                    <div class="col-2">Difficult</div>
                    <div class="col-10 text-secondary" class:text-info={vulnerability?.Vulnerability.easeOfExploitation === "Difficult"}>This vulnerability is challenging to exploit. It requires advanced technical skills, detailed knowledge, and possibly custom-created tools.</div>
                  </div>
                  <div class="row mb-2">
                    <div class="col-2">Theoretical</div>
                    <div class="col-10 text-secondary" class:text-info={vulnerability?.Vulnerability.easeOfExploitation === "Theoretical"}>Exploitation is only possible in theory. There are no known practical methods to exploit this vulnerability under normal conditions.</div>
                  </div>

                </div>
            </div>
          {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Endpoint</div>
          <div class="datagrid-content">{vulnerability?.Vulnerability.endpoint || 'N/A'}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Images</div>
          <div class="datagrid-content">
            {#if vulnerability?.Vulnerability.images && vulnerability?.Vulnerability.images.length > 0}
              <div class="avatar-list avatar-list-stacked cursor-pointer">
                {#each vulnerability?.Vulnerability.images as image, index}
                  <img src={`data:image/png;base64,${image}`} alt={`Image ${index}`} on:click={() => openModal(index)} class="avatar avatar-xs rounded"/>
                {/each}
              </div>
            {:else}
              <span>N/A</span>
            {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Slack URL</div>
          <div class="datagrid-content">
            {#if vulnerability.SlackUrl}
              <a href="{replaceUrlWithSlackPrefix(vulnerability.SlackUrl)}">Open in slack <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-external-link" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 6h-6a2 2 0 0 0 -2 2v10a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-6" /><path d="M11 13l9 -9" /><path d="M15 4h5v5" /></svg>
              </a>
            {:else}
              <span>N/A</span>
            {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Location</div>
          <div class="datagrid-content">
            {#if vulnerability?.Vulnerability.isPublicFacing}
            <Icon icon="world" stroke="0.75"/> Internet Facing
            {:else}
            <Icon icon="eye-closed" stroke="0.75" /> Restricted Access
            {/if}
          </div>
        </div>

        <div class="datagrid-item">
          <div class="datagrid-title">Issue tracker</div>
          <div class="datagrid-content">
            {#if vulnerability?.Vulnerability?.issueTrackerURL}
              {#if vulnerability?.Vulnerability?.issueTrackerURL?.includes("gitlab")}
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-brand-gitlab" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M21 14l-9 7l-9 -7l3 -11l3 7h6l3 -7z" /></svg>
              {:else if vulnerability?.Vulnerability?.issueTrackerURL?.includes("azure")}
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-brand-azure" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M6 7.5l-4 9.5h4l6 -15z" /><path d="M22 20l-7 -15l-3 7l4 5l-8 3z" /></svg>
              {:else if vulnerability?.Vulnerability?.issueTrackerURL?.includes("github")}
                <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-brand-github" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 19c-4.3 1.4 -4.3 -2.5 -6 -3m12 5v-3.5c0 -1 .1 -1.4 -.5 -2c2.8 -.3 5.5 -1.4 5.5 -6a4.6 4.6 0 0 0 -1.3 -3.2a4.2 4.2 0 0 0 -.1 -3.2s-1.1 -.3 -3.5 1.3a12.3 12.3 0 0 0 -6.2 0c-2.4 -1.6 -3.5 -1.3 -3.5 -1.3a4.2 4.2 0 0 0 -.1 3.2a4.6 4.6 0 0 0 -1.3 3.2c0 4.6 2.7 5.7 5.5 6c-.6 .6 -.6 1.2 -.5 2v3.5" /></svg>
                {/if}
                <a href="{vulnerability?.Vulnerability.issueTrackerURL}" target="_blank" rel="noopener noreferrer">
                  {extractLastIdFromUrl(vulnerability?.Vulnerability.issueTrackerURL)}
                </a>
            {:else}
              N/A
            {/if}
          </div>
        </div>

      </div>
        <div class="col-6">
          <!-- <Lifecycle bind:activeItem={vulnerability.Status}/> -->
        </div>
      </div>

    <hr />

    <div class="row mb-3 mt-5">
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <!-- svelte-ignore a11y-mouse-events-have-key-events -->
      <div class="col-2" >
        <div class="datagrid-title" style="position: relative">Evidence

        </div>
      </div>
      <div class="col-10">
        <Markdown markdown={vulnerability?.Vulnerability?.evidence} />
      </div>
    </div>
    <div class="row mb-3">
      <div class="col-2">
        <div class="datagrid-title">Remediation</div>
      </div>
      <div class="col-10">
        <Markdown markdown={vulnerability?.Vulnerability?.remediation} />
      </div>
    </div>
  </div>
</div>
</div>

<ImageCarousel images={vulnerability.Vulnerability.images} bind:showModal />
{/if}

<style>
  .stroke-normal{
    stroke-width: 1.75 !important;
  }

  .br-2 {
    border-right: 2px solid #1F2E42;
    margin-right: 10px;
    padding-right: 10px;
  }

  input[type="text"]
{
    font-size:24px;
    font-family: monospace;
    text-align: center;
    margin-top: 20px;
}
</style>