# Universal Digital Radio Protocol written in Golang
![logo](media/logo.svg)

`UDARP` is an advanced digital radio protocol that enables reliable HF communication in noisy environments. Its flexibility offers a wide range of features, including messaging, control, BBS, SMS, email, and beacons, and is designed to work with low power transmitters. Whether you need to establish communication in remote areas or transmit data over long distances, UDARP provides a powerful and efficient solution.

## Announcing the New Maps Service: [map.udarp.com](http://map.udarp.com) 🎉 🥳 🍾
Introducing new, free maps service at [map.udarp.com](http://map.udarp.com)! Visualize UDARP, WSPR, FT4/8, VARAC, JT65 and other transmissions for a deeper understanding of radio spectrum and signal performance tracking. Ideal for amateurs, professionals, or radio enthusiasts, explore signal data through interactive maps at [map.udarp.com](http://map.udarp.com)!
![map](media/map_demo.png)

## Join the UDARP Community 🎉
We invite you to become a part of our growing community of developers, enthusiasts, and users who share a passion for UDARP. Connect with like-minded individuals, exchange ideas, discuss features, and contribute to the project's growth. To get started, join our vibrant community on [Slack](https://join.slack.com/t/udarp/shared_invite/zt-1sd4e2l39-R2pdafaylJ0uCc7wmhYioQ) and [Groups.io](https://groups.io/g/udarp/signup?u=8269483101481904438). By participating in these platforms, you'll gain access to valuable resources, receive support from fellow members, and stay updated on the latest news and announcements. Don't miss this opportunity to collaborate, learn, and help shape the future of UDARP!

### Slack
https://join.slack.com/t/udarp/shared_invite/zt-1sd4e2l39-R2pdafaylJ0uCc7wmhYioQ

### Groups.io
https://groups.io/g/udarp/signup?u=8269483101481904438

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

## Status for next release
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
If you encounter any issues, concerns, or simply wish to get in touch, we invite you to join our dedicated [Slack](https://join.slack.com/t/udarp/shared_invite/zt-1sd4e2l39-R2pdafaylJ0uCc7wmhYioQ) and [Groups.io](https://groups.io/g/udarp/signup?u=8269483101481904438) spaces.