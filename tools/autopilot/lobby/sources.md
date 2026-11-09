# Sources for lane L, and what each one actually yields

Every row below was probed on 2026-11-09 from this machine. A row marked
**proven** was fetched and its shape read. A row marked **discover** is a
candidate the lobby-scout must prove before the tracker uses it. Do not promote
a discover row to proven without a fetch and a field list in this file.

## Proven

### Stortinget roster
- **Endpoint:** `https://data.stortinget.no/eksport/representanter?format=json&sesjonid=YYYY-YYYY`
- **Yields:** one row per representative. Keys include `id`, `fornavn`,
  `etternavn`, `fylke`, `parti` (`id`, `navn`), `vara_representant`.
- **Use:** the roster and the allowlist. 169 rows for 2025-2026.
- **Note:** the payload carries no `verv`. A person's roles are not here.

### Stortinget hearings
- **Endpoint:** `https://data.stortinget.no/eksport/horinger?format=json&sesjonid=YYYY-YYYY`
- **Yields:** one row per hearing. Keys include `id`, `komite`, `start_dato`,
  `horing_status`, `type`, and `horing_sak_info_liste` (case id, title,
  publication URL).
- **Use:** which committee held a hearing on which case, and when. 351 rows for
  2025-2026.
- **Gap:** the list payload does not carry the invited organisations. The
  per-hearing detail endpoint is not `horing?id=` (404). **Discover.**

### Brønnøysund company roles
- **Endpoint:** `https://data.brreg.no/enhetsregisteret/api/enheter/{orgnr}/roller`
- **Yields:** the roles in one company: role type, person name, birth date,
  `fra`/`til` dates, and the company reference.
- **Use:** company board seats and leadership roles, with dates.
- **Gap:** this is company → persons. A person → companies reverse lookup is not
  this endpoint. **Discover:** the bulk `roller` dataset or a person search.

### Government appointments (already in the repo)
- **File:** `internal/data/content/government/ministers.yaml`
- **Yields:** 245 dated rows with `minister`, `party`, `post`, `department`,
  `from`, `to`, `source_url`.
- **Use:** the minister timeline is already here. Lane L should join to it, not
  re-collect it.

## Discover (prove before use)

| source | what to prove |
|---|---|
| Stortinget "verv og økonomiske interesser" register | the URL per representative, the fields, and whether it is HTML or PDF |
| Person → company reverse lookup | whether Brreg serves it, or only the bulk roller dataset |
| Norwegian lobby register | whether a statutory or voluntary register exists in 2025-2026, and who publishes it |
| Hearing participants | the correct detail endpoint or page that lists invited organisations |
| Stortinget committee membership over time | which endpoint or page gives per-person, per-session committee seats |
| Party donations / `Partiloven` register | whether it names donors or only totals |

## Probed and failed on 2026-11-09 (do not repeat blindly)

These were tried and did not answer. A later scout may retry with a different
route, but record the new result here, not in a chat.

| probe | result |
|---|---|
| `data.stortinget.no/eksport/horing?id=<id>` | 404 |
| `.../eksport/horing?format=json&horingid=<id>` | 404 |
| `.../eksport/horing/<id>?format=json` | 404 |
| `.../eksport/horingdetaljer` and `horingdeltager` | 404 |
| `.../eksport/horinger?id=<id>&format=json` | 200, but `id` is ignored: returns the full list |
| `.../eksport/verv?PersonId=<id>` | 404 |
| `www.stortinget.no/.../Verv-og-okonomiske-interesser/` (two path guesses) | 404 |
| `.../Registrering-av-verv/` | 404 |
| `data.brreg.no/enhetsregisteret/api/opendata` and `/opendata/` | 404 |
| `data.brreg.no/enhetsregisteret/api/roller` and `/roller/{orgnr}` | 404 |
| `data.brreg.no/enhetsregisteret/api/v3/api-docs` and `/v2/api-docs` | 404 |

What still works, and what it means:

- `.../eksport/horinger?format=json&sesjonid=YYYY-YYYY` — the list, not the
  participants. Hearing participants are likely on the HTML hearing page, so the
  next step is an HTML scrape of one hearing, not another API guess.
- `.../eksport/person?format=json&PersonId=<id>` — 200, but keys are only
  `fornavn`, `etternavn`, `id`, `kjoenn`, `foedselsdato`, `doedsdato`. No roles.
- `.../eksport/representanter?format=json&sesjonid=YYYY-YYYY` — 200, the roster.
- `.../eksport/komiteer?format=json&sesjonid=YYYY-YYYY` — 200, committee metadata.
- `data.brreg.no/enhetsregisteret/api/enheter/{orgnr}/roller` — 200, company to
  persons. The only proven role endpoint.

## Hard rules


- A source is proven only when this file lists its live endpoint and its keys.
- A register that answers 404 is not a source. Record the 404 here.
- No press as tier A. Press points to a primary document; cite the document.
- A paywalled or login-only page is not a source for a public record.
