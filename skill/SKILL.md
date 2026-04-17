---
name: lmk
description: Show a desktop notification dialog when a task finishes. Use when the user says "lmk when done", "let me know when X finishes", "notify me when this is done", "ping me when ready", "lmk", or any similar phrasing asking to be alerted on completion. Invoke once at the very end of the work, with a short summary of what happened.
---

# lmk — notify on completion

When the user has asked to be notified on completion, run `lmk` **once, after the task is fully done** (whether it succeeded or failed). The dialog pops on the user's desktop; they may be AFK.

## Command

```
lmk -m "<one-line summary>"
```

- Summary should be short (under ~80 chars), factual, and say what finished.
- Include success/failure state if relevant. Use "✅" for success, "❌" for failure.
- If the user said "make sure I see it" / "don't let me miss it" / similar, add `-a` (ack mode) — it re-prompts until the user clicks Ack.

## Examples

User: "run the tests and lmk when done"
→ Run the tests. Then:
```
lmk -m "✅ tests passed (142 ok, 0 failed)"
```
or on failure:
```
lmk -m "❌ tests failed — 3 failures in auth_test.go"
```

User: "fix the flaky test and lmk"
→ Fix it, confirm it works, then:
```
lmk -m "✅ flaky test fixed — rerun 10x clean"
```

User: "lmk when the build's done, don't let me miss it"
→ Run the build, then:
```
lmk -a -m "✅ build done"
```

## Rules

- **Only once, at the end.** Do not run `lmk` between subtasks or as progress updates. The dialog is intrusive.
- **Do not run `lmk`** if the user did not ask for notification — this skill only activates on explicit "lmk" / "notify me" / "let me know" phrasing.
- **Do not use `lmk <command>` form** (i.e. `lmk npm test`) inside the skill — run the command yourself, then call `lmk -m` with the result. You need to see the output to summarize it.
- **Summarize what happened**, not what was asked. "tests passed" is better than "ran the tests you asked for".
