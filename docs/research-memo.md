# Ekte valgomat: etterforsknings- og dataplattform for transparens, integritet og utenlandsk påvirkning

> Original research memo this project is built from (PDF supplied
> 2026-09-23). Kept verbatim, not edited, so its own scope — a
> 12-week, 8-9 FTE, 70-90 person-week institutional project — stays
> visible next to what phase 1 (`../README.md`, `roadmap.md`) actually
> descoped down to and why. Do not silently expand phase 1 to match this
> document's scope without a deliberate roadmap decision first.

## Sammendrag

Prosjektet bør bygges som en **åpen evidensdatabase med en valgomat på toppen**, ikke som en klassisk quiz som sammenligner brukeren med partienes egne formuleringer. Grunnregelen bør være:

> **«Transparens og konsekvens må gjelde i alle ledd for et folkestyre.»**

Den operative testen blir derfor:

**verdi → løfte → forslag → votering → penger → gjennomføring → resultat → fordelingsvirkning → konsistens over tid.**

For hvert av de 42 utsagnene opprettes én rad for hvert av de ni stortingspartiene — **378 statement–party-celler** — som igjen lenkes til alle tilgjengelige evidensobjekter. Stortinget bekrefter at de ni partigruppene i perioden 2025–2029 er Ap, FrP, H, SV, Sp, Rødt, MDG, KrF og Venstre. Stortingets åpne datatjeneste gir maskinlesbare saker, representantforslag, publikasjoner, voteringer, voteringsforslag, vedtak og individuelle voteringsresultater; voteringsdata går tilbake til sesjonen 2011–2012.

Den første fasen bør prioritere **demokratisk åpenhet**, fordi kvaliteten på informasjonen om finansiering, lobbyisme, politisk reklame og tredjeparter påvirker hvor godt vi senere kan bedømme de andre åtte dimensjonene. De tre første fullverdige casene bør være **lobbyregister**, **politisk reklame/tredjepartskampanjer** og **utenlandsk påvirkning i sosiale medier**.

Det er allerede to vesentlige funn:

1. Norge har et relativt tydelig regime for direkte partifinansiering: partier kan ikke motta anonyme eller utenlandske bidrag, og fra 1. juli 2025 ble reglene om bakenforliggende givere via enkelte mellomledd skjerpet. Samtidig viser lovarbeidet og det nye Partilovutvalgets mandat at kampanjer fra **tredjeparter og anonyme aktører utenfor partienes formelle økonomi fortsatt er et regulatorisk spørsmål**.

2. FFIs analyse av stortingsvalget 2025 fant reell utenlandsk påvirkningsaktivitet, inkludert den pro-russiske **#OpNorway**, men understreker samtidig hvor vanskelig attribusjon er. FFI fant også et koordinert nettverk av syv anonyme TikTok-profiler som angrep Ap, SV og Rødt, men kunne **ikke** fastslå at aktørene var utenlandske og vurderte at de kunne være norske. Dette er den evidensstandarden prosjektet må kopiere: **narrativ likhet er ikke bevis på finansiering, koordinering eller utenlandsk kontroll.**

Prosjektet bør derfor aldri publisere én «hyklerscore» eller én «utenlandsk påvirket»-etikett. Det bør publisere **beviskjeden**, motbevis, usikkerheten og partiets egen forklaring.

| Leveranse | Første mål |
|---|---|
| Evidensdatabase | 42 × 9 grunnmatrise + alle kildeobjekter |
| Politisk historikk | Program, forslag, votering, budsjett, gjennomføring |
| Pengestrømmer | Partifinansiering, selskaper, eierskap, kampanjer |
| Påvirkningsregister | annonser, tredjeparter, PR, SoMe, attribusjon |
| Resultatdatabase | SSB/NVE/NSM/FFI/Riksrevisjonen-indikatorer |
| Fordelingsanalyse | hvem får gevinst, hvem betaler, tidsprofil |
| Valgomat | verdi-/saksmatch holdt separat fra integritet og effekt |
| Offentlig dokumentasjon | kilde, dato, rådata, metode, korreksjonslogg |

**Snapshot-dato for første versjon bør låses til 23. september 2026.** Senere endringer lagres som nye observasjoner, ikke ved at historikken overskrives.

## Metode og undersøkelsesroadmap

De 42 utsagnene passer inn i de ni ønskede verdidimensjonene uten å omskrive innholdet. Flere bør samtidig få sekundærtagger, fordi for eksempel digital suverenitet også er robusthetspolitikk.

| Dimensjon | Utsagn | Primært spørsmål |
|---|---:|---|
| **Demokratisk åpenhet** | 1–5 | Kan offentligheten se hvem som påvirker hvem, for hvem og med hvilke penger? |
| **Norsk handlefrihet** | 6–11 | Beholder Norge reell beslutnings- og operasjonsevne? |
| **Forvaltning av fellesskapets ressurser / energi** | 12–18 | Hvem kontrollerer ressursene, risikoen og verdiskapingen? |
| **Teknologisk kapasitet** | 19, 20, 24 | Kan kritiske digitale funksjoner og kompetanse opprettholdes under press? |
| **Innovasjon / omstillingsevne** | 21–23 | Bygger staten og markedet ny kapasitet eller bare kjøper ferdige løsninger? |
| **Universell grunnmur** | 25–30 | Fungerer samfunnets grunnleggende tjenester uansett bosted og inntekt? |
| **Barns muligheter** | 31–36 | I hvilken grad kompenserer samfunnet for forhold barn ikke har valgt? |
| **Robusthet / adaptabilitet** | 38, 39, 41, 42 | Tåler systemet bortfall, sjokk og feilantakelser? |
| **Generasjonsansvar** | 37, 40 | Synliggjøres konsekvensene 10–30 år frem? |

Arbeidet bør skje i fire parallelle evidensstrømmer.

**Politisk evidens** starter med programmet for 2025–2029, men programmet regnes bare som et løfte. Deretter hentes Dokument 8-forslag, komitéinnstillinger, alternative forslag, voteringsresultater, budsjettmerknader, bevilgninger, forskrifter og regjeringsvedtak. Stortingets API støtter nettopp saker, publikasjoner, voteringer, vedtak og høringsinnspill.

**Økonomisk evidens** kobler statsbudsjettet og de alternative budsjettene til tiltakene. Stortinget har samlet alternative statsbudsjetter for alle åtte opposisjonspartiene for 2026 og forklarer at disse normalt uttrykkes gjennom partienes egne dokumenter og merknader til budsjettinnstillingene. For Ap eller et annet parti med regjeringsansvar på det aktuelle tidspunktet må regjeringens budsjettforslag og endelig vedtatt budsjett holdes atskilt fra partiprogrammet.

**Resultatevidens** hentes fra SSB, NVE, NSM, FFI, PST og Riksrevisjonen. LOTTE-systemet er særlig viktig for fordeling: LOTTE-Skatt beregner proveny- og fordelingseffekter av endringer i inntekts- og formuesskatt, LOTTE-Arbeid modellerer arbeidstilbud, og LOTTE-Konsum brukes på indirekte skatter. LOTTE-siden dokumenterer et modellsystem, ikke et vanlig åpent sluttbruker-API; arkitekturen bør derfor i første omgang høste **publiserte LOTTE-resultater, budsjettdokumenter og tabeller**, og ikke bygge på en antakelse om fri programmatisk tilgang til selve modellen.

**Makt- og påvirkningsevidens** kobler Partifinansiering, Brønnøysundregistrene, tredjepartskampanjer, annonsebiblioteker og offentlige trussel-/påvirkningsrapporter. Partifinansiering.no publiserer partienes rapporterte regnskap og bidragsytere, mens SSBs valgkampstatistikk viser bidrag over rapporteringsgrensen. I 2025 ble det rapportert over 152 millioner kroner i valgkampbidrag allerede før valgdagen, og SSB rapporterte senere samlede partiinntekter i 2025 på om lag 1,046 milliarder kroner.

Det viktigste metodiske grepet er at **alle partier undersøkes med samme søkeord, tidsvinduer og beviskrav**. Et funn som rammer eget foretrukket parti skal behandles på samme måte som et funn som rammer motparten.

For hver av de 378 cellene brukes denne innhentingssekvensen:

| Prioritet | Spørsmål | Evidens |
|---|---|---|
| A | Hva lovet partiet? | program 2025–29 |
| A | Hva foreslo partiet selv? | Dok. 8, komitéforslag |
| A | Hvordan stemte representantene? | Stortingets voterings-API |
| A | Hva prioriterte de når penger måtte fordeles? | alternativt/regjeringsbudsjett |
| A | Ble politikk vedtatt? | lover, forskrifter, anmodningsvedtak |
| B | Ble vedtaket faktisk implementert? | departement, direktorat, tildelingsbrev |
| B | Hva ble målbar effekt? | SSB, NVE, Riksrevisjonen, sektorindikator |
| B | Hvem fikk gevinst og kostnad? | LOTTE, SSB, fordelingsanalyse |
| C | Hvilke interesser påvirket prosessen? | høringer, lobbydata, økonomiske forbindelser |
| C | Finnes motstridende evidens? | eldre voteringer, endret program, forklaring |

**Integritet må måles separat fra politisk enighet.** Et parti kan være svært konsekvent i en politikk brukeren er uenig i. Det bør gi høy dokumentert konsistens, men lav saksmatch. Omvendt kan et parti matche brukerens verdi i programmet, men ha svak samsvar mellom løfte og handling.

Derfor bør sluttproduktet minst vise tre separate mål:

**Saksmatch**: hva partiet faktisk mener/har gjort mot brukerens svar.
**Konsistens**: samsvarer prinsipp, program, stemmegivning og handling over tid?
**Evidensstyrke**: hvor sikkert vet vi dette?

Ingen av dem bør skjules i én samlet prosent.

## Datamodell og første innsamling

En relasjonsmodell er bedre enn ett enormt regneark. Kjernen kan være PostgreSQL, mens rå HTML, PDF, JSON og annonsekreativer lagres i et uforanderlig kildearkiv med hentetidspunkt og SHA-256-hash.

> **Fase 1-avvik**: `../README.md` forklarer hvorfor phase 1 ikke bruker
> PostgreSQL eller noen mutable database i det hele tatt — evidens ligger
> som versjonskontrollert YAML, og git-historikken er rettelsesloggen.

Den sentrale evidenstabellen kan se slik ut:

| Felt | Type | Betydning |
|---|---|---|
| `evidence_id` | UUID | unik observasjon |
| `statement_id` | FK | ett av de 42 utsagnene |
| `dimension_id` | FK | én av ni primærdimensjoner |
| `party_id` | FK/null | Ap/FrP/H/SV/Sp/R/MDG/KrF/V |
| `actor_id` | FK/null | selskap, organisasjon, konto, myndighet |
| `date_from` / `date_to` | date | hendelsesperioden |
| `evidence_type` | enum | program, proposal, vote, budget, implementation, outcome osv. |
| `claim_action` | text | presis normalisert beskrivelse |
| `verbatim_excerpt` | text | kort kildeutdrag der nødvendig |
| `source_id` | FK | kilde og arkivkopi |
| `source_tier` | enum | primær, sekundær, OSINT-lead |
| `status` | enum | proposed/adopted/rejected/implemented/unknown |
| `position` | -2…+2/null | kodet holdning til selve utsagnet |
| `beneficiaries` | array/FK | grupper med beregnet gevinst |
| `cost_bearers` | array/FK | grupper med kostnad/risiko |
| `amount_nok` | numeric/null | dokumentert budsjett-/finansieringsbeløp |
| `foreign_dependency` | enum | none/low/medium/high/unknown |
| `time_horizon` | enum | kort/mellom/lang/generasjon |
| `consistency_flag` | enum | se nedenfor |
| `attribution_confidence` | enum/null | særskilt for påvirkningssaker |
| `claim_confidence` | enum | dokumentert/sannsynlig/usikkert/ukjent |
| `counterevidence_ids` | FK[] | materiale som trekker motsatt vei |
| `reviewer_1` / `reviewer_2` | FK | uavhengig dobbelkoding |
| `review_status` | enum | machine/manual/double-reviewed |
| `retrieved_at` | timestamp | når kilden ble hentet |

Rundt denne tabellen bør det ligge egne tabeller for `party`, `statement`, `source_archive`, `vote`, `budget_measure`, `entity`, `ownership_edge`, `financial_transfer`, `campaign`, `advertisement`, `social_account`, `influence_case`, `outcome_indicator` og `distributional_effect`.

**Eierskap må modelleres som en graf, ikke ett `owner`-felt.** Brønnøysundregistrene tilbyr åpne API-er og datasett for virksomheter og roller. Registeret over reelle rettighetshavere har også maskinelle grensesnitt, men tilgangsnivået varierer: Brønnøysund opplyser blant annet at totalbestanden og oppslag på fødselsnummer er begrenset, mens virksomhetsoppslag og API-funksjoner følger egne tilgangsregler. Dette må avklares før automatisering; databasen må ikke anta at alle eierskapsopplysninger er fritt tilgjengelige.

**Prioritert innhøstingsliste**

| Pri. | Kilde | Hva hentes først | Tilgang |
|---|---|---|---|
| A | [Stortingets åpne data](https://data.stortinget.no/) | saker, Dok. 8, voteringer, voteringsresultater, vedtak, publikasjoner, høringsinnspill | API/XML/JSON |
| A | [Partifinansiering](https://www.partifinansiering.no/) | bidrag, givere, årsregnskap, partiledd | offentlig portal/data |
| A | [SSB partifinansiering](https://www.ssb.no/valg/partifinansiering) | historiske tabeller og valgkampbidrag | SSB |
| A | [Brønnøysund: datasett og API](https://www.brreg.no/bruke-data-fra-bronnoysundregistrene/datasett-og-api/) | org.nr., roller, status, virksomhetsdata | API/fullfiler |
| A | [Register over reelle rettighetshavere](https://www.brreg.no/bruke-data-fra-bronnoysundregistrene/datasett-og-api/data-om-reelle-rettighetshavere/) | eier-/kontrollkjeder der tilgang tillater | API, delvis begrenset |
| A | [Stortinget: alternative budsjetter 2026](https://www.stortinget.no/no/Saker-og-publikasjoner/Statsbudsjettet/statsbudsjettet-2026/alternative-statsbudsjetter/) | åtte opposisjonspartiers prioriteringer | dokumenter |
| A | Offisielle partinettsteder | alle ni programmer 2025–29, med versjonsdato | HTML/PDF |
| B | [SSB LOTTE](https://www.ssb.no/forskning/arbeidsmarked-og-skatt/skatteberegningsmodellen-lotte) | publiserte fordelingsberegninger | rapporter/tabeller |
| B | [FFI](https://www.ffi.no/publikasjoner/arkiv/okende-komplekst-og-uoversiktlig-kartlegging-av-utenlandsk-pavirkning-i-forbindelse-med-stortingsvalget-2025) | påvirkningscaser, metoder, aktørvurdering | rapport |
| B | [NSM Risiko 2026](https://nsm.no/regelverk-og-hjelp/rapporter/risiko-2026) | sårbarheter, eierskap, leverandør-/skyavhengighet | rapport |
| B | [PST Nasjonal trusselvurdering 2026](https://www.pst.no/wp-content/uploads/2026/02/Nasjonal-trusselvurdering-2026.pdf) | statlige påvirknings-/økonomiske virkemidler | rapport |
| B | [NVE langsiktig kraftmarkedsanalyse](https://www.nve.no/energi/analyser-og-statistikk/langsiktig-kraftmarkedsanalyse/langsiktig-kraftmarkedsanalyse-2025/) | kraftbalanse, produksjon, priser, scenarioer | rapport/data |
| B | [Riksrevisjonens rapporter](https://www.riksrevisjonen.no/rapporter/) | implementerings- og resultatrevisjon | rapportarkiv |
| C | [Meta Transparency Center](https://transparency.meta.com/reports/) | Ad Library-data og arkiverte annonser | søk/API der tilgjengelig |
| C | [TikTok Commercial Content Library](https://library.tiktok.com/) | annonser/promotert innhold og metadata | offentlig bibliotek |
| C | [Google Ads Transparency Center](https://adstransparency.google.com/) | annonser, advertiser-identitet og kreativ | søk |
| C | EU/EØS politisk-reklame-regime | sponsor, betaler, målretting, finansieringsopprinnelse | regulatoriske data |

Meta beskriver sitt Ad Library som en søkbar database over annonser på Facebook og Instagram. TikTok opplyser at annonser og annet kommersielt innhold kan søkes via Commercial Content Library. Tilgjengelighet, norsk historisk dekning og feltnivå må likevel testes systematisk før datakildene gis høy evidensvekt.

For EU-reglene er det viktig å ikke blande sammen **DSA Transparency Database**, som primært handler om modereringsbeslutninger, med annonseringsregistre. Forordning (EU) 2024/900 om politisk reklame krever blant annet informasjon om sponsor/betaler, pengenes opprinnelse og målretting, og legger opp til et europeisk register for politiske nettannonser. Reglene har vært gjeldende i EU siden 10. oktober 2025, men den norske regjeringens EØS-notat, oppdatert 21. august 2026, oppgir at forordningen fortsatt var til vurdering for innlemmelse i EØS/norsk rett.

Dette betyr at prosjektet ikke skal late som Norge allerede har et komplett offentlig EU-politisk-annonsearkiv. I stedet bør innhøstingen være modulær slik at registeret kobles inn straks norsk/EØS-rett og teknisk dekning tillater det.

## Første evidens i de tre prioritetssakene

> Fase 1's `internal/data/content/evidence/*.yaml` implementerer nøyaktig
> denne seed-tabellen, felt for felt — se `../README.md`.

Følgende er en **første seed-tabell**, ikke sluttvurderinger av partiene. Den viser hvordan råbevis skal lagres før den komplette ni-parti-kodingen.

| Sak | Dato | Parti/aktør | Evidenstype | Dokumentert handling/funn | Status | Tillit |
|---|---|---|---|---|---|---|
| Lobbyregister | 25.11.2025 | Venstre | Representantforslag | Abid Raja, Grunde Almeland og Guri Melby fremmet Dok. 8:53 S om modeller for registrering av påvirkning mot Stortinget og politisk ledelse i regjeringen | Fremmet | Høy |
| Lobbyregister | 26.03.2026 | Stortingets presidentskap | Innstilling | Anbefalte at behovet for og mulig utforming av lobbyregister skulle utredes | Innstilt | Høy |
| Lobbyregister | 21.04.2026 | Stortinget | Vedtak | Vedtok å be presidentskapet sørge for utredning av behovet for og mulig utforming av et register | Vedtatt | Høy |
| Lobbyregister | 21.04.2026 | Rødt | Voteringsspor | Eget forslag nr. 1 ble votert over og falt; eksakt innhold skal kobles fra voteringsforslaget før politisk posisjon kodes | Forkastet | Høy på hendelsen, ennå ukodet posisjon |
| Lobbyregister | 21.04.2026 | MDG | Voteringsspor | Tre egne forslag, nr. 2–4, ble votert over og falt; eksakt tekst må kobles før posisjon kodes | Forkastet | Høy på hendelsen, ennå ukodet posisjon |
| Politisk reklame | 01.07.2025 | Norge | Lovendring | Reglene om bakenforliggende givere til partier ble skjerpet; partier må i definerte tilfeller kjenne og rapportere hvem som står bak støtte via mellomledd | Gjeldende | Høy |
| Politisk reklame | 19.12.2025 | Partilovutvalget | Mandat | Skal eksplisitt drøfte regulering av valgkamp fra tredjeparter og anonyme aktører | Pågående, NOU innen 01.06.2027 | Høy |
| Politisk reklame | 21.08.2026 | EU/EØS/Norge | Rettsstatus | EU-reglene om politisk reklame gjelder i EU, men var fortsatt til vurdering for EØS/norsk innlemmelse | Ikke ferdig norsk implementert | Høy |
| Utenlandsk SoMe-påvirkning | 2025-valget | #OpNorway / pro-russisk hackergruppe | FFI-attribusjon | FFI fant en operasjon eksplisitt knyttet til det norske valget, med Telegram/X-aktivitet og cyberaktivitet | Fremmed påvirkningsforsøk dokumentert | Høy |
| Utenlandsk SoMe-påvirkning | 2025-valget | #OpNorway | Effekt | FFI fant ikke at operasjonen hadde effekt på valgresultatet | Ingen påvist valgeffekt | Høy |
| Utenlandsk SoMe-påvirkning | 2025-valget | 7 anonyme TikTok-profiler | Nettverksanalyse | Koordinert uautentisk aktivitet rettet mot Ap, SV og Rødt med innvandrings-/islaminnhold | Koordinering dokumentert | Høy |
| Utenlandsk SoMe-påvirkning | 2025-valget | samme profiler | Attribusjon | FFI kunne ikke knytte kontoene til en fremmed stat og opplyser at aktørene kan ha vært norske | Opprinnelse ukjent | Høy |
| Utenlandsk SoMe-påvirkning | 2024–25 | russiske aktører | FFI-monitorering | Omtale av Norge økte; FFI observerte narrativer rettet mot støtte til Ukraina, konflikt og mistillit | Aktivitet dokumentert | Høy |
| Utenlandsk SoMe-påvirkning | 2024–25 | kinesiske aktører | FFI-monitorering | Stor vekst i norskspråklig materiale, inkludert profesjonelt innhold og betalt annonsering, men FFI fant ikke belegg for et direkte kinesisk forsøk på å påvirke stortingsvalget | Aktivitet dokumentert; valgintensjon ikke dokumentert | Høy |

Lobbyregistercaset er allerede spesielt velegnet som første «integritetstest». Venstres opprinnelige forslag ba om modeller både for personer som oppsøker Stortinget og personer som oppsøker politisk ledelse i departementene/Statsministerens kontor for å påvirke saker. Presidentskapet endte med en bredere utredning av behov og mulig utforming, ikke et umiddelbart vedtak om å opprette et register.

Voteringsoversikten viser samtidig at Rødt hadde ett alternativt forslag og MDG tre, og at disse falt, mens presidentskapets tilråding ble vedtatt. Før prosjektet koder «for/mot lobbyregister» må **ordlyden i hvert alternativt forslag og de individuelle stemmene hentes gjennom API-et**. At et parti stemmer mot ett forslag kan ellers feilaktig tolkes som motstand mot prinsippet når uenigheten egentlig gjelder utforming. Dette er et godt eksempel på hvorfor en enkel avstemningsrobot ikke er nok.

Partifinansieringen viser samtidig hvorfor politisk reklame må være et eget spor. Partiloven forbyr direkte anonyme og utenlandske partibidrag, og 2025-endringen søker å hindre at bestemte typer mellomledd skjuler den bakenforliggende finansieringen. Men proposisjonen forklarer også at partiloven er bygget rundt bidrag **til partier**, mens organisasjoner kan ha egne politiske aktiviteter; Partilovutvalgets mandat om tredjeparter bekrefter at dette området ennå ikke er ferdig regulert.

Den umiddelbare maskinelle arbeidslisten bør derfor være:

**Stortinget først:** last inn alle saker og publikasjoner fra 2011–2026 med ordstammer som `lobby`, `partifinans*`, `politisk reklame`, `valgkamp`, `bidrag`, `bakenforliggende giver`, `utenlandsk påvirkning`, `desinformasjon`, `sosiale medier`, `habilitet`, `karantene`, `interessekonflikt`, `åpenhet`. Koble deretter `sak_id → votering_id → voteringsforslag → voteringsresultat → representant → parti`. Stortinget dokumenterer egne endepunkter for både voteringer og høringsinnspill.

**Deretter programmer og budsjetter:** parse alle ni programmer og åtte alternative 2026-budsjetter, lag tekstanker rundt hvert av de 42 utsagnene og gjør manuell validering av hvert maskinfunn.

**Så penger og aktører:** match organisasjonsnummer fra Partifinansiering og kampanje-/annonseavsendere mot Brønnøysund, roller og tilgjengelige reelle rettighetshavere. Brønnøysund opplyser at virksomhetsdata kan hentes via API og komplette datasett, med sanntidsoppslag og regelmessig oppdaterte filer.

**Til slutt effekt:** knytt hvert gjennomført tiltak til en på forhånd definert resultatindikator. Riksrevisjonens rapportarkiv kan filtreres etter blant annet digitalisering/IKT, energi, familie/barn, forsvar/beredskap, helse, samferdsel og utdanning, hvilket treffer nesten alle de ni dimensjonene.

## Attribusjon, jus og kvalitetssikring

Den største metodiske risikoen er å forveksle **politisk syn**, **utenlandsk nytte**, **utenlandsk forbindelse**, **finansiering**, **koordinering** og **statlig styring**.

De må være forskjellige databasefelt.

Et norskt parti, en organisasjon eller en privatperson kan argumentere for en politikk Russland, Kina, USA, EU eller en hvilken som helst annen ekstern aktør ønsker, uten at det foreligger noen fremmed påvirkning. Omvendt kan en utenlandsk aktør finansiere eller distribuere innhold som i seg selv er sant. Det er forbindelsen og metoden som må attribueres, ikke om budskapet oppleves «utenlandsk».

FFIs 2025-analyse er et godt mønster. Forskerne fant både russisk aktivitet, sterkt økt norskspråklig kinesisk aktivitet og koordinert TikTok-aktivitet, men skilte eksplisitt mellom disse og det de faktisk kunne attribuere til et fremmed valgrettet forsøk. Plattformenes manglende transparens gjorde enkelte attribusjoner vanskelige eller umulige.

PSTs Nasjonale trusselvurdering 2026 gir samtidig grunnlag for å ta påvirkning alvorlig: PST forventer at autoritære stater vil gjennomføre påvirkningsoperasjoner i Norge, også med interesse for personer med formell eller uformell politisk innflytelse. PST beskriver også hvordan skjulte pengestrømmer, investeringer og økonomiske virkemidler kan brukes til teknologiinnhenting, beslutningspåvirkning og avhengighet.

NSM peker tilsvarende på at eierskap, økonomisk aktivitet og leverandørforhold kan gi tilgang til eller påvirkning over viktige verdier, og at sterk bruk av utenlandske skytjenester kan innebære tap av kontroll over infrastruktur, systemer og data.

Attribusjonsmodellen bør derfor kreve følgende:

| Nivå | Hva kan publiseres | Minimumsevidens |
|---|---|---|
| **Ukjent** | «Opprinnelse ikke fastslått» | innhold/konto observert |
| **Indikasjon på koordinering** | «Atferden fremstår koordinert» | synkronitet, identisk materiale, nettverksmønster |
| **Aktørkobling** | «Knyttet til aktør/organisasjon X» | teknisk, økonomisk, kontraktuell eller konto-/domeneevidens |
| **Sannsynlig utenlandsk kobling** | «Sterk evidens for forbindelse til X» | minst to uavhengige evidensklasser |
| **Høy attribusjon** | «Operasjonen kan med høy tillit attribueres til X» | direkte finansiell/teknisk evidens + uavhengig bekreftelse eller autoritativ attribusjon |
| **Bekreftet statlig styring/finansiering** | uttrykkelig påstand om statlig forbindelse | dokumentasjon som faktisk viser styring/finansiering, ikke bare interessefellesskap |

Evidensklassene bør blant annet omfatte **betalingsspor, kontrakter, annonsebetaler, virksomhet/eierskap, felles teknisk infrastruktur, domeneregistrering og sporings-ID-er, plattformattribusjon, autentiserte interne dokumenter, offisiell FFI/PST/NSM-attribusjon, sanksjonsopplysninger og offentlig erkjennelse**. Språklig/narrativ likhet kan brukes til å oppdage et spor, men skal aldri alene være tilstrekkelig for aktørattribusjon.

Videre bør attribusjon deles opp i seks uavhengige spørsmål:

`hvem produserte innholdet?`
`hvem betalte?`
`hvem distribuerte/amplifiserte?`
`fantes koordinering?`
`fantes en statlig forbindelse?`
`hvilken effekt kan faktisk dokumenteres?`

Det hindrer for eksempel at «russisk finansiering» automatisk blir omskrevet til «russisk styring», eller at stort publikum automatisk tolkes som endrede stemmer.

**Konsistensflagget** bør på samme måte unngå moralske etiketter:

| Verdi | Betydning |
|---|---|
| `consistent` | uttalt prinsipp og handling peker samme vei |
| `tension` | handlingen skaper en reell, men forklarbar spenning |
| `contradiction` | to sammenlignbare handlinger/prinsipper peker direkte motsatt |
| `changed_position` | partiet har eksplisitt endret standpunkt |
| `coalition_constraint` | program og faktisk kompromiss kan skilles dokumentert |
| `insufficient_evidence` | ikke grunnlag for konklusjon |
| `not_comparable` | sakene ligner, men premissene er for forskjellige |

En «motsetning» bør bare settes etter **manuell dobbelgjennomgang og søk etter partiets egen begrunnelse**. Endret politikk er ikke i seg selv hykleri; det kan skyldes nye fakta eller legitim politisk revurdering.

**Kildehierarkiet** bør være like tydelig:

**A – primær og autoritativ:** lov, offisielt voteringsresultat, budsjettvedtak, regnskap, registerdata.
**B – primær institusjonell analyse:** SSB, FFI, NSM, PST, NVE, Riksrevisjonen.
**C – aktørens egen publisering:** partiprogram, pressemelding, plattformens eget annonsebibliotek.
**D – verifisert sekundærkilde:** etablert presse/forskning som peker til dokumenter.
**E – OSINT-lead:** konto, anonym nettside, påstand eller nettverksindikasjon som ennå ikke er verifisert.

Det offentlige grensesnittet bør alltid vise **både kildeklasse og konklusjonens sikkerhet**.

Det finnes også viktige juridiske og etiske grenser. Politisk oppfatning er en særlig kategori personopplysninger etter personvernregelverket, og et prosjekt som profilerer identifiserbare privatpersoner må derfor ha et dokumentert behandlingsgrunnlag og særskilt vurdering av slike data. Personopplysningsloven/GDPR har også informasjonsregler når opplysninger samles fra andre kilder enn personen selv. Det bør gjennomføres en formell personvern-/DPIA-vurdering før dashboardet publiserer personnettverk eller politiske profiler.

I praksis betyr dette:

- Databasen bør primært undersøke **partier, virksomheter, profesjonelle påvirkningsaktører og offentlige roller**, ikke bygge politiske profiler av vanlige borgere.
- Sosiale mediekontoer som ikke kan knyttes forsvarlig til en offentlig aktør bør pseudonymiseres internt og normalt ikke navngis offentlig.
- Det skal ikke brukes skjulte kontoer, hacking, sosial manipulering eller innhenting av private data.
- En aktør som beskyldes for alvorlig koordinering, skjult finansiering eller fremmed styring bør få påstanden og dokumentasjonen forelagt før publisering.
- Dashboardet trenger synlig **rettelseslogg**, versjonshistorikk og metodeendringer.
- Partiene må vurderes med identiske bevisregler.
- Valgomatens brukerdata bør ikke brukes til annonsemålretting eller bygging av personlige politiske profiler.

Dette er også grunnen til at fulltekstdokumenter kan oppbevares i et internt forskningsarkiv med hash og kildeproveniens, mens det offentlige grensesnittet primært viser korte utdrag, strukturerte fakta og lenke til originalkilden.

## Tidslinje, ressurser og startarkiv

> **Fase 1-avvik**: dette avsnittet beskriver en 12-ukers, 8-9 FTE
> institusjonell prosess. `../README.md` og `roadmap.md` forklarer
> hvorfor fase 1 er bygget solo på en ZimaBoard i stedet, og hva som
> derfor bevisst ikke finnes ennå.

Med «ingen spesifikk ressursbegrensning» ville jeg organisert et **12-ukers MVP-løp med omtrent 8–9 fulltidsressurser**, fremfor én stor sekvensiell undersøkelse. Dette er et prosjektanslag, ikke en ekstern kostnadsfasit.

Et godt kjerneteam er omtrent: to dataingeniører, tre politiske/gravende researchere, én backend/datautvikler, én frontend/datavisualiserer, én redaksjonell QA/metoderessurs og en jurist/personvernressurs på omtrent halv tid. Totalt er størrelsesordenen **70–90 personuker** frem til et offentlig reviderbart MVP dersom mye arbeid kjøres parallelt.

Den første uken bør produsere fire konkrete ting: **database-skjemaet, de 378 tomme vurderingscellene, et råkildearkiv og en fungerende Stortinget-importør**. Deretter kan forskerne arbeide fra samme system fremfor å samle tilfeldige lenker i separate dokumenter.

For Stortinget bør importøren ikke stoppe ved partinivå. API-et støtter voteringsresultat per representant, slik at avvik fra partigruppen, fravær og historiske endringer kan registreres uten å anta partidisiplin.

For resultatmåling opprettes en **indikator før politikken bedømmes**. For eksempel skal energiutsagnene kobles til kraftbalanse, effekt, nett, pris, import/eksport og investerings-/byggetid. NVEs 2025-analyse er en naturlig baseline og peker på en strammere norsk kraftbalanse frem mot 2030 før forventet bedring senere; dette gjør det mulig å bedømme partienes energiplaner mot samme eksterne scenario fremfor mot partienes egne premisser.

For robusthet og digital suverenitet opprettes tilsvarende indikatorer rundt leverandørkonsentrasjon, alternative leverandører, gjenopprettingstid, nasjonal kompetanse og kontroll over data/infrastruktur. NSMs Risiko 2026 gir et direkte grunnlag for slike indikatorer.

**Startarkiv / manifest v0.1.** Følgende dokumenter og datasett er allerede identifisert og verifisert som første innhøstingssett i denne researchrunden. Ved implementering bør hvert dokument lagres som originalfil/HTML + metadata + hentetid + SHA-256, slik at en senere nettstedendring ikke endrer det historiske bevisgrunnlaget.

| Arkiv-ID | Dokument/datasett | Status i første pass |
|---|---|---|
| `STO-API-001` | Stortingets åpne datatjeneste og API-dokumentasjon | verifisert |
| `STO-LOB-001` | Dok. 8:53 S (2025–2026), lobbyregister | identifisert |
| `STO-LOB-002` | Innst. 135 S (2025–2026) | hentet/verifisert |
| `STO-LOB-003` | voteringsoversikt 21.04.2026 | hentet/verifisert |
| `STO-BUD-001` | alternative statsbudsjetter 2026 | hentet/verifisert |
| `PARTI-2025` | Partifinansiering.no, valgkampbidrag/regnskap | verifisert |
| `PARTILOV-001` | Prop. 144 L (2024–2025) | hentet/verifisert |
| `PARTILOV-002` | lovendring 20.06.2025 / ikrafttredelse 01.07.2025 | verifisert |
| `PARTILOV-003` | Partilovutvalgets mandat | hentet/verifisert |
| `FFI-26003` | *Økende, komplekst og uoversiktlig* | hentet/verifisert |
| `PST-NTV-2026` | Nasjonal trusselvurdering 2026 | hentet og visuelt kontrollert |
| `NSM-RISK-2026` | Risiko 2026 | hentet/verifisert |
| `BRREG-API-001` | Brønnøysund datasett/API-katalog | verifisert |
| `BRREG-RRH-001` | Register over reelle rettighetshavere – API/tilgang | verifisert |
| `SSB-LOTTE-001` | Skatteberegningsmodellen LOTTE | verifisert |
| `NVE-LKA-2025` | Langsiktig kraftmarkedsanalyse 2025 | verifisert |
| `RREV-IDX-001` | Riksrevisjonens rapportarkiv | verifisert |
| `EU-POLADS-001` | norsk EØS-notat om forordning 2024/900 | hentet/verifisert |
| `META-ADLIB-001` | Meta Ad Library/Transparency Center | verifisert som kilde |
| `TIKTOK-CCL-001` | TikTok Commercial Content Library | verifisert som kilde |

FFI-rapporten får spesielt høy prioritet i arkivet fordi den gir både positive og negative funn: over 22 millioner innlegg og kommentarer fra Facebook, X, TikTok og Telegram ble analysert for perioden fra september 2024 til valget 2025, men forskerne kunne — bortsett fra #OpNorway — ikke dokumentere utenlandsk påvirkning direkte rettet mot selve valget. Det er akkurat denne evnen til å lagre **«ingen evidens funnet»** som et legitimt resultat som vil beskytte prosjektet mot bekreftelsesbias.

Partilovutvalgets arbeid bør også behandles som en levende datakilde frem mot fristen **1. juni 2027**, fordi mandatet eksplisitt inkluderer finansiering, offentliggjøring, mulig valgkampregulering, bidragstak og kampanjer fra tredjeparter/anonyme aktører.

Den viktigste tekniske leveransen etter startmanifestet er derfor ikke selve spørreskjemaet. Det er en **reproduserbar politisk hendelseslogg** der enhver bruker kan åpne eksempelvis utsagn 1 om lobbytransparens og se:

**Venstres program → Venstres Dok. 8-forslag → presidentskapets endringer → Rødts og MDGs alternative forslag → alle individuelle stemmer → endelig vedtak → eventuell senere utredning/lov → faktisk implementert register → hvem som omfattes og unntas.**

Samme struktur brukes så på kjernekraft, EØS, utenlandsk eierskap, sykehus, barns lavinntekt, KI/regnekraft eller beredskapslagre.

Det er dette som gjør prosjektet til en **ekte valgomat**: ikke hvem som formulerer verdiene penest, men hvor godt **verdier, penger, makt, handling og dokumenterte konsekvenser faktisk henger sammen**.
