# Universal DigitAl Radio Protocol written in Golang

## Usage
### Decoding with test data<br>
```bash
cd cmd/dhf64
./scripts/testStdin.sh < samples/test.raw
```

### List audio devices
```bash
cd cmd/dhf64
go run *.go -l
```

### Decode from audio device
Replace PLAYBACK_DEVICE and CAPTURE_DEVICE with your device IDs from the list of audio devices
```bash
PLAYBACK_DEVICE=756435816c28459543a90b1bcfd5800a CAPTURE_DEVICE=d26717373e0a8e99f2d549435a7a1f7c go run main.go
```

## TODO:
- [ ] Merge soft decoding/viterbi/rs decoding
- [x] Break up main.go into multiple files
- [ ] Soft decoding<br>
https://www.gaussianwaves.com/2009/12/hard-and-soft-decision-decoding-2/<br>
https://www.tutorialspoint.com/hard-and-soft-decision-decoding<br>
- [x] Merge gfsk generator
- [ ] UI
- [ ] Maps in UI in [ui/maps](ui/maps)
