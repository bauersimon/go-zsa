# ZSA Keymapp API client for Go

https://github.com/zsa/kontroll

Thin Go wrapper client around the Keymapp API for ZSA keyboards. The raw protobuf spec felt a little too less "Go" idomatic for me.

> Just threw this together for now so there might be some problems. I.e. untested on MacOS/LINUX!

## Dev Requirements

- `go`
- `protoc`
- `google.golang.org/protobuf/cmd/protoc-gen-go`
- `google.golang.org/grpc/cmd/protoc-gen-go-grpc`

## Updating the API

```sh
wget https://raw.githubusercontent.com/zsa/kontroll/main/proto/keymapp.proto -o api/keymapp.proto
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative api/keymapp.proto
```

## Plans

- [ ] maybe create the same wrapper API but for Python?
- [ ] configure the client to be non-blocking (i.e. just fire the requests and discard any responses)
- [ ] better error messages with more context, maybe actual error objects (and maybe stacktraces)
- [ ] ability to configure the client with a logger?
- [ ] mappings for actual keyboard keys
  - i.e. `const VOYAGER_LEFT_1x1 = 0` to provide the index for the left Voyager half, first row, first key (from the right) - and so on...
  - can only obtain the LED indices for my own Voyager, not Moonlander or ErgoDox, so getting those from other people would be cool