# Universal Digital Radio Protocol written in Golang
![logo](media/logo.svg)

`UDARP` is an advanced digital radio protocol that enables reliable HF communication in noisy environments. Its flexibility offers a wide range of features, including messaging, control, BBS, SMS, ema[...]

---

## Project Status Update 🚦

**August 2025:**  
The UDARP project has not been abandoned! I'm now back working on it again, but development will continue as time allows. Please note that the Maps functionality is currently down, but there's still a plan to bring it back in future updates. Thank you for your continued interest and support!

---

## Announcing the New Maps Service: [map.udarp.com](http://map.udarp.com) 🎉 🥳 🍾
Introducing new, free maps service at [map.udarp.com](http://map.udarp.com)! Visualize UDARP, WSPR, FT4/8, VARAC, JT65 and other transmissions for a deeper understanding of radio spectrum and signal p[...]
![map](media/map_demo.png)

## Join the UDARP Community 🎉
We invite you to become a part of our growing community of developers, enthusiasts, and users who share a passion for UDARP. Connect with like-minded individuals, exchange ideas, discuss features, and[...]

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
 Hamlibs' rigctld is used to control the radios PTT and frequency, and the binaries for it can be found in pkg/txControl/bin, which are embedded into the binary at compile time. UDARP automatically de[...]

## In progress:
- [ ] Real life stress testing

## Complete:
- [x] Merge soft decoding/viterbi/rs decoding
- [x] Soft decoding
- [x] Merge gfsk generator
- [x] Live tracking with maps

## Roadmap
- [ ] Web UI

## Issues
If you encounter any issues, concerns, or simply wish to get in touch, we invite you to join our dedicated [Slack](https://join.slack.com/t/udarp/shared_invite/zt-1sd4e2l39-R2pdafaylJ0uCc7wmhYioQ) and[...]
