# Native mutation input correction

Go1.26.6, root execution revision 9926ce00a89c4f3187b12b4b25e3268158e138e6.
The previous root package fingerprint a1805f6e0ae909b67cc36bf3cea7a80d8e51d552661d63be7fe8cfabfd767e8f
excluded root production/tests because go list reports the exact root .Dir
without the separator used by startswith(root + "/"). Historical native
campaigns did execute complete source, but their incomplete fingerprints do
not establish reuse validity and their 742 result is not current evidence.

Actual gate-input-digest.sh mutation . . was invoked in a task-owned archived
snapshot with normal public dependencies, no replacements, GOWORK=off and
disposable GOCACHE/GOMODCACHE/GOTMPDIR. Source/test probes were comments;
the descendant probe altered apikey/static.go; unrelated control added a
non-package-graph unrelated-path/control.txt. No live package input was probed.

Old script:
- baseline: a1805f6e0ae909b67cc36bf3cea7a80d8e51d552661d63be7fe8cfabfd767e8f
- changed root authority.go: same (incorrect)
- additionally changed root authority_json_profile_test.go: same (incorrect)

Repaired script, exact root directory OR descendants:
- baseline: 0eeb4f787ba0a9bf8536773592edb1d0c719053bbb58ed307eefca47c1bb8672
- changed root production: 0dfe5abb405e25da35480f937f8172a5d68b0d7902fa8205982c0b41bfb99155
- additionally changed root test: 7005297876dc2bd5d5502d7ea7455ae74defaee69697378064a41dbc21f5c4ad
- additionally changed descendant: 2d8d413bcddb82bc09e078f7332a199d264bfb63c6122a889ba4c0fc3c4707c7
- additionally added unrelated path: same 2d8d413bcddb82bc09e078f7332a199d264bfb63c6122a889ba4c0fc3c4707c7

The five directory filters in package_digest now include exact root .Dir
or descendant .Dir. Filename filtering, source graph ownership, test observers,
reserved module boundaries, verifier operators and thresholds are unchanged.
All incomplete historical evidence remains retained, not rewritten as current.
