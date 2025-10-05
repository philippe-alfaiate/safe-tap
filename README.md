# Safe Tap 🔐📲

**Safe Tap** lets you unlock your Linux session securely using your **fingerprint** or **Face ID** from your Android or iOS device — all over your **local network**, with no cloud involved.

## ✨ Features

- 🔒 **Biometric Authentication**  
  Use your phone’s built-in fingerprint or Face ID for instant verification.

- 📝 **Signed Unlock Requests**  
  Each unlock request is cryptographically signed by the mobile app before being sent.

- 🐧 **Linux Integration**  
  A lightweight daemon runs on your Linux machine, verifies the signature, and unlocks your session.

- 🌐 **Local & Private**  
  No external servers. Your data stays on your network.

---

## 🧩 Components

- **📱 Mobile App (Android / iOS)**  
  Handles biometric verification and signs the unlock request.

- **💻 Linux Daemon**  
  Listens for signed messages, verifies them, and triggers the session unlock.

---

## 🚀 Getting Started

### Download project 
```bash
git clone https://github.com/philippe-alfaiate/safetap.git
```


### Use the Linux Daemon
run the daemon:

```bash
cd safetap/linux-daemon
go run main.go db.go
```
build and run the daemon:

```bash
cd safetap/linux-daemon
go build -o safetapd main.go db.go && ./safetapd
```
### Use the linux script as device

Genreate device and register it:
```bash
cd safetap/tools/scriptGeneration
./genFakeDevice.sh Device1 unlock
```

Send verification request to exec unlock script/profil:
```bash
cd safetap/tools/scriptGeneration
./sendVerification.sh Device1 unlock
```

## Todo
- [x] (Daemon) PoC (Proof of Concept) backend
- [x] (Client) Script to test backend
- [ ] (Client/Mobile) PoC mobile client only key/signature **[work in progress]**
- [ ] (Client/mobile) Add fingerprint/Face ID to mobile part
- [ ] (Daemon) Add endpoint for profile selection
- [ ] (Client/mobile) Add registration process and profile selection
- [ ] (Daemon) Add endpoint for device and profil modification
- [ ] (Client/mobile) Add device and profil modification
- [ ] (All) Refactor naming Device/Profile/Endpoint/Struct/Payload/...
- [ ] (Daemon) Use proper DB sqlite or else.
- [ ] [NicetoHave] (Client/mobile) Android App
- [ ] [NicetoHave] (Client/mobile) iOS App (almost never)

## AI - Usage/Assistance

- daemon - github copilot
- pwa - github copilot
- Readme - chatgpt