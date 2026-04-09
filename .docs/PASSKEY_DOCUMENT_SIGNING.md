# Passkey Document Signing Specification

## Overview

This specification describes how WebAuthn passkeys can be used to cryptographically sign vulnerability reports and other documents within PRISM. This provides non-repudiable proof that a specific user approved specific content at a specific time.

## Motivation

- **Accountability**: Prove who signed off on a vulnerability report before it was sent to a client.
- **Tamper detection**: If report content changes after signing, verification will fail.
- **Audit trail**: Stronger than logging "user clicked approve" — backed by a cryptographic assertion from the user's authenticator.
- **No external tooling**: Uses the same passkeys already registered for 2FA, no PGP keys or certificates to manage.

## How It Works

### Signing Flow

1. User triggers "Sign" action on a vulnerability report (or other document).
2. Backend hashes the document content using SHA-256, producing a **content digest**.
3. Backend creates a WebAuthn authentication ceremony where the **challenge** is the content digest.
4. Frontend calls `navigator.credentials.get()` — the user physically interacts with their authenticator (biometric, security key, etc.).
5. The authenticator signs the challenge (content digest) with the private key.
6. Frontend sends the assertion back to the backend.
7. Backend verifies the assertion against the user's stored public key.
8. Backend stores the **signature record**: assertion data, content digest, signer email, credential ID, and timestamp.

### Verification Flow

1. Re-hash the current document content using SHA-256.
2. Compare the new hash against the stored content digest.
3. If they match, the document has not been tampered with since signing.
4. The assertion itself proves the identity of the signer (tied to their registered passkey).

## Data Model

### SignatureRecord

| Field            | Type      | Description                                              |
|------------------|-----------|----------------------------------------------------------|
| `ID`             | uint      | Primary key                                              |
| `DocumentType`   | string    | Type of document (e.g., `vulnerability`, `assessment`)   |
| `DocumentID`     | uint      | ID of the signed document                                |
| `ContentDigest`  | string    | SHA-256 hex digest of the document content at sign time  |
| `SignerEmail`    | string    | Email of the user who signed                             |
| `CredentialID`   | []byte    | WebAuthn credential ID used for the signature            |
| `AuthData`       | []byte    | Authenticator data from the assertion                    |
| `ClientDataJSON` | []byte    | Client data JSON from the assertion                      |
| `Signature`      | []byte    | Raw cryptographic signature from the authenticator       |
| `SignedAt`       | timestamp | Server timestamp of when the signature was recorded      |

## API Endpoints

### `POST /api/sign/:documentType/:documentID/begin`

Starts the signing ceremony.

**Response:**
```json
{
  "publicKey": {
    "challenge": "<base64url-encoded SHA-256 of document content>",
    "rpId": "prism.example.com",
    "allowCredentials": [...],
    "userVerification": "required",
    "timeout": 60000
  }
}
```

### `POST /api/sign/:documentType/:documentID/finish`

Completes the signing ceremony. Receives the authenticator assertion and stores the signature record.

**Request body:** Standard WebAuthn assertion response (same format as passkey 2FA verification).

**Response:**
```json
{
  "message": "Document signed successfully",
  "signatureId": 42,
  "signedAt": "2026-04-09T12:00:00Z"
}
```

### `GET /api/sign/:documentType/:documentID`

Returns all signatures for a document.

**Response:**
```json
{
  "signatures": [
    {
      "id": 42,
      "signerEmail": "alice@example.com",
      "signedAt": "2026-04-09T12:00:00Z",
      "valid": true
    }
  ]
}
```

The `valid` field indicates whether the current document content still matches the digest that was signed. If the document was modified after signing, `valid` will be `false`.

### `GET /api/sign/:documentType/:documentID/:signatureId/verify`

Verifies a specific signature against the current document content.

**Response:**
```json
{
  "valid": true,
  "contentMatch": true,
  "signerEmail": "alice@example.com",
  "signedAt": "2026-04-09T12:00:00Z",
  "credentialId": "<base64url>"
}
```

- `valid`: The cryptographic assertion is valid.
- `contentMatch`: The current document content matches what was signed. If `false`, the document was modified after signing.

## Content Hashing

The content to be hashed depends on the document type:

- **Vulnerability report**: JSON-serialized vulnerability data (the `vulnerability` JSON field from `json_data`), excluding metadata like timestamps and status.
- **Assessment**: JSON-serialized assessment data.

The hashing function must be deterministic — the same content must always produce the same digest. This means:
- JSON keys must be sorted.
- No whitespace differences.
- Use `json.Marshal` on the Go side (produces deterministic output with sorted keys).

## Security Considerations

- **User verification required**: The `userVerification` option is set to `"required"`, ensuring the user must prove presence (biometric, PIN, etc.) — not just that the device is present.
- **Challenge binding**: The challenge is the content hash itself, cryptographically binding the signature to the exact document content.
- **Replay protection**: Each assertion includes a counter from the authenticator, and the server timestamp provides additional context.
- **Origin binding**: The signature is bound to the PRISM origin via WebAuthn — it cannot be replayed on a different site.
- **No private key exposure**: The private key never leaves the authenticator hardware.

## Limitations

- **Not a standard signature format**: The output is a WebAuthn assertion (CBOR/COSE), not PGP, CMS/PKCS#7, or PDF signatures. Verification requires PRISM or a tool that understands WebAuthn assertions.
- **Origin-bound**: Signatures can only be verified against the PRISM relying party. External parties cannot independently verify without the public key and RP configuration.
- **Not legally binding**: In most jurisdictions, this does not constitute a qualified electronic signature under eIDAS, ESIGN, or similar regulations. It is an internal audit mechanism.
- **Credential lifecycle**: If a user's passkey is reset/removed, existing signatures remain verifiable using the stored public key data, but the user cannot create new signatures until they register a new passkey.

## UI Considerations

- A "Sign" button appears on vulnerability reports and assessments when the user has a registered passkey.
- Signed documents display a signature badge showing who signed and when.
- If the document was modified after signing, the badge shows a warning indicating the signature no longer matches the current content.
- Signature history is viewable in a collapsible panel on the document detail page.

## Future Enhancements

- **Multi-signature workflows**: Require multiple users to sign before a report is considered approved.
- **Signature export**: Export signature records in a portable format for external auditing.
- **Countersigning**: Allow a reviewer to countersign after the original author signs.
