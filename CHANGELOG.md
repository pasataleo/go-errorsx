# v0.2.0

## FEATURES

- `Encode` and `Encoded.Decode` serialise an error tree and rebuild it, preserving codes, wrapping, annotations, aggregation, and stack traces rather than collapsing them to a message
- `Encoded` is a plain struct with JSON tags, so it composes into a larger document with `encoding/json` and needs no custom marshaller in either direction
- `WithoutStacks` omits stack frames from an encoded tree, for errors that travel over a wire on every call
- `Frame` and `StackFramer` for errors whose stack frames have already been resolved, as decoded errors are

## IMPROVEMENTS

- `FormatStack` renders stacks from either `StackTracer` or `StackFramer`, so a decoded error still reports the stack captured in the process that created it
- This package still imports nothing outside the standard library, and in particular does not import `encoding/json` — struct tags are inert, so consumers targeting TinyGo or WASM pay nothing for the new API unless they marshal an `Encoded`

<!--
## BUG FIXES
Issues that have been resolved.
-->

<!--
## SECURITY
Vulnerabilities or security-related changes addressed in this release.
-->

<!--
## DEPRECATIONS
Functionality that will be removed in a future release.
-->

<!--
## BREAKING CHANGES
Changes that are not backwards compatible and require updates from consumers.
-->

<!--
## UPGRADE NOTES
Steps required when upgrading from a previous version.
-->
