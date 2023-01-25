# Universal Digital Radio Protocol written in Golang
![logo](media/logo.svg)

## Usage
### Run
```bash
cd cmd/udarp
go run *.go config.env
```

### List audio devices
```bash
cd cmd/udarp
go run *.go -l
```

### Decoding with test data<br>
```bash
cd cmd/udarp
./scripts/testStdin.sh < samples/test.raw
```

### RigCtl (Hamlib) https://github.com/Hamlib/Hamlib
 Hamlibs' rigctld is used to control the radios PTT and frequency, and the binaries for it can be found in pkg/txControl/bin, which are embedded into the binary at compile time. UDARP automatically determines the OS and architecture and uses the correct binary to start rigctld.

## TODO:
- [x] Merge soft decoding/viterbi/rs decoding
- [x] Soft decoding
- [x] Merge gfsk generator
- [ ] Stress test with real tranceivers

## Roadmap
- [ ] Web UI
- [ ] Maps in UI in [ui/maps](ui/maps)