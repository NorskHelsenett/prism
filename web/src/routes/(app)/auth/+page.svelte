<script>
  import { onMount } from 'svelte';
  import SvgQR from '@svelte-put/qr/svg/QR.svelte';
	import { Fetch } from '$lib/fetchUtil';
	import OtpField from '$lib/components/otp-field.svelte';
	import PasskeyAuth from '$lib/components/PasskeyAuth.svelte';
  let otpData = $state();
  let showOtpEntry = $state(false);
  let canGoBackToQr = $state(false);
  let hasPasskeys = $state(false);

  onMount(async () => {
    const redirectShare = localStorage.getItem('redirectToAfterLogin');
    if (redirectShare?.startsWith("/s/") && !redirectShare.startsWith("/+")) {
      window.location.href = redirectShare
    } else if (redirectShare?.startsWith("/+")) {
      localStorage.removeItem('redirectToAfterLogin');
    }

    // Check if user has passkeys registered (for 2FA option)
    const passkeyStatus = await Fetch('/api/session/passkey/has');
    if (passkeyStatus && !passkeyStatus.error) {
      hasPasskeys = passkeyStatus.hasPasskeys;
    }

    otpData = await Fetch('/api/session/otp/generate');
    if (otpData?.otp_activated){
      showOtpEntry = true
      canGoBackToQr = false
    }
  });

  const logo = "/img/favicon.png"
  let warningAccepted = true

  let showSecret = $state(false)

  function addSpacesToText(text){
    return text.replace(/(.{4})/g, '$1 ').trim()
  }
</script>

<div class="page page-center">
      <div class="container container-tight py-4">
        <div class="text-center mb-4">
          <a href="." class="navbar-brand">
            <img src="/img/logo.png" width="110" height="32" alt="PRIME" class="navbar-brand-image">
          </a>
        </div>
        {#if !showOtpEntry && otpData?.url}
        <div class="card card-md" style="max-height: 80vh;">
          <div class="card-body text-center py-4 p-sm-5">
            <h1 class="mt-5 text-teal">Almost there!</h1>
            <p class="text-secondary">We take security serious, and we've jazzed up our system with OTP for all. You are one of those lucky few who havent enable OTP yet. But dont worry, we'll get you compliant in no time. Just scan this QR code and join our league of the extraordinarily secure.</p>
          </div>
          <div class="hr-text hr-text-center hr-text-spaceless">QR Code</div>
          <div class="card-body mb-3">
            <div class="mb-3 d-flex justify-content-center mh-20 cursor-pointer" onclick={() => showSecret =!showSecret}>
                <SvgQR class="text-teal bg-transparent" data={otpData.url} {logo} shape="circle"/>
                {#if showSecret}
                <div id="secret-box" class="card card-body text-teal">
                  <p class="text-secondary">
                    <svg  xmlns="http://www.w3.org/2000/svg"  width="24"  height="24"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="2"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-info-circle"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0" /><path d="M12 9h.01" /><path d="M11 12h1v4h1" /></svg>
                    Enter this code without the spaces</p>
                  <p class="d-flex justify-content-center code-text">{addSpacesToText(otpData.secret)}</p>
                </div>
              {/if}
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
                <button disabled={!warningAccepted} class="btn btn-teal btn-ghost-teal d-none d-sm-inline-block" onclick={() => {
                  showOtpEntry = true;
                  canGoBackToQr = true;
                }}>Continue</button>
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
            <OtpField requiresActivation={!otpData?.otp_activated} />
          </div>
          {#if hasPasskeys && otpData?.otp_activated}
            <PasskeyAuth />
          {/if}
          {#if canGoBackToQr}
            <div class="card-footer d-flex justify-content-start">
              <button class="btn btn-outline-secondary" onclick={() => {
                showSecret = false;
                showOtpEntry = false;
              }}>Back to QR</button>
            </div>
          {/if}
          </div>
          {/if}
      </div>
    </div>

<style>
  .mh-20{
    height: 20em;
  }
  #secret-box{
    position: absolute;
    top: 32%;
    text-align: center;
    left: 0;
    margin: 1em;
  }

  .code-text{
    letter-spacing: 5px;
  }
</style>
