# F22 — qctl exits 0 on every error

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H21 (01:44:07Z).

## Minimal reproduction
`results/qctl job get 99999999; echo $?` ; `results/qctl queue frob x; echo $?` (qctl built from the export).

## Expected (oracle)
POSIX/CLI convention: a failed command exits non-zero (README presents qctl as the CLI equivalent of the API; scripts depend on exit status).

## Observed
```
HTTP 404: {"error":{"code":"job_not_found",…}}   qctl_exit=0
unknown queue verb                                qctl_exit=0
```

## Variables isolated
HTTP error (404) and local usage error both exit 0. Message goes to stderr (main.go:17).

## Sibling check
cmd/qctl/main.go:16-18 only. `worker run` returns errors the same way.

## Severity / status
Low: automation cannot detect failed operations. Status: open.
