# Method: one person's public timeline

A lane-L item is one person id. The job is to produce
`staging/<person-id>/person.yaml`, the record `lobby/schema.md` defines.

Read `lobby/README.md` and `lobby/sources.md` first. The source catalog says
which registers are proven and which are still a discovery item.

## What you collect

Only public-office facts, each from a source you opened in this run:

1. **Representative role.** From the roster payload for each session the person
   appears in: `from` is the session start the record supports, `to` is empty
   while the current session is the latest. Do not invent a start date; a
   session gives a year, and a year is a valid `from`.
2. **Committee seats.** Only when a source names the person and the committee
   with a date. A committee list without a date is a lead, not a role.
3. **Government post.** Join to `internal/data/content/government/ministers.yaml`
   by exact name and post. That file is already dated and sourced; copy its
   `from`, `to` and `source_url` and set `source_tier: A`.
4. **Company roles.** From `data.brreg.no/enhetsregisteret/api/enheter/{orgnr}/roller`.
   Each role carries the person's name, role type, and `fra`/`til` dates. Set
   `role` to `styrerverv` or `daglig_leder`, and copy the 9-digit `orgnr`.
5. **Declared interests.** Only from an official register the lobby-scout has
   marked proven. If the register is a discovery item, do not use it.

## What you never do

- Never write a private citizen, a family member, an address, a phone number,
  a health fact or a private financial fact. The gate rejects the record, and
  the site's promise is the reason.
- Never merge two people with the same name. Match on the person id, or on name
  plus a second identifier (party, constituency, birth year) from the same
  source. A name-only match is a lead.
- Never invent a date. A year from a session is a year; a guessed month is a
  fabrication.
- Never write outside `staging/<person-id>/`.
- Never state that a role change is a "lobby job" or an "influence operation".
  Record the role and the date. The reader draws the conclusion.

## Output

`staging/<person-id>/person.yaml` with the field set in `lobby/schema.md`, plus
`meta.yaml` with `id`, `scope`, `roles_collected`, and `sources_opened`.

Then the driver runs `gate_people.py` and `ekv-person-verifier`. Only a PASS and
a verifier `OK` reach `promote_people.py`.
