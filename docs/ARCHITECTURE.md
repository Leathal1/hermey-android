# Architecture

Hermey is planned as an Android-native companion client for Hermes Agent.

## High-Level Shape

```text
Android device
  Hermey app
    UI layer
    local state
    notification handling
    approval workflow
    connection adapter
        |
        v
Hermes Agent endpoint or relay
        |
        v
Agent runtime / integrations
```

## Core Responsibilities

Hermey should own:

- Mobile inbox UX
- Operator notifications
- Local preferences
- Explicit approval decisions
- Device-local session state
- User-visible audit trail for mobile actions

Hermey should not own:

- Long-running agent orchestration
- Tool execution policy
- Backend identity authority
- Source-of-truth task state
- Sensitive action execution without upstream validation

## Security Boundaries

The mobile app is a client, not the authority. Sensitive operations should be validated by the upstream Hermes Agent or relay before execution.

Important boundaries:

- Device identity is not the same as user authorization.
- Push notifications should not contain unnecessary sensitive payloads.
- Approval intent should be explicit and auditable.
- Local storage should avoid long-lived secrets where possible.
- Logs should not leak prompts, credentials, tokens, or private agent output.

## Initial Modules

Expected Android modules or packages:

```text
app/                  Android application shell
core/model/           Shared domain models
core/network/         API client and transport adapters
core/security/        Device trust, secure storage, auth helpers
core/notifications/   Push and local notification handling
feature/inbox/        Message and brief inbox
feature/approval/     Sensitive action approval UI
feature/settings/     Operator settings
```

## Open Design Questions

- Direct Hermes Agent connection vs relay-first connection
- Auth model and device binding strategy
- Notification provider and payload minimization
- Offline behavior for messages and approvals
- Approval schema for sensitive tool actions
- Whether to support multiple Hermes Agent endpoints
