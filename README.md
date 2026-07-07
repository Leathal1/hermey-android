# Hermey

Android-native companion client for Hermes Agent.

Hermex brought Hermes Agent to iOS. Hermey brings the same mobile-first operator experience to Android.

> Status: early development. Public namespace reserved and project foundation in progress.

## What is Hermey?

Hermey is a pocket command surface for Hermes Agent: briefs, messages, approvals, alerts, and agent control from Android.

The goal is not to replace Hermes Agent. The goal is to provide a clean Android-native client for operators who want Hermes available from their phone with strong security defaults and a mobile-first workflow.

## Planned Features

- Hermes Agent mobile inbox
- Push notifications for agent messages
- Approval flows for sensitive actions
- Secure local identity and device binding
- Voice-to-agent command path
- Background sync for briefs and alerts
- Optional encrypted relay mode
- Local-first preferences and session state
- Operator-grade audit trail for approvals

## Project Principles

- **Operator first:** the phone should be a command surface, not a toy demo.
- **Security by default:** sensitive actions require explicit approval and clear context.
- **Mobile-native:** Android UX should feel native, fast, and reliable.
- **Protocol-friendly:** avoid unnecessary coupling so Hermey can evolve with Hermes Agent.
- **Transparent governance:** roadmap, security policy, and contribution rules are public from the start.

## Repository Layout

```text
.github/                 GitHub community health, templates, and workflows
docs/                    Architecture, branding, and project notes
README.md                Project overview
ROADMAP.md               Near-term build plan
CONTRIBUTING.md          Contribution workflow
SECURITY.md              Vulnerability reporting policy
CODE_OF_CONDUCT.md       Community expectations
LICENSE                  Apache-2.0 license
```

Android source code will be added after the MVP architecture is finalized.

## Roadmap

See [ROADMAP.md](ROADMAP.md).

## Contributing

Contributions are welcome once the initial Android scaffold lands. Start with [CONTRIBUTING.md](CONTRIBUTING.md) and open an issue before large design or architecture changes.

## Security

Please do not open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

## Disclaimer

Hermey is an independent Android client for Hermes Agent. It is not affiliated with, endorsed by, or sponsored by Nous Research unless otherwise stated.
