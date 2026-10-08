# Architecture Decision Records

Design decisions for this repository are recorded as **ADRs** in this
directory (MADR-style, numbered). An accepted ADR is immutable — supersede it,
don't edit it.

- Copy [`template.md`](./template.md) to `NNNN-slug.md`, using the next free
  number, zero-padded (`0001-...`). Never reuse or renumber.
- `status: proposed` while in review; `accepted` once merged or signed off.
- A record that replaces an earlier one names it in `supersedes:` and the
  replaced record's status becomes `superseded`.
- This directory holds **durable design only**. Work state is tracked in Huly;
  GitHub issues are public intake.
