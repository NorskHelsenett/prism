<script>
    import { onMount } from 'svelte';
    import { Fetch } from '$lib/fetchUtil';
    import { toast } from 'svelte-sonner';

    let credentials = [];
    let registering = false;
    let loading = true;

    let supported = false;
    if (typeof window !== 'undefined' && window.PublicKeyCredential) {
        supported = true;
    }

    onMount(async () => {
        await loadCredentials();
    });

    async function loadCredentials() {
        loading = true;
        const result = await Fetch('/api/session/passkey/credentials');
        if (result && !result.error) {
            credentials = result;
        }
        loading = false;
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

    async function registerPasskey() {
        if (registering) return;
        registering = true;

        try {
            // Step 1: Get registration options from server
            const options = await Fetch('/api/session/passkey/register/begin', {
                method: 'POST'
            });

            if (!options || options.error) {
                toast.error(options?.error || 'Failed to start passkey registration');
                return;
            }

            // Step 2: Convert options for the browser API
            const publicKeyOptions = {
                challenge: base64urlToBuffer(options.publicKey.challenge),
                rp: options.publicKey.rp,
                user: {
                    id: base64urlToBuffer(options.publicKey.user.id),
                    name: options.publicKey.user.name,
                    displayName: options.publicKey.user.displayName
                },
                pubKeyCredParams: options.publicKey.pubKeyCredParams,
                timeout: options.publicKey.timeout,
                attestation: options.publicKey.attestation,
                authenticatorSelection: options.publicKey.authenticatorSelection,
                excludeCredentials: (options.publicKey.excludeCredentials || []).map(cred => ({
                    id: base64urlToBuffer(cred.id),
                    type: cred.type,
                    transports: cred.transports
                }))
            };

            // Step 3: Create credential via browser
            const credential = await navigator.credentials.create({
                publicKey: publicKeyOptions
            });

            // Step 4: Send credential to server
            const body = {
                id: credential.id,
                rawId: bufferToBase64url(credential.rawId),
                type: credential.type,
                response: {
                    attestationObject: bufferToBase64url(credential.response.attestationObject),
                    clientDataJSON: bufferToBase64url(credential.response.clientDataJSON)
                }
            };

            const result = await Fetch('/api/session/passkey/register/finish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (result?.error) {
                toast.error(result.error);
                return;
            }

            toast.success('Passkey registered successfully');
            await loadCredentials();
        } catch (err) {
            if (err.name === 'NotAllowedError') {
                toast.error('Registration was cancelled or timed out.');
            } else {
                console.error('Passkey registration error:', err);
                toast.error('Passkey registration failed.');
            }
        } finally {
            registering = false;
        }
    }

    async function deleteCredential(id) {
        const result = await Fetch(`/api/session/passkey/credentials/${id}`, {
            method: 'DELETE'
        });
        if (!result?.error) {
            toast.success('Passkey removed');
            await loadCredentials();
        } else {
            toast.error('Failed to remove passkey');
        }
    }

    function formatDate(dateString) {
        const options = { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false };
        return new Date(dateString).toLocaleDateString('en-US', options).replace(/\//g, '.').replace(',', '');
    }
</script>

<h3 class="card-title mt-4">Passkeys
    {#if credentials.length > 0}
        <span class="badge bg-green-lt">{credentials.length} registered</span>
    {:else}
        <span class="badge bg-secondary-lt">None</span>
    {/if}
</h3>
<p class="card-subtitle">
    Passkeys provide a passwordless, phishing-resistant way to verify your identity as a second factor.
    When MFA is enabled, you can use a passkey instead of a TOTP code to complete your login. Passkeys work with
    biometrics (fingerprint, face), security keys, or your device's screen lock.
</p>

{#if !supported}
<div class="alert alert-warning">
    Your browser does not support passkeys (WebAuthn). Please use a modern browser to register passkeys.
</div>
{:else}

{#if loading}
    <div class="text-secondary">Loading passkeys...</div>
{:else}
    {#if credentials.length > 0}
    <div class="table-responsive">
        <table class="table table-vcenter card-table">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Registered</th>
                    <th class="w-1"></th>
                </tr>
            </thead>
            <tbody>
                {#each credentials as cred}
                <tr>
                    <td>
                        <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-fingerprint me-1" width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                            <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                            <path d="M18.9 7a8 8 0 0 1 1.1 5v1a6 6 0 0 0 .8 3" />
                            <path d="M8 11a4 4 0 0 1 8 0v1a10 10 0 0 0 2 6" />
                            <path d="M12 11v2a14 14 0 0 0 2.5 8" />
                            <path d="M8 15a18 18 0 0 0 1.8 6" />
                            <path d="M4.9 19.5a20 20 0 0 0 .6 1.5" />
                            <path d="M6 7a8 8 0 0 1 12 0" />
                        </svg>
                        {cred.name}
                    </td>
                    <td class="text-secondary">{formatDate(cred.createdAt)}</td>
                    <td>
                        <button class="btn btn-ghost-danger btn-sm" on:click={() => deleteCredential(cred.id)}>Remove</button>
                    </td>
                </tr>
                {/each}
            </tbody>
        </table>
    </div>
    {/if}

    <div class="btn-list justify-content-start mt-2">
        <button
            class="btn btn-outline-teal d-inline-block"
            on:click={registerPasskey}
            disabled={registering}
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-plus me-1" width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                <path d="M12 5l0 14" />
                <path d="M5 12l14 0" />
            </svg>
            {#if registering}
                Registering...
            {:else}
                Register new passkey
            {/if}
        </button>
    </div>
{/if}
{/if}
