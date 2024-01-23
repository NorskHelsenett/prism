<script>
  import { onMount } from 'svelte';
  import SvgQR from '@svelte-put/qr/svg/QR.svelte';
	import { Fetch } from '$lib/fetchUtil';
	import OtpField from '$lib/components/otp-field.svelte';
  let data;
  let alreadyActivated = false

  onMount(async () => {
    data = await Fetch('/api/session/otp/generate');
    if (data?.otp_activated){
      alreadyActivated = data.otp_activated
    }
  });

  const logo = "/favicon.png"
  let warningAccepted = true

</script>

<div class="page page-center">
      <div class="container container-tight py-4">
        <div class="text-center mb-4">
          <a href="." class="navbar-brand">
            <img src="/logo.png" width="110" height="32" alt="PRIME" class="navbar-brand-image">
          </a>
        </div>
        {#if !alreadyActivated && data?.url}
        <div class="card card-md" style="max-height: 80vh;">
          <div class="card-body text-center py-4 p-sm-5">
            <h1 class="mt-5 text-teal">Almost there!</h1>
            <p class="text-secondary">We take security serious, and we've jazzed up our system with OTP for all. You are one of those lucky few who havent enable OTP yet. But dont worry, we'll get you compliant in no time. Just scan this QR code and join our league of the extraordinarily secure.</p>
          </div>
          <div class="hr-text hr-text-center hr-text-spaceless">QR Code</div>
          <div class="card-body mb-3">
            <div class="mb-3 d-flex justify-content-center mh-20">
                <SvgQR class="text-teal bg-transparent " data={data.url} {logo} shape="circle" />
            </div>
          </div>
        </div>
        <div class="row align-items-center mt-3">
          <div class="col-2">
            <div class="progress" hidden>
              <div class="progress-bar" style="width: 50%" role="progressbar" aria-valuenow="25" aria-valuemin="0" aria-valuemax="100" aria-label="25% Complete">
                <span class="visually-hidden">50% Complete</span>
              </div>
            </div>
          </div>
          <div class="col-12">
            <div class="btn-list justify-content-end">

                <button disabled={!warningAccepted} class="btn btn-teal btn-ghost-teal d-none d-sm-inline-block" on:click={() => alreadyActivated = true}>Continue</button>

            </div>
          </div>
        </div>
        {:else}
        <div class="card card-md" style="max-height: 80vh;">

          <div class="card-body text-center py-4 p-sm-5">
            <h1 class="mt-5 text-teal">One-Time Password Required</h1>
            <p class="text-secondary">Please input the provided OTP to gain access to our platform. This added layer of security is in place to protect your account and our sensitive information.</p>
          </div>
          <div class="hr-text hr-text-center hr-text-spaceless">TOTP</div>
          <div class="card-body mb-3">
            <OtpField />
          </div>
          </div>
          {/if}
      </div>
    </div>

<style>
  .mh-20{
    height: 20em;
  }
</style>