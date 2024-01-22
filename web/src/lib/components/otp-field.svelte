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
      } else if (otpCode[index].length === 1 && index < otpLength - 1) {
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
        try{
          const response = await Fetch("/api/session/otp/validate", {method: "POST", body: JSON.stringify({"otp_code": otp})})
          if (!response.error) {
            window.location.href = '/';
          }
          setTimeout(() => {
            submitting = false
            otpCode = Array(otpLength).fill('');
          }, 300);
        } catch {
          submitting = false
        }
    }
  };

    const handlePaste = (event) => {
        const paste = event.clipboardData.getData('text');
        const numbers = paste.split('').filter(char => !isNaN(char) && char !== ' ').slice(0, otpLength);

        if (numbers.length === otpLength) {
            otpCode = numbers.map(num => num); // Create a new array to trigger reactivity

            // Set a slight delay to ensure the DOM updates before submitting
            setTimeout(() => {
                submitOTP();
            }, 0);
        }
    };

  const handleInput = (index, event) => {
    if(event.target.value.length == 6){
      const numbers = event.target.value.split('').filter(char => !isNaN(char) && char !== ' ').slice(0, otpLength);

        if (numbers.length === otpLength) {
            otpCode = numbers.map(num => num); // Create a new array to trigger reactivity

            // Set a slight delay to ensure the DOM updates before submitting
            setTimeout(() => {
                submitOTP();
            }, 0);
        }
    } else {
      otpCode[index] = event.target.value.slice(0, 1);
    }
      submitOTP();
  };

</script>

<div class="row row-cards d-flex justify-content-center" on:paste={handlePaste}>
    {#each Array(otpLength) as _, index}
        <input
            bind:this={otpInputs[index]}
            type="text"
            inputmode="numeric"
            class="form-control qr-code"
            maxlength="6"
            pattern="[0-9]*"
            name="otp"
            autocomplete="one-time-code"
            disabled={submitting}
            value={otpCode[index]}
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
