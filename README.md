# Universal Digital Radio Protocol written in Golang
![logo](media/logo.svg)

## Usage
### Decoding with test data<br>
```bash
cd cmd/udarp
./scripts/testStdin.sh < samples/test.raw
```

### List audio devices
```bash
cd cmd/udarp
go run *.go -l
```

### Decode from audio device
Replace PLAYBACK_DEVICE and CAPTURE_DEVICE with your device IDs from the list of audio devices
```bash
PLAYBACK_DEVICE=756435816c28459543a90b1bcfd5800a CAPTURE_DEVICE=d26717373e0a8e99f2d549435a7a1f7c go run main.go
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