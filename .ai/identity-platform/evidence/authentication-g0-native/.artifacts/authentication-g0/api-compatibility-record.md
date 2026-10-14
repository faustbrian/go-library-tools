# Original baseline compatibility

Execution revision: 9926ce00a89c4f3187b12b4b25e3268158e138e6 (dirty implementation tree).
Toolchain: go1.26.6 darwin/arm64.
Tool: golang.org/x/exp/cmd/apidiff@v0.0.0-20260718201538-764159d718ef.
Original committed baseline retained as original-api-root.txt.

Executed before replacing api/root.txt:

```
go run "golang.org/x/exp/cmd/apidiff@${APIDIFF_VERSION}" -m -w "$task_cache/current" github.com/faustbrian/go-authentication
go run "golang.org/x/exp/cmd/apidiff@${APIDIFF_VERSION}" -m -incompatible api/root.txt "$task_cache/current"
```

Both exited 0; the incompatibility report was empty. The generated snapshot
was then copied to api/root.txt. This old-to-new comparison is separate from
the ordinary final api gate, which validates the updated source baseline.
Disposable build, module and temporary caches were removed on exit.
