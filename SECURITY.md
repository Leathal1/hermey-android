# Security Policy

Hermey is intended to become a mobile command surface for agent workflows. Security-sensitive design decisions are part of the core product, not an afterthought.

## Supported Versions

Hermey is in early development and has not yet published a stable release.

| Version | Supported |
| --- | --- |
| `main` | Best effort |

## Reporting a Vulnerability

Please do not report vulnerabilities through public GitHub issues.

Until a dedicated security contact is published, use a private channel with the maintainer. Include:

- A clear description of the issue
- Reproduction steps
- Impact assessment
- Affected files, commits, or flows
- Suggested remediation, if available

## Security-Sensitive Areas

Extra scrutiny is required for changes involving:

- Authentication and session handling
- Device binding and trust decisions
- Local storage and encrypted data
- Push notification payloads
- Agent command approval flows
- Network transport and relay behavior
- Logging, telemetry, and audit trails
- Secrets, tokens, API keys, and credentials

## Disclosure Expectations

Please give the maintainer reasonable time to investigate and remediate before public disclosure.

## Project Security Goals

- No silent execution of sensitive agent actions from mobile notifications
- Clear user confirmation for destructive, financial, privileged, or externally visible actions
- Minimal secret exposure on device
- Secure defaults for local persistence
- Explicit boundaries between mobile client, relay, and agent runtime
