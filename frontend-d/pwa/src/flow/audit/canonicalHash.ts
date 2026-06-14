// Client-side reproducible hash of the Intent Handshake canonical
// JSON — must byte-identical match the Go backend's intent.CanonicalJSON()
// + sha256(...) implementation, so judges can see the manifest_hash
// chain consistency in the audit drawer.
//
// Go reference (internal/intent/disclosure.go):
//   type stable struct {
//     ExpectedRequests  []string `json:"expected_requests"`
//     ForbiddenRequests []string `json:"forbidden_requests"`
//     SafetySummary     string   `json:"safety_summary"`
//     Source            string   `json:"source"`
//   }
//   json.Marshal(stable{...}) → sha256 → "0x" + hex
//
// Go's encoding/json:
//   - field order = struct order
//   - no trailing whitespace
//   - HTML-safe escape on < > & by default (we don't have those chars
//     in the manifest, so this doesn't matter in practice)
//   - empty slices encoded as []
//   - no map keys (we use struct fields, deterministic)
//
// We mirror this exactly with manual JSON.stringify so we don't rely
// on JS engine's iteration order being stable across browsers.

import type { IntentManifest } from '../../types/passport'

/** Build the canonical JSON string that mirrors Go's
 *  intent.CanonicalJSON output for an IntentManifest. */
export function canonicalManifestJSON(m: IntentManifest): string {
  // Field order MUST match the Go struct field order:
  //   ExpectedRequests, ForbiddenRequests, SafetySummary, Source
  // Use manual concatenation to guarantee deterministic key ordering.
  const parts: string[] = []
  parts.push('"expected_requests":'  + JSON.stringify(m.expected_requests  ?? []))
  parts.push('"forbidden_requests":' + JSON.stringify(m.forbidden_requests ?? []))
  parts.push('"safety_summary":'     + JSON.stringify(m.safety_summary     ?? ''))
  parts.push('"source":'             + JSON.stringify(m.source             ?? ''))
  return '{' + parts.join(',') + '}'
}

/** Compute sha256 hex of the canonical manifest JSON, prefixed "0x".
 *  Returns null if Web Crypto unavailable (very old browsers). */
export async function manifestSha256Hex(m: IntentManifest): Promise<string | null> {
  if (typeof crypto === 'undefined' || !crypto.subtle) return null
  const canon = canonicalManifestJSON(m)
  const buf = new TextEncoder().encode(canon)
  const digest = await crypto.subtle.digest('SHA-256', buf)
  const bytes = new Uint8Array(digest)
  let hex = '0x'
  for (let i = 0; i < bytes.length; i++) {
    hex += bytes[i].toString(16).padStart(2, '0')
  }
  return hex
}
