# Pilot guide

Status: recruitment materials only. No pilot participant, session, feedback, result, or adoption is claimed here. Completion requires sessions with two real people outside this repository.

## Audience

Safe operators, security reviewers, incident responders, or developers who inspect Safe Transaction Service history and can run Go 1.24 plus `uv` locally.

## Safety and consent

- Use repository synthetic fixtures or the zero-address public-data walkthrough.
- Do not share credentials, API keys, private Safe addresses, private transaction data, calldata, signatures, local paths, or screen contents unrelated to the task.
- Participation is optional. Participant may skip any prompt or stop at any time.
- Ask before taking notes. Offer anonymous notes, attributed notes, or no retained notes.
- Delete notes on request. Do not publish quotes or identity without separate explicit permission.

Consent choice before starting:

- [ ] No notes retained.
- [ ] Anonymous notes retained for product improvement.
- [ ] Attributed notes retained for product improvement. Name or handle: __________

## Ten-minute session

1. State scope: service observations only, not chain truth or signature proof.
2. Open [WALKTHROUGH.md](WALKTHROUGH.md).
3. Run setup and tracked synthetic snapshot extraction.
4. Validate both snapshots with Go and Python.
5. Run offline JSON diff twice and confirm identical output.
6. Ask participant to explain finding and exit code in their own words.
7. Ask feedback prompts below.
8. Reconfirm note consent and remove anything participant declines to retain.

No account, wallet, Safe ownership, network credential, signing action, transaction, server, or database is required.

## Feedback prompts

- Which step was unclear or slow?
- What did you expect before seeing the report?
- Can you explain why finding is observation rather than chain truth?
- Was changed-result exit code `1` understandable?
- What would block use during review or incident response?
- Which metadata would you avoid storing or sharing?
- Would you use this again? Why or why not?

## Blank session note

Leave this template blank until real, consented session occurs. Copy it outside repository per chosen retention policy; do not commit completed participant notes.

```text
Session ID:
Date/time:
Facilitator:
Participant role (optional):
Consent choice:
Environment:
Completed steps:
Elapsed time:
Observed output:
Confusing or slow steps:
Participant explanation of observation vs chain truth:
Feedback:
Follow-up permission:
Redactions/deletions requested:
```

## Outreach copy

```text
Seeking two volunteers for a 10-minute local usability pilot of safe-svc-diff, an open-source tool that compares Safe Transaction Service observations offline. You will use synthetic or zero-address public data only. No wallet, account, credentials, private Safe data, signing, or transaction is needed. Participation is optional; you may choose anonymous notes, attributed notes, or no retained notes, and may stop anytime. This is product feedback, not a security audit or chain-truth verification. Reply privately if interested.
```

## Honest completion record

Do not mark recruitment or pilot complete until two real people participate. Record aggregate outcomes only after sessions happen and consent permits retention. Keep non-response, decline, incomplete session, and negative feedback as valid outcomes; never replace them with invented success.
