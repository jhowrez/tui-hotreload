
### Sample TUI Option File
``` yaml
command:
  build: go build -gcflags="all=-N -l" -o ./tmp/tui.run ./cmd/test1
  exec: ./tmp/tui.run
watch:
  folders:
    - "."
    - "./cmd/test1"
```

### Sample Debug Option File

``` yaml
command:
  build: go build -gcflags="all=-N -l" -o ./tmp/tui.run ./cmd/test1
  exec: |
    go run github.com/go-delve/delve/cmd/dlv@v1.22.1 \
      exec \
      --headless --api-version=2 --listen=127.0.0.1:43000 \
      --continue --accept-multiclient \
      ./tmp/tui.run
watch:
  folders:
    - "."
    - "./cmd/test1"
```
