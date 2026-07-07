# Contributing to Hermey

Thanks for your interest in Hermey.

Hermey is early-stage. The most valuable contributions right now are architecture review, Android project scaffolding, security model review, protocol design, and focused implementation PRs.

## Ground Rules

- Be respectful and technical.
- Open an issue before large changes.
- Keep PRs small and reviewable.
- Do not add secrets, tokens, credentials, private endpoints, or personal data.
- Avoid claiming official affiliation with Nous Research unless the project status changes.

## Contribution Workflow

1. Check the existing issues and roadmap.
2. Open an issue for new feature proposals or architecture changes.
3. Fork the repository.
4. Create a topic branch.
5. Make the change with tests or documentation where practical.
6. Open a pull request using the PR template.

## Commit Style

Use short, imperative commit messages:

```text
Add Android project scaffold
Document device binding model
Implement inbox screen skeleton
```

## Pull Request Expectations

A good PR should include:

- What changed
- Why it changed
- How it was tested
- Screenshots or recordings for UI changes
- Security implications for auth, storage, networking, or approvals

## Android Direction

The expected stack is Kotlin-first Android. Specific framework choices will be finalized in the architecture docs as the MVP takes shape.

## Security-Sensitive Changes

Changes involving authentication, device binding, encryption, approval flows, networking, storage, permissions, or agent command execution require extra review. Describe the threat model and failure modes in the PR body.
