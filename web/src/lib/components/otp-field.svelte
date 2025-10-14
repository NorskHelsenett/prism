<script>
  import { Fetch } from '$lib/fetchUtil';
  import { onMount } from 'svelte';

    export let requiresActivation = false;

  let otpInputs = [];
  const otpLength = 6;
  let otpCode = Array(otpLength).fill('');
  let submitting = false;
  let errorMessage = '';

  const focusNextInput = (index, event) => {
      if (event.key === 'Backspace' && !otpCode[index]) {
          if (index > 0) {
              otpInputs[index - 1].focus();
          }
      } else if (otpCode[index] && index < otpLength - 1) {
          otpInputs[index + 1].focus();
      }
  };

  onMount(() => {
      if (otpInputs[0]) {
          otpInputs[0].focus();
      }
  });

  const submitOTP = async () => {
      const otp = otpCode.join('');
      if (otp.length === otpLength && !submitting) {
          // Add fetch API call here
          submitting = true
          errorMessage = ''
          let validationSucceeded = false
          try {
              if (requiresActivation) {
                  const activationResponse = await Fetch("/api/session/otp/generate", {
                      method: "PATCH",
                      headers: {
                        "Content-Type": "application/json"
                      },
                      body: JSON.stringify({ "otp_code": otp })
                  });

                  if (activationResponse?.error) {
                      errorMessage = activationResponse.error;
                      if (typeof errorMessage === 'string' && errorMessage.toLowerCase().includes('invalid otp')) {
                          otpCode = Array(otpLength).fill('');
                      }
                      return;
                  }
              }

              const validationResponse = await Fetch("/api/session/otp/validate", {
                  method: "POST",
                  headers: {
                    "Content-Type": "application/json"
                  },
                  body: JSON.stringify({ "otp_code": otp })
              });

              if (validationResponse?.error) {
                  errorMessage = validationResponse.error;
                  if (typeof errorMessage === 'string' && errorMessage.toLowerCase().includes('invalid otp')) {
                      otpCode = Array(otpLength).fill('');
                  }
                  return;
              }

              validationSucceeded = true;
              window.location.href = '/';
          } catch (err) {
              console.error(err);
              errorMessage = 'Failed to verify OTP. Please try again.';
          } finally {
              submitting = false
              if (validationSucceeded) {
                  otpCode = Array(otpLength).fill('');
              }
              setTimeout(() => {
                otpInputs[0]?.focus();
              }, 10);
          }
      }
  };

  const handlePaste = (event) => {
      const paste = event.clipboardData.getData('text');
      const numbers = paste.split('').filter(char => !isNaN(char) && char !== ' ');

      // Distribute the pasted characters into the fields
      otpCode.forEach((_, index) => {
          if (index < numbers.length) {
              otpCode[index] = numbers[index];
              if (otpInputs[index]) {
                  otpInputs[index].value = numbers[index]; // Ensure UI updates
              }
          }
      });

      // Check if the pasted content fills all fields and if so, submit
      if (numbers.length === otpLength) {
          setTimeout(submitOTP, 10);
      }
  };

  const handleInput = (index, event) => {
      otpCode[index] = event.target.value.slice(0, 1);
      errorMessage = '';

      // Focus next input and submit only if all fields are filled
      if (index < otpLength - 1) {
          focusNextInput(index, event);
      }
      if (otpCode.every(num => num)) {
          submitOTP();

      }
  };
</script>

<div class="row row-cards d-flex justify-content-center" on:paste={handlePaste}>
  {#each otpCode as _, index (index)}
      <input
          bind:this={otpInputs[index]}
          type="text"
          inputmode="numeric"
          class="form-control qr-code"
          maxlength="1"
          pattern="[0-9]*"
          name={`otp-${index}`}
          autocomplete="one-time-code"
          disabled={submitting}
          bind:value={otpCode[index]}
          on:keyup={event => focusNextInput(index, event)}
          on:input={event => handleInput(index, event)}
      />
  {/each}
</div>

{#if errorMessage}
    <div class="alert alert-danger d-flex align-items-center mt-3" role="alert">
        <div class="me-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-alert-triangle" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none" />
                <path d="M12 9v4" />
                <path d="M10.363 3.591l-8.106 13.534a1.914 1.914 0 0 0 1.636 2.871h16.214a1.914 1.914 0 0 0 1.636 -2.87l-8.106 -13.536a1.914 1.914 0 0 0 -3.274 0z" />
                <path d="M12 16h.01" />
            </svg>
        </div>
        <div>
            <p class="mb-0">{errorMessage}</p>
        </div>
    </div>
{/if}

<style>
  .qr-code {
      width: 3em;
      margin-right: 0.5em;
      text-align: center;
  }
</style>