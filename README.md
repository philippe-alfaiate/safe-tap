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

### 1️⃣ Install the Linux Daemon
Clone the repo and build the daemon:

```bash
git clone https://github.com/your-username/safetap.git
cd safetap/linux-daemon
//TODO
```
