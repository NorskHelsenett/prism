<script>
    import { Fetch } from '$lib/fetchUtil';
    import { slide } from 'svelte/transition';

    let authenticating = $state(false);
    let errorMessage = $state('');
    let supported = $state(false);

    // Check if WebAuthn is supported
    if (typeof window !== 'undefined' && window.PublicKeyCredential) {
        supported = true;
    }

    function bufferToBase64url(buffer) {
        const bytes = new Uint8Array(buffer);
        let str = '';
        for (const byte of bytes) {
            str += String.fromCharCode(byte);
        }
        return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    }

    function base64urlToBuffer(base64url) {
        const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
        const padLen = (4 - (base64.length % 4)) % 4;
        const padded = base64 + '='.repeat(padLen);
        const binary = atob(padded);
        const bytes = new Uint8Array(binary.length);
        for (let i = 0; i < binary.length; i++) {
            bytes[i] = binary.charCodeAt(i);
        }
        return bytes.buffer;
    }

    async function authenticateWithPasskey() {
        if (authenticating) return;
        authenticating = true;
        errorMessage = '';

        try {
            // Step 1: Get authentication options from server
            const options = await Fetch('/api/session/passkey/begin', {
                method: 'POST'
            });

            if (!options || options.error) {
                errorMessage = options?.error || 'Failed to start passkey authentication';
                return;
            }

            // Step 2: Convert options for the browser API
            const publicKeyOptions = {
                challenge: base64urlToBuffer(options.publicKey.challenge),
                timeout: options.publicKey.timeout,
                rpId: options.publicKey.rpId,
                allowCredentials: (options.publicKey.allowCredentials || []).map(cred => ({
                    id: base64urlToBuffer(cred.id),
                    type: cred.type,
                    transports: cred.transports
                })),
                userVerification: options.publicKey.userVerification
            };

            // Step 3: Prompt user for passkey
            const assertion = await navigator.credentials.get({
                publicKey: publicKeyOptions
            });

            // Step 4: Send assertion to server
            const body = {
                id: assertion.id,
                rawId: bufferToBase64url(assertion.rawId),
                type: assertion.type,
                response: {
                    authenticatorData: bufferToBase64url(assertion.response.authenticatorData),
                    clientDataJSON: bufferToBase64url(assertion.response.clientDataJSON),
                    signature: bufferToBase64url(assertion.response.signature),
                    userHandle: assertion.response.userHandle
                        ? bufferToBase64url(assertion.response.userHandle)
                        : ''
                }
            };

            const result = await Fetch('/api/session/passkey/finish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (result?.error) {
                errorMessage = result.error;
                return;
            }

            // Success - redirect to home
            window.location.href = '/';
        } catch (err) {
            if (err.name === 'NotAllowedError') {
                errorMessage = 'Authentication was cancelled or timed out.';
            } else {
                console.error('Passkey authentication error:', err);
                errorMessage = 'Passkey authentication failed. Please try again.';
            }
        } finally {
            authenticating = false;
        }
    }
</script>

{#if supported}
<div class="mt-3">
    <div class="hr-text hr-text-center hr-text-spaceless">or</div>
    <div class="card-body mb-3 text-center">
        <button
            class="btn btn-outline-teal w-100 passkey-btn"
            onclick={authenticateWithPasskey}
            disabled={authenticating}
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-fingerprint me-2" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <path d="M18.9 7a8 8 0 0 1 1.1 5v1a6 6 0 0 0 .8 3" />
                <path d="M8 11a4 4 0 0 1 8 0v1a10 10 0 0 0 2 6" />
                <path d="M12 11v2a14 14 0 0 0 2.5 8" />
                <path d="M8 15a18 18 0 0 0 1.8 6" />
                <path d="M4.9 19.5a20 20 0 0 0 .6 1.5" />
                <path d="M6 7a8 8 0 0 1 12 0" />
            </svg>
            {#if authenticating}
                Verifying passkey...
            {:else}
                Verify with Passkey
            {/if}
        </button>
    </div>
</div>

{#if errorMessage}
    <div class="alert alert-danger d-flex align-items-center mt-3" role="alert" transition:slide={{ duration: 200, y: -8 }}>
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
{/if}

<style>
  .passkey-btn {
    max-width: 21em;
    margin-top: 1em;
  }
</style>
