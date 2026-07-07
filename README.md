# Hermdroid

Android-native companion client for Hermes Agent.

Hermex brought Hermes Agent to iOS. Hermdroid brings the same mobile-first operator experience to Android.

> Status: Phase 6 in progress. Offline cache, settings, error/empty/loading states, Material 3 dynamic color, adaptive icon, splash screen, and Play Store listing draft are now in place.

## What is Hermdroid?

Hermdroid is a pocket command surface for Hermes Agent: briefs, messages, approvals, alerts, and agent control from Android.

## Project Principles

- **Operator first**
- **Security by default**
- **Mobile-native**
- **Protocol-friendly**
- **Transparent governance**

## Repository Layout

```text
.github/                 GitHub community health, templates, and workflows
app/                     Android application
  src/main/
    java/ai/greymattr/hermdroid/
      core/cache/        JNI bridge to Go bbolt cache
      data/local/        EncryptedSharedPreferences settings
      data/repository/   Offline cache repository
      feature/
        common/          Error, empty, and loading skeleton states
        inbox/           Session/message inbox
        settings/        Server URL, token, model, theme
        skills/          Empty skill list placeholder
        tasks/           Empty task list placeholder
      ui/theme/          Material 3 dynamic color theme
core/
  cache/                 Go bbolt session + message cache library
  cachebridge/           gomobile / JNI bindings for the cache
docs/
  ARCHITECTURE.md        Architecture notes
  BRANDING.md            Branding and naming guidelines
  playstore/listing.md   Play Store listing draft
```

## Building

### Android app

1. Install Android Studio and the Android SDK.
2. Open the project.
3. Build the native cache bridge (choose one path):

   **Option A — gomobile AAR (recommended)**
   ```bash
   go install golang.org/x/mobile/cmd/gomobile@latest
   gomobile init
   ./scripts/build-gomobile-aar.sh
   ```

   **Option B — NDK shared libraries + JNI**
   ```bash
   export ANDROID_NDK_HOME=/path/to/ndk
   ./scripts/build-native-cache.sh
   ```
4. Run `./gradlew assembleDebug`.

### Go cache (host tests)

```bash
cd core/cache
go test ./...
```

## Security

Please do not open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

## Disclaimer

Hermdroid is an independent Android client for Hermes Agent. It is not affiliated with, endorsed by, or sponsored by Nous Research unless otherwise stated.
