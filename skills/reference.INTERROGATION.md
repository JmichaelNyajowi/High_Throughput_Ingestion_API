# Interrogation

How to interview the user. Shared by `/discovery`, `/plan`, `/design`, and `/care`.

The goal is a shared understanding neither of you had at the start. Not information extraction — decisions, stress-tested.

## Rules

**Always lead with a recommendation.** Never present a neutral menu. State what you'd do and why, in one or two sentences, then ask. The user decides by accepting or correcting, which is far easier than generating an answer cold. A menu makes them do the work you were supposed to do.

**Plain language.** When a question involves a technical term, define it in one plain sentence inside the question. A question the user can't fully parse can't be corrected — they'll default to yes, and the value of asking is lost.

**One question at a time by default.** Later questions genuinely depend on earlier answers; asking five at once means questions 3–5 were written without knowing answers 1–2.

**Batch when it's free.** If the user asks for batching, batch. If the remaining questions are genuinely independent of each other, batch them and say that's what you're doing. **Never batch a dependency chain** — where each answer changes the next question, stay sequential even if asked, and explain why once.

**Look up facts. Ask about decisions.** If it can be found in the filesystem, git history, package files, or the codebase, find it. Never ask the user something you could have read. Only judgment calls come to them.

**Challenge, don't accept.** When an answer conflicts with something established earlier, say so immediately. When a term is fuzzy or doing two jobs, name it. When the code contradicts what they just said, surface the contradiction. An interview that agrees with everything produces nothing.

**Stress-test with concrete scenarios.** Abstract agreement hides disagreement. "So if a customer cancels after partial delivery, the order stays open but the line items close — is that right?" surfaces what "cancellation" actually means.

**Record inline, never batched.** When a term or decision settles, write it down right then. Batching to the end loses the precise wording and the reasoning.

**Don't act until confirmed.** No files written beyond the inline records, no code, no plans, until the user says you've reached shared understanding.

## Ending

Summarise what was settled, in a short list. Ask whether anything is still open. Only then produce the phase's outputs.
