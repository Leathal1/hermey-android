# Changelog

All notable changes to Hermey (Hermdroid) will be documented in this file.

## v0.1.0-hermdroid — 2026-07-07

### Added

- **Go core library** (`core/`): gomobile-bindable AAR with auth, REST client,
  SSE streaming, models, and offline cache.
  - `core/auth/` — login, logout, cookie-jar persistence, CSRF/Origin rules
  - `core/client/` — APIClient with 45 typed REST endpoint methods
  - `core/models/` — lenient JSON decoding (json.RawMessage fallbacks)
  - `core/sse/` — SSE parser for 7 event types + heartbeat handling
  - `core/cache/` — bbolt-backed offline cache with message/byte eviction
  - `core/cachebridge/` — JNI bindings for Go↔Kotlin cache access
- **Kotlin/Compose Android app** (`app/`): package `ai.greymattr.hermdroid`
  - Material 3 dynamic color theme, adaptive icon, splash screen
  - Error/empty/loading state composables
  - EncryptedSharedPreferences-backed settings
  - Offline repository backed by bbolt cache
- **Build system**: Makefile (`make aar` → `make apk`), Gradle Kotlin DSL,
  gomobile AAR consumed as local dependency.
- **GitHub Actions CI**: Go tests + gomobile bind + Gradle APK build
  (workflow file pending push — requires `workflow` scope token).
- **Documentation**: ARCHITECTURE.md, ROADMAP.md, BRANDING.md, Play Store
  listing draft, AEGIS review report, postmortem.

### Fixed (AEGIS Round 7)

- `ErrorStates.kt` — missing `is` keyword in `AuthExpired` type check
  (compile error).
- `core/cachebridge/cmd/hermdroidcache/main.go` — JNI string-reference
  leak: `GetStringUTFChars` called twice per parameter (once for GoString,
  once for ReleaseStringUTFChars), leaking the first pointer. Fixed by
  capturing the C string pointer once.
- `app/build.gradle.kts` — `versionName` aligned from `0.6.0` to `0.1.0`.
- `core/cache/cache_test.go` — `TestEviction` used literal `"s%d"` string
  instead of `fmt.Sprintf`, creating only 1 session and never exercising
  eviction. Rewritten with 6 distinct sessions verifying real eviction.

### Security

- No secrets or credentials in APK — all auth is cookie-based at runtime.
- No cleartext HTTP traffic — `usesCleartextTraffic="false"`.
- EncryptedSharedPreferences for sensitive settings.
- Backup exclusion rules (`data_extraction_rules.xml`).

### Known Limitations

- SSE streaming latency not validated end-to-end (no live server
  integration yet — inbox currently reads from offline cache only).
- APK build requires Android SDK/Gradle not available in CI container;
  release APK built locally and attached to GitHub release.
- CI workflow file (`.github/workflows/ci.yml`) not yet on remote due to
  GitHub token lacking `workflow` scope.

## 0.0.0

Initial public namespace reservation.
