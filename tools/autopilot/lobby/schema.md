# The person record

One YAML file per person: `lobby/staging/<person-id>/person.yaml`, promoted to
`internal/data/content/people/<person-id>.yaml`. The gate enforces every rule
below.

```yaml
id: MARNIL                 # the Stortinget person id, or a government-file id
name: "..."                # public name as the official source spells it
scope: representant        # representant | minister | statssekretar
consent_basis: public_office
party_id: v                # optional; the party the person represents now
roles:                     # one entry per dated role, or per role change
  - role: stortingsrepresentant
    organisation: Stortinget
    from: "2025-10-01"     # YYYY or YYYY-MM-DD
    to: ""                 # empty means "still holds it"
    source_url: "..."
    source_tier: A
    retrieved_at: "2026-11-09"
  - role: styrerverv
    organisation: "Company AS"
    orgnr: "123456789"     # required when the role is a company role
    from: "2019-05-02"
    to: "2021-04-30"
    source_url: "https://data.brreg.no/enhetsregisteret/api/enheter/123456789/roller"
    source_tier: A
    retrieved_at: "2026-11-09"
hearings:                  # optional; an organisation heard, not the person
  - case_id: 104908
    committee: ARBSOS
    organisation: "..."
    heard_on: "2025-10-27"
    source_url: "..."
    source_tier: A
    retrieved_at: "2026-11-09"
```

## Invariants the gate checks

1. `scope` is one of `representant`, `minister`, `statssekretar`.
2. `consent_basis` is `public_office`.
3. No key from `scope.yaml:forbidden_fields` appears anywhere in the file.
4. Every role has `role`, `organisation`, `from`, `source_url`, `source_tier`,
   `retrieved_at`.
5. `source_tier` is `A` or `B`.
6. `from` parses to a year, and `to` (when set) sorts after `from`.
7. A company role (`styrerverv`, `daglig_leder`) carries `orgrnr`, nine digits.
8. No role is dated after `SnapshotDate`.
9. `source_url` is fetchable (the gate may skip the fetch with `--no-fetch`,
   but a promoted record must have passed with the fetch on).

## What a role change looks like

A job change is two roles for one person: the old role ends, the new begins.
The record does not have a `changed_job` flag; the dates carry the change, the
same way the konsistens score reads two dates instead of a label.

## What this record is not

It is not a biography, a profile, or a judgement. It is a list of public
functions with dates and sources. A reader can see what a person did and check
it. That is the whole claim.
