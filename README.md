### Motivation
I've been using [air](https://github.com/cosmtrek/air) for every project thus far, but recently I wanted to built feature rich internal application through TUI and since I did not manage to make air work with bubble, this helper tool was born. 

All it does is simply watch for file changes and serve your application through [creack/pty](https://github.com/creack/pty) as defined in your local **app.yaml** file (check configuration below).

### Sample configuration files app.yaml

#### Sample TUI Option File
``` yaml
command:
  build: go build -gcflags="all=-N -l" -o ./tmp/tui.run ./cmd/test1
  exec: ./tmp/tui.run
watch:
  folders:
    - "."
    - "./cmd/test1"
```

#### Sample Debug Option File

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
