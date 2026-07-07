# Roadmap

Hermey is early-stage. This roadmap defines the first credible path from public namespace to usable Android client.

## Phase 0: Open Source Foundation

- [x] Reserve public repository
- [x] Add README and project positioning
- [x] Add license
- [x] Add contribution guide
- [x] Add code of conduct
- [x] Add security policy
- [x] Add issue and PR templates
- [x] Add initial architecture notes

## Phase 1: Android Scaffold

- [ ] Create Kotlin Android project
- [ ] Establish package namespace
- [ ] Add basic CI for Android build and tests
- [ ] Add dependency management
- [ ] Add formatting and linting
- [ ] Add basic app shell

## Phase 2: Mobile Operator MVP

- [ ] Implement onboarding shell
- [ ] Implement local settings store
- [ ] Implement Hermes Agent connection model
- [ ] Implement mobile inbox view
- [ ] Implement message detail view
- [ ] Implement manual refresh path
- [ ] Add basic notification strategy

## Phase 3: Approval and Control Flows

- [ ] Define sensitive action schema
- [ ] Implement approval cards
- [ ] Add explicit confirm / deny flows
- [ ] Add action audit trail
- [ ] Add failure states and retry behavior

## Phase 4: Security Hardening

- [ ] Threat model mobile client
- [ ] Threat model relay mode
- [ ] Define device binding model
- [ ] Add encrypted local persistence where needed
- [ ] Add secure logging guidelines
- [ ] Review notification payload privacy

## Phase 5: Public Preview

- [ ] Add screenshots
- [ ] Add demo video or GIF
- [ ] Publish first tagged preview release
- [ ] Document install path
- [ ] Collect feedback from Hermes Agent users

## Phase 6: Offline Cache + Store Polish

- [x] Go bbolt offline read-only cache for sessions and messages
- [x] LRU eviction policy by message count and byte size
- [x] Kotlin cache bridge and repository exposing cached data when offline
- [x] Settings screen (server URL, password/token, default model, theme)
- [x] EncryptedSharedPreferences for auth tokens
- [x] Error states: network down, auth expired, stream dropped, server unreachable
- [x] Empty states: no sessions, no skills, no tasks
- [x] Skeleton loading screens
- [x] Material 3 dynamic color theming
- [x] Adaptive launcher icon and splash screen
- [x] Play Store listing draft
