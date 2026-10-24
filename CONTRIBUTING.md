# Bidra til Ekte valgomat

Prosjektet er en valgomat der hver partiposisjon bygger på et dokument noen
kan åpne og lese selv. Det som gjør den ærlig er ikke koden, det er at
påstandene kan sjekkes. Derfor er det også den letteste måten å bidra på.

Fire innganger, i stigende innsats:

| Du vil | Tid | Gjør dette |
|---|---|---|
| Se hva som faktisk mangler nå | 2 min | [Det som mangler](/bidra) — generert fra dataene, med ferdig utfylt issue per rad |
| Si fra om en kilde som er tolket feil | 5 min | [Meld fra om feil](https://github.com/jkorsvik/ektevalgomat/issues/new?template=rapporter-kilde.md) |
| Foreslå et utsagn, eller en ny formulering | 10 min | [Foreslå endring](https://github.com/jkorsvik/ektevalgomat/issues/new?template=foresla-utsagn.md) |
| Hente kilder, kode posisjoner eller bygge videre | timer | denne filen |

`/bidra` er ikke en huskeliste noen vedlikeholder for hånd. Den regnes ut fra
dataene ved hver bygging, så den kan ikke påstå at noe mangler som allerede er
gjort, eller la være å nevne noe som faktisk mangler. Radene skiller mellom to
jobber som ser like ut: celler uten noen kilde, der noen må finne en, og celler
der kilden alt ligger på siden men ingen har skrevet hva den sier.

Du trenger ikke GitHub-konto for det første: det står en e-postadresse i
footeren. Men et issue er lettere å følge opp, og det havner i historikken.

## Hvordan dataene er bygget

Alt innhold ligger i `internal/data/content/`, som YAML, ikke i en database:

```
statements/    ett utsagn per fil: teksten, dimensjonen, og hva som står på spill
parties/       de ni stortingspartiene
evidence/      én påstand om én partiposisjon, med kilde
programs/      partiprogrammene, med lenke
budget/        de alternative budsjettene
corrections/   rettelsesloggen, som vises på /om
```

Samme innhold serveres også som filer på `/data` (JSON og CSV, generert fra
disse filene ved hvert kall). Endrer du skjemaet, endrer du det ett sted:
`evidenceColumns` i `internal/server/opendata.go`, som gir CSV-overskriften,
JSON-feltene og tabellen på /data. En test krever at de tre er enige.

En påstand i `evidence/` er den enheten du bidrar med. Den sier hva et parti
har gjort eller ment, på hvilket grunnlag, og hvor det står:

```yaml
id: ev-2025-11-25-venstre-lobby-forslag
statement_id: utsagn-01
party_id: v
date: "2025-11-25"
evidence_type: representantforslag
evidence_class: partiets_begrunnelse
scan_summary: "Venstre fremmet forslag om registreringsmodeller for påvirkning mot Stortinget."
claim_action: >
  Fremmet Dok. 8:53 S (2025–2026) om modeller for registrering av påvirkning
  mot Stortinget og politisk ledelse i regjeringen.
status: fremmet
position: 2
source_tier: A
source_name: "Dok. 8:53 S (2025–2026)"
source_url: "https://www.stortinget.no/..."
verbatim_excerpt: "..."      # valgfritt, men se under
retrieved_at: "2026-09-23"
```

## Reglene som er testet

Dette er ikke smak. Hver av dem har en test som feiler, og testen er navngitt
så du kan lese hvorfor regelen finnes.

- **En påstand uten kilde er ikke en påstand.** `source_url`, `source_name`,
  `source_tier` (A–E) og `retrieved_at` er påkrevd på hver enkelt.
- **Et sitat må være ordrett.** `uv run tools/verify/verify_excerpts.py`
  henter kilden, og krever at utdraget faktisk står der. En omskrevet
  "quoting" blir avvist som `EDITED`, ikke godtatt som sitat.
- **Ingen påstand datert etter øyeblikksbildet.**
  `TestNoEvidenceAfterSnapshotDate`.
- **`position` er −2 til 2, eller utelatt.** Utelatt betyr "ikke kodet", og
  det er en ærlig tilstand, ikke en mangel. `0` betyr "nøytral", ikke "vet
  ikke". `TestKnownUncodedCellsSayNoCodedPosition`.
- **Ikke fyll en celle fordi den er tom.** Saker Stortinget har mottatt, men
  ikke behandlet, skal stå uten posisjon.
  `TestAddendumPendingCasesHaveNoFabricatedPartyPositions`.
- **En rettelse skal i loggen.** Hvis en leser kunne sett det som var feil,
  hører den hjemme i `corrections/` med `kind: rettelse` (eller `nesten` hvis
  den ble stoppet før den nådde siden). Loggen tåler egne feil; den skal ikke
  bare vise de som slapp ut.

## Kildeteksten rundt sitatet

For hvert ordrette sitat lagrer prosjektet noen hundre tegn av kilden rundt
sitatet, med datoen det ble lest og dokumentets SHA-256. Tre grunner, ingen av
dem "vi vil ha en kopi":

- En lenke som dør, sletter beviset. Utdraget bevarer ordene påstanden hviler på.
- Et sitat uten kontekst kan leses i en retning kilden ikke mente.
- en leser som ikke stoler på denne siden skal kunne se teksten uten å klikke.

Det er ikke dokumentet. Hele rapporter, avisartikler og partiprogrammer ligger
fortsatt bare hos dem som eier dem, og lenken til hele kilden står på hvert
kort. Verktøyet som bygger og gjenkontrollerer arkivet ligger i
`tools/archive/` med sin egen README, så du kan bygge det på nytt og sammenligne
med det som serveres. Alt ligger også som fil:
`/data/kildetekst.json` (ett utdrag per påstand, med kilde, dato og SHA-256).

Vil du at et utdrag fjernes, [meld fra](https://github.com/jkorsvik/ektevalgomat/issues/new?template=rapporter-kilde.md);
det tas ut uten at påstanden mister lenken sin.

## Kjøre lokalt

```sh
go build -o /tmp/ektevalgomat ./cmd/server
PORT=8099 ADDR=127.0.0.1 /tmp/ektevalgomat
```

Datasettet ligger innebygd i binæret, så en endring i YAML krever ny bygging.
Ingen database, ingen innlogging, ingen skrivetilgang til noe.

## Gatene, før du åpner en PR

```sh
gofmt -l .                                  # skal være tom
go vet ./...
go test ./...                               # data, scoring, server, web
for f in web/static/js/*.js; do node --check "$f"; done
for t in tools/*.test.js; do node "$t"; done
python3 tools/autopilot/test_gate.py
python3 tools/archive/check_archive.py --archive internal/data/archive
uv run tools/verify/verify_excerpts.py      # henter kilder, krever nett
```

Den siste er den viktigste og den tregeste: den er den eneste som oppdager at
en kilde har flyttet seg eller at et sitat er pyntet på.

## Hvordan en endring blir vurdert

- Én påstand per pull request når det er en kildeendring. Den skal kunne
  leses og avvises alene.
- En påstand som ikke kunne sjekkes blir ikke med. "Ærlig tomt" er et gyldig
  svar og står som det på siden.
- Kodeendringer skal ha en test som feiler uten endringen. Repoet har flere
  vakter som er skrevet fordi noe gikk galt i produksjon; de står i filene
  sine med dato og hendelse.
- Ingenting slettes. En påstand som er feil rettes ved å legge til, og
  rettelsen loggføres.

## Lisens

Koden, utsagnene, partiposisjonene og påstandene er AGPL-3.0. Bidrag faller
under samme lisens.

Kildeteksten i `internal/data/archive/` er ikke det. Der ligger sitater fra
Stortinget, partiene og pressen, og ordene tilhører dem som skrev dem.
Utdragene lagres for at en død lenke ikke skal slette beviset, ikke fordi
prosjektet kan viderelisensiere dem. Derfor har `kildetekst.json` en annen
`licence`-verdi enn de andre filene, og derfor kan et utdrag fjernes på
forespørsel.
