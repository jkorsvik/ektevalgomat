# Blind coding review

This protocol tests the editorial step that source verification cannot test:
whether the evidence supports the coded party position.

## Scope

A review sample is a pilot, not a measure of the whole dataset. Publish the
sample size, selection rule, reviewer types, exact agreement and every
disagreement. Never convert a small pilot into a general reliability claim.

## Blind input

The sampler gives each reviewer the statement, party and source material. It
removes `position`, `context_note`, `scan_summary`, evidence ids and every
other field that can reveal the published interpretation. Reviewers must not
read the repository's evidence files, another reviewer's answers or the key
that maps packets back to cells.

Each reviewer works in a separate context. Use different model families when
the review is performed by agents, and name that limit. Agent review is not
independent human review.

## What to code

Judge the exact statement, not the general topic.

1. Decide whether each evidence row answers the statement for that party.
   A vote against one design does not by itself prove opposition to the
   principle. A minister speaking for the government does not automatically
   speak for a party. Return `null` when the source does not answer the exact
   question.
2. For a row that answers it, use the published scale in
   `tools/autopilot/method/position-scale.md`:
   - `+2` or `-2`: the party's own instrument states the claim or its
     opposite directly, with a concrete commitment, goal or figure.
   - `+1` or `-1`: the direction is clear, but no own instrument states it.
   - `0`: the source is explicitly balanced or split. Missing evidence and
     different positions on different dates are not zero.
3. Give the cell position supported by the newest dated evidence row. Also
   give a confidence of `high`, `medium` or `low`, and a short reason that
   names the decisive source fact.

Open the primary source when the packet does not contain enough text. Never
invent missing context. Mark the row `null` with a low-confidence explanation
when the source cannot be checked.

## Comparison

Compare reviewer A with reviewer B first. Then compare both with the
published cell. Report three counts separately:

- exact agreement between reviewers;
- same direction but a different magnitude;
- disagreement about direction or whether a position exists.

Publish every disputed packet with both reasons. An adjudicator reads the
primary source and either resolves the dispute or records it as unresolved.
Never average two positions and never pick a side because its prose sounds
more confident.

## Reproduction

The original sample, key, both reviewer answers, adjudication and immutable
reviewed-evidence identities are committed under
`tools/audit/data/2026-11-01/`. Its source revision is `0ae4e11`. The audit JSON
records that revision and the SHA-256 of every input. Reviewer answers are
inputs for reproducibility, not independent human labels.

Rebuild and verify the public audit from a clean checkout:

```sh
uv run --with pyyaml python tools/audit/build_coding_review.py
uv run --with pyyaml python tools/audit/verify_coding_review.py
```

To create the same blind sample again, run the sampler in a worktree at the
recorded source revision and require a byte-for-byte match:

```sh
git worktree add /tmp/ekv-coding-review-base 0ae4e11
cd /tmp/ekv-coding-review-base
uv run --with pyyaml python tools/audit/blind_coding_sample.py /tmp/ekv-new-sample \
  --expected tools/audit/data/2026-11-01/blind.json
```

A sample built from another evidence revision is a new sample. The magnitude
buckets depend on positions, so corrections can change the selected cells even
when the seed is unchanged. Never replace the committed original with such a
sample.

`blind.json` is safe to give reviewers. `key.json` stays with the driver until
both reviews are complete. `reviewed-evidence.json` records the evidence IDs,
dates and content hashes the reviewers saw. Future evidence may change a live
cell without changing this historical audit. Only the correction rows
themselves are checked against the current tree.
