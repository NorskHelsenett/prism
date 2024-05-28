<script>
  import { Fetch } from '$lib/fetchUtil';
  import { onMount } from 'svelte';

  let otpInputs = [];
  const otpLength = 6;
  let otpCode = Array(otpLength).fill('');
  let submitting = false;

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
          try {
              const response = await Fetch("/api/session/otp/validate", {
                  method: "POST",
                  body: JSON.stringify({ "otp_code": otp })
              });
              if (!response.error) {
                  window.location.href = '/';
              }
          } catch {
              // Handle error

          } finally {
              submitting = false
              otpCode = Array(otpLength).fill('');
              setTimeout(() => {
                otpInputs[0].focus();
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

<style>
  .qr-code {
      width: 3em;
      margin-right: 0.5em;
      text-align: center;
  }
</style>