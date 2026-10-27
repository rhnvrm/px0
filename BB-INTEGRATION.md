# BB integration fork

This fork adds opt-in authenticated embedding for the BB px0-review plugin.
Normal launches retain upstream's CSP and localhost-only mutation policy.

`px0 -embed-origin https://bb.example -public-origin https://review.example -no-open /absolute/worktree`
requires a random `PX0_ACCESS_TOKEN` environment value (at least 32 characters).
Only the exact parent origin is allowed by frame-ancestors. The browser-facing
public origin is explicitly allowed for same-origin mutations. Every workspace
request requires a bearer token or authenticated session cookie; an initial
`?px0_token=...` request sets the cookie and redirects to remove the token.
HTTPS cookies are Secure, HttpOnly and partitioned. Serve behind authenticated
BB Connect; never configure wildcard origins. Embedding disables native agent
execution, updates, and telemetry. File/Git editing remains available.

The frontend exchanges messages only with the configured parent origin:
`px0:selection` carries a selected path, line range and bounded snippet;
`px0:ready` signals startup; `bb:theme`, `bb:open` and `bb:refresh` sync BB.
The BB plugin validates both the source iframe and origin before handling them.
Feedback is submitted explicitly through the plugin into the associated BB
thread, preserving its existing provider, model, account and approval policy.
