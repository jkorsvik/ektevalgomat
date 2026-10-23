# Konkurrentanalyse: norske valgomater (2026-09-24) — arkivert rådata + kritisk gjennomgang

**Status: research-notes, ikke app-innhold.** Ingenting her leses av
serveren. Dette er browser-agent-evalueringer av tre konkurrerende
valgomater (NRK, VG, Aftenposten), arkivert fordi de gir konkret,
etterprøvbar evidens for hvorfor ektevalgomat er bygget som det er —
ikke fordi de skal siteres ukritisk. Se "Kritisk gjennomgang" nedenfor
før noe av dette gjenbrukes i noe brukervendt (metode-siden,
markedsføring, README).

Evalueringene ble kjørt av en separat browser-styrende AI-agent (se
prompten i sesjonsloggen 2026-09-24), etter en rubrikk med fem
kriterier: Funksjonalitet, Tillit til metoden, Klarspråk, Ærlighet om
usikkerhet, Tilgjengelighet/robusthet. Rubrikken er identisk for alle
tre, men kriterie-B ("Tillit til metoden") er eksplisitt tilpasset
siden ingen av de tre er kildesporende slik ektevalgomat er.

---

## Kritisk gjennomgang og kontekst

### Hva jeg selv har verifisert (ikke bare tatt agentens ord for)

Jeg stikkprøvde de mest dramatiske og mest sjekkbare påstandene direkte
i nettleser (playwright) 2026-09-24, samme dag rapportene kom inn:

| Påstand | Kilde | Verifisert? | Metode |
|---|---|---|---|
| VG: `document.documentElement.lang === "en"` | VG-rapport F8 | ✅ Bekreftet eksakt | `eval` mot live side |
| VG: "Null JS-feil, null 404-feil" | VG-rapport, seksjon A | ❌ **Motsagt** | Fant 2 reelle konsollfeil (400 fra `session-service.payment.schibsted.no/user-context`) ved sideload uten innlogget Schibsted-konto |
| Aftenposten: "Stortingsvalget 2021" synlig i toppen | AP-rapport F1 | ✅ Bekreftet visuelt | Skjermbilde av live side |
| Aftenposten: Erna Solberg-bilde fremtredende i hero | AP-rapport F2 | ✅ Bekreftet visuelt | Skjermbilde — gråhåret kvinne med perlekjede og NTB-mikrofon, umiskjennelig Solberg, i samme billedcollage som dagens partiledere |
| Aftenposten: JSON-LD `keywords: "stortingsvalget 2021,afghansk politikk"` | AP-rapport F3 | ✅ Bekreftet ord-for-ord | Dumpet rå JSON-LD-script fra DOM |
| Aftenposten: `datePublished: 2021-09-08`, `dateModified: 2025-08-12` | AP-rapport F9 | ✅ Bekreftet eksakt | Samme JSON-LD-dump |
| Aftenposten: "årets valgomat" uten årstall i Om-teksten | AP-rapport F4 | ✅ Bekreftet ord-for-ord | Lest `document.body.innerText` |
| Aftenposten: personvern "ingenting" forklart | AP-rapport F5 | ⚠️ **Nyansert, ikke helt riktig** | Footeren har faktisk generiske "Personvernpolicy" og "Informasjonskapsler"-lenker — akkurat som Aftenpostens vanlige artikkelfooter. Det som faktisk mangler er en valgomat-*spesifikk* forklaring (typen VG har på forsiden: "vi lagrer ikke noe som kan koble svar... til deg"), ikke fravær av personvernlenke helt og holdent |
| TV2/Elevkanalen: `lang="en"` på norsk side | TV2-rapport F1 | ✅ Bekreftet eksakt | `<html lang="en">` i rå HTML fra `valgomat.elevkanalen.no` |
| TV2/Elevkanalen: "Om valgomaten"-tekst, "kun hoppe over fire spørsmål" | TV2-rapport, seksjon C | ✅ Bekreftet ord-for-ord | Samme HTML-fetch, `info-card__description` |
| TV2/Elevkanalen: metodeartikkelen er "bak TV 2-innlogging" | TV2-rapport F2 | ⚠️ **Nyansert, ikke helt riktig** | Artikkelen er faktisk bak et *gratis* TV 2-konto-krav ("Om du ikke har en TV 2-konto fra før kan du lage en, helt uten kostnad!"), ikke et betalt abonnement. "Paywall" i vanlig betydning (penger) er feil begrep — riktigere er "kontokrav" eller "innloggingsmur". Terskelen for brukeren er likevel reell: kontoopprettelse er en ekstra friksjon en anonym valgomat-bruker ikke forventer å måtte gjøre for å lese om metoden |

**Ny oppdagelse jeg gjorde selv, som ingen av rapportene fanget helt:**
Både VG- og Aftenposten-valgomaten kaster samme feil —
`session-service.payment.schibsted.no/user-context` svarer 400 for en
ikke-innlogget besøkende. Dette er sannsynligvis ikke et
valgomat-spesifikt bug, men en delt Schibsted-konsern-feil i
SPiD-identitetsplattformen som begge produktene arver uendret. Verdt å
vite: dette er IKKE noe NRK- eller VG-/Aftenposten-redaksjonen for
valgomaten kontrollerer eller nødvendigvis er klar over — det sitter et
lag opp i konsern-infrastrukturen. Riktig sted å nevne dette er "delt
Schibsted-plattform", ikke "VGs valgomat har en bug".

### Hva som IKKE er verifisert av meg (stol på det med forbehold)

Alt som krever full quiz-gjennomkjøring, cookie-navn i DevTools, eller
tallsammenligning over flere sesjoner er tatt fra agentens rapport uten
uavhengig etterprøving:

- NRK: 70% totalscore vs. 66.9% dimensjons-snitt (F1) — det mest
  interessante enkeltfunnet i hele analysen, men jeg har ikke kjørt
  quizen selv til 100% for å bekrefte tallene.
- NRK: `_sp_id.eecb`-cookien og "4 sporingsscripts" (F6).
- VG: "8 sporingsscripts", TCF-implementasjon, 5-minutters cache
  (bekreftet av rapporten selv via "timestamp-manipuleringstest" — god
  metode, men jeg har ikke re-kjørt den).
- Alle "poengtabell"-vurderinger (1-5) er agentens subjektive skjønn,
  ikke noe jeg har kalibrert mot en ekstern standard.
- TV2/Elevkanalen: "null sporingsscripts, null cookies, null
  nettverkskall" (rapportens sterkeste påstand) — jeg har kun bekreftet
  at `<head>` i den statiske HTML-en ikke inneholder synlige
  analytics-tags (ingen `gtag`, ingen Snowplow-loader). Jeg har IKKE
  overvåket faktisk nettverkstrafikk under en hel quiz-sesjon (ville
  kreve DevTools/Playwright, ikke en enkel HTML-fetch), så
  "null nettverkskall gjennom hele sesjonen" er fortsatt kun agentens
  påstand, ikke min bekreftelse.

### Ny egen oppdagelse: TV2s poengmodell er vektet, ikke binær

Ved å dumpe rå spørsmålsdata fra siden fant jeg at TV2/Elevkanalens
scoring-modell er mer sofistikert enn både rapporten og "Om
valgomaten"-teksten antyder. "Om valgomaten" sier bare at partier som
"svarer på samme side som deg" får poeng — det leses som binært (enig
= 1 poeng, uenig = 0). Rådata viser derimot at hvert svaralternativ har
en egen `weightings`-liste per parti, med vekt **1 eller 1.5** (ikke
bare 0/1). Et parti kan altså få mer poeng for et svar de mener sterkt
enn for et svar de bare svakt støtter — en gradert enighet, ikke en
ja/nei-match. Dette er teknisk en styrke (mer presis enn NRK/VGs enklere
modeller), men det er ikke forklart noe sted for brukeren, og det gjør
rapportens egen påstand "prosentberegningen er verifiserbar" upresis —
den er verifiserbar *gitt* vektene, men vektene i seg selv er ikke
dokumentert eller innsyn-tilgjengelige for brukeren, kun for noen som
leser rå JS-payload.

**Regelen for videre bruk:** alt i tabellen over ("Hva jeg selv har
verifisert") kan siteres som fakta. Alt i "ikke verifisert"-lista må
re-sjekkes rett før publisering — nettsider endrer seg, cookie-banner
og adaptiv spørsmålslogikk kan gi andre tall på en annen dag, og en
"agent sa det" er et kvitteringsvitnesbyrd, ikke selve artefaktet.

### Mønsteret som går igjen — hvorfor dette faktisk betyr noe for ektevalgomat

Alle tre analysene, uavhengig av hverandre, lander på de samme
strukturelle svakhetene hos de dominerende valgomatene:

1. **Ingen kildelenke per spørsmål, i noen av de tre.** Brukeren må ta
   redaksjonens ord for partiposisjonen. Dette er ordrett den
   metodiske hullet ektevalgomats designnotat (`docs/research-memo.md`)
   ble skrevet for å tette — se metode.html.tmpl: "sammenlignes mot en
   kodet, kildesporet posisjon". Tre uavhengige stikkprøver av markedet
   bekrefter at hullet er reelt, ikke en påstått fordel uten motstykke.
2. **"Jonas Gahr Støre bør være/fortsette som statsminister"** finnes
   hos NRK (adaptivt, sent i løpet) OG VG (som **spørsmål 1**). Et
   personvalg maskert som politisk sak, strukturelt umulig for AP å
   "tape". Ektevalgomat kan ikke ha dette problemet fordi hvert utsagn
   må ha en kildesporet handling (forslag/stemme/lov) — det filtrerer
   automatisk bort spørsmål om en navngitt persons egnethet.
3. **NRKs 70%/66.9%-avvik** (hvis bekreftet ved re-kjøring) er akkurat
   eksempelet på hvorfor ektevalgomats "Tre separate mål — aldri én
   score"-prinsipp (metode.html.tmpl) finnes: en kollapset totalprosent
   som ikke stemmer med sine egne undertall, uten forklaring til
   brukeren, er presis den brukerfellen prinsippet er designet mot.
4. **Ingen fungerende noscript-fallback hos NRK eller VG.**
   Aftenpostens noscript-elementer er CSS/bilder, ikke meningsfull
   tekst — samme reelle mangel. Ektevalgomat har en ekte fallback (hele
   beviskjeden lesbar via `/utsagn/N` uten JS).
5. **Aftenposten-funnet er en annen kategori feil helt** — ikke en
   designsvakhet, men driftsforfall: et 2021-produkt gjenbrukt uendret
   visuelt for 2025-valget. Mindre relevant som "hvorfor ektevalgomat
   er bedre designet", mer relevant som "hvorfor et versjonert,
   git-committed innhold (som ektevalgomats YAML-baserte evidensdata)
   ikke kan gå like stille ut av synk" — stille innholdsforfall er
   nøyaktig risikoen versjonert, gjennomsiktig evidens er ment å unngå.
6. **`lang="en"` på en norsk side er en gjenganger, ikke et enkelttilfelle.**
   VG og nå TV2/Elevkanalen har begge denne feilen — to av fire
   uavhengig bygde produkter. Sannsynlig forklaring: et felles
   rammeverksoppsett (SvelteKit-standardmal har `lang="en"` i
   utgangspunktet) som ingen av teamene overstyrer. Verdt å eksplisitt
   sjekke i ektevalgomats egne `.html.tmpl`-filer at `<html lang="nb">`
   faktisk er satt, ikke bare anta at det er riktig fordi innholdet er
   norsk.
7. **TV2/Elevkanalen er eneste produkt med reelt null tracking** (så
   langt jeg har kunnet bekrefte fra statisk HTML) — dette er
   sannsynligvis en direkte konsekvens av at det er et skoleprodukt
   under NRK-lignende personvernkrav for barn/unge, ikke en generell
   bransjenorm. Ektevalgomat bør ikke ta dette som bevis på at "ingen
   sporing" er standarden å slå — det er unntaket, ikke normen, blant
   de fire evaluert her.
8. **Innloggingskrav for å lese metode er en ny variant av kildeproblemet.**
   TV2/Elevkanalen har (i motsetning til de tre andre) en reell
   metodeartikkel — men den krever kontoopprettelse. Funksjonelt
   samme brukeropplevelse som "ingen kildelenke": brukeren kan ikke,
   i praksis, verifisere noe uten en ekstra handling de ikke forventer
   å måtte gjøre. Ektevalgomats prinsipp om at hele beviskjeden er
   lesbar uten innlogging (og uten JS, se punkt 4) dekker denne
   varianten av problemet også, ikke bare den rene "ingen kilde"-varianten.

### Forsiktighetsregel før noe av dette blir offentlig

Dette er kritikk av navngitte, virkelige mediehus. Før noe av
konklusjonene ovenfor limes inn i en offentlig side (metode.html.tmpl,
README, en artikkel):

- Hver kvantitativ påstand (cookie-navn, feilantall, prosentavvik) må
  re-verifiseres samme dag den publiseres, ikke gjenbrukes fra denne
  filens dato.
- Skriv om funn som **egne, navngitte kilder** ("testet 24. september
  2026, se skjermbilde"), ikke som ferdig konklusjon uten dato — sider
  endrer seg, spesielt rundt et valg.
- Aftenposten-funnet (2021-innhold) er det mest alvorlige og mest
  "nyhetsverdige" — men også det som mest sannsynlig blir fikset raskt
  hvis noen fra Aftenposten ser det. Anta at det kan være borte neste
  gang siden sjekkes.

---

## Arkivert rådata: de tre fullstendige rapportene

### Rapport 1: NRK Valgomat 2025

**URL:** `nrk.no/valg/2025/valgomat/storting/nb` · **Testet:** 24. september 2026

#### 1. Executive Summary

NRKs valgomat er et gjennomarbeidet, teknisk velfungerende produkt med
et visuelt imponerende design, null JS-feil og ingen 404-feil under
hele testforløpet. Det er likevel **ikke en kilde-sporende quiz** — den
er en klassisk politisk avstandsmåler der partienes posisjoner er satt
av NRK/partiene i samarbeid, uten per-spørsmål-kildelenker i selve
grensesnittet. Den adaptive algoritmen (dynamisk spørsmålssett mot
slutten) er metodisk interessant, men er dårlig kommunisert til
brukeren og skaper uforklarlige prosentforskjeller mellom kjøringer.
Ingen noscript-fallback finnes, og fire sporingsscripts kjører uten å
nevnes eksplisitt på forsiden.

#### 2. Poengtabell (kriteriene tilpasset NRKs faktiske produkt, ikke ektevalgomat.no-spesifikke krav)

**A. Funksjonalitet — 4/5**
- ✅ Quiz, navigasjon, partiposisjoner, resultater og tema-nedtrekk fungerer uten feil
- ✅ Ingen JS-feil, ingen 404-feil gjennom hele sesjonen
- ✅ Tilbake-knapp finnes fra spørsmål 2 og utover
- ⚠️ Progress-bar oppdateres ikke ved første svar — viser 0% til du trykker "Neste påstand"
- ⚠️ "Ta Valgomaten på nytt"-knappen i footer sender brukeren til den personlige valomaten, ikke tilbake til startskjermen for standardversjonen — navigasjonsproblem

**B. Tillit til metoden — 2/5**
- ✅ Metode-artikkelen ("Slik har vi laga valomatane") er grundig og beskriver algoritmen, kildearbeid og endringslogg
- ✅ Endringsloggen er detaljert og konkret — spesifikke partiposisjoner som er rettet navngis
- ❌ Ingen kildelenker per spørsmål i selve quiz-grensesnittet — brukeren kan ikke verifisere noe mens de svarer
- ❌ "Argumenter for og imot"-seksjonen mangler alle kildereferanser — ingen lenker til NOU-er, stortingsdebatter eller forskning
- ❌ "Jonas Gahr Støre bør være statsminister de neste fire årene" er et spørsmål om en persons egnethet, ikke en politisk sak — metodisk problematisk og umulig å forankre i et partiprogram
- ❌ Partienes posisjoner er "gjort av NRK i samarbeid med partiene" — dvs. partiene selv er med på å sette sin egen posisjon, noe som ikke er transparent overfor brukeren

**C. Klarspråk — 4/5**
- ✅ Spørsmålstekstene er korte, presise påstander
- ✅ "Argumenter for og imot"-tekstene er balanserte og lettleste
- ✅ Resultat-siden med "Hvorfor fikk jeg X?", tema-nedtrekk og "Hva skiller dere mest?" er pedagogisk god
- ⚠️ Forsiden sier "rundt 30 spørsmål" — jeg fikk 27 i begge kjøringer. Det stemmer teknisk sett, men sett opp mot "rundt 30" kan brukere føle seg lurt
- ⚠️ Metode-lenken på forsiden er gjemt i brødteksten som "her" — ikke et synlig navigasjonselement

**D. Ærlighet om usikkerhet — 2/5**
- ✅ "Husk: Valgomaten er ikke en fasit" — caveat finnes
- ❌ Ingen forklaring på at spørsmålssettet endres dynamisk — to brukere med identiske svar vil få ulike spørsmål (og dermed potensielt ulike resultater), uten at dette er kommunisert
- ❌ Totalscoren (70%) stemmer ikke med gjennomsnittet av dimensjonene (66.9%) — algoritmen beregner på individspørsmålsnivå mens dimensjonsvisningen gir gjennomsnittet per tema, og dette er aldri forklart i resultatsiden
- ❌ Ingen markering av at partiposisjoner er usikre (se endringsloggen: FRP, Rødt, Ap, Venstre med flere ble rettet etter lansering — men dette er usynlig i selve quizen)
- ❌ Spørsmålet om statsministerkandidaten presenteres på lik linje med politiske saker uten kontekstualisering

**E. Tilgjengelighet/robusthet — 3/5**
- ✅ Null JS-feil, null 404-feil
- ✅ Viewport meta er riktig satt, CSS er responsivt (flexbox)
- ✅ ARIA-labels finnes på progressbar og knapper
- ❌ Ingen `<noscript>`-fallback — siden er en blank side uten JavaScript. En statsfinansiert allmennkringkaster bør ha en meningsfull fallback
- ⚠️ 4 sporingsscripts kjører (Snowplow x2, Unispring, Medietall) uten å nevnes i forsidesteksten om anonymitet — "Ditt resultat er helt anonymt" er teknisk sant (ingen svar sendes til server), men Snowplow setter persistente ID-cookies (`_sp_id.eecb`) som kan knytte sesjoner til en person over tid
- ⚠️ Viewport kan ikke verifiseres på 375px i dette testmiljøet, men CSS-strukturen er flexbox og responsiv etter kode-inspeksjon

#### 3. Faktiske funn

| # | Funn | Alvorlighet | Reproduksjon |
|---|------|-------------|--------------|
| F1 | Totalscoren stemmer ikke med dimensjonsgjennomsnittet — AP fikk 70% totalt, men snittet av alle 13 dimensjoner = 66.9% | Høy | Kjør quiz → Se resultat → Vis mer enige om → Summer alle dimensjoner |
| F2 | Spørsmålet "Jonas Gahr Støre bør være statsminister" er en personvurdering, ikke en politisk sak. Alle andre partier enn AP er per definisjon "uenige" — og det forstyrrer scoring for disse | Høy | Kjør quiz til 100% — spørsmålet dukker opp som siste adaptive spørsmål |
| F3 | "Ta Valgomaten på nytt" sender til personlig valgomat, ikke standardversjonen | Middels | Gå til resultatsiden → scroll ned → klikk "Ta Valgomaten på nytt med dine temaer" |
| F4 | Ingen kildelenker i grensesnittet — "Argumenter for og imot" har ingen referanser. Brukeren kan ikke etterprøve noe | Høy | Se seksjonen under hvert spørsmål |
| F5 | Ingen noscript-fallback — siden er blank uten JS | Middels | Deaktiver JS → last `nrk.no/valg/2025/valgomat/storting/nb` |
| F6 | Snowplow-cookies settes (`_sp_id.eecb`) uten å nevnes eksplisitt i forsidesteksten om anonymitet | Middels | Åpne DevTools → Application → Cookies etter første sidebesøk |
| F7 | Progress-bar oppdateres ikke ved første svar — viser 0% frem til "Neste påstand" klikkes | Lav | Svar på Q1 uten å trykke neste — progress-badge sier fortsatt "0%" |
| F8 | Metode-lenken er visuelt gjemt som "her" i brødteksten — ingen synlig "Metode"-lenke i header eller nav | Lav | Se forsiden — "Les mer om hvordan Valgomaten fungerer her" |
| F9 | Artikkelen er mer enn ett år gammel — advarselen på metodeartikkelen er misvisende for en valgomatforklaring fra 2025 | Lav | Gå til metodeartikkelen — advarsel vises øverst |

#### 4. Topp 3 forbedringer, rangert etter tillitseffekt

**🥇 1. Legg til kildelenke per spørsmål i "Argumenter for og imot".** Eliminerer F4 og sterkt reduserer F2. Selv en enkel lenke til relevant partiprogram-paragraf eller stortingsvotering per spørsmål ville transformere tilliten til produktet.

**🥈 2. Forklar score-beregningen og den adaptive algoritmen på resultatsiden.** Eliminerer F1 og adresserer det grunnleggende problemet med at dimensjonsprosenter ikke summerer til totalscoren.

**🥉 3. Fjern eller redefiner "Jonas Gahr Støre bør være statsminister".** Adresserer F2 direkte — enten fjerne det, eller reformulere til en politisk sak (f.eks. "Arbeiderpartiet bør lede neste regjering").

---

### Rapport 2: VG Valgomat 2025

**URL:** `vg.no/spesial/2025/valgomat/` · **Testet:** 24. september 2026

#### 1. Executive Summary

VGs valgomat er den mest transparente av de to store norske valgomatene
testet her. Forsiden forklarer algoritmen, scoringsmodellen og
personvernpraksis fullt ut FØR brukeren starter — noe NRK ikke gjør.
Videokonseptet med programledere og undertekst er engasjerende og
unikt, og 5-minutters cache-løftet er teknisk korrekt implementert.
Svakhetene er de samme som for alle norske valgomater: ingen
kildelenker per spørsmål, ingen noscript-fallback, og spørsmålet om
"Jonas Gahr Støre bør fortsette som statsminister" er et
personlighetsvalg, ikke en politisk sak — og det er faktisk spørsmål 1,
som setter tonen for hele opplevelsen. Det mangler også helt en
temabasert dimensjonsanalyse i resultatet.

#### 2. Poengtabell

**A. Funksjonalitet — 4/5**
- ✅ Quiz, video, partiposisjoner, resultat og sammenligning fungerer uten feil
- ✅ Null JS-feil, null 404-feil gjennom hele sesjonen *(se "Kritisk gjennomgang" over — motsagt av egen stikkprøve: 2 reelle 400-feil fra Schibsteds session-service)*
- ✅ 5-punkts skala med "Hverken eller" — brukeren slipper å ta stilling
- ✅ "Du kan se resultatet nå, eller velge flere spørsmål" — eksplisitt brukerkontroll på dybden
- ✅ "Fullfør"-knapp tilgjengelig etter 10 svar, uten å tvinge deg gjennom alt
- ⚠️ Ingen numerisk progress-indikator — bare en tynn hvit linje. Vet ikke "spørsmål 7 av 20"
- ⚠️ "Jonas Gahr Støre bør fortsette som statsminister" er spørsmål 1 — metodisk problematisk start

**B. Tillit til metoden — 3/5**
- ✅ Algoritmen er fullt forklart på forsiden (25/50/75/100%, dobbel vekt ved ytterpunkter) — solid forsprang på NRK
- ✅ "Logg over rettelser VG har gjort" i footer — lenker til redaksjonell åpenhetsside
- ✅ "VG har formulert påstandene, partiene har selv svart" — tydelig ansvarsfordeling
- ❌ Ingen kildelenker per spørsmål — som NRK finnes det ingen mulighet til å etterprøve partiposisjonene
- ❌ Spørsmål 1 er "Jonas Gahr Støre bør fortsette som statsminister" — en AP er per definisjon enig, alle andre uenige. Forsiden sier ingenting om at ett spørsmål er om en navngitt person
- ❌ Ingen separate rettelser-logg for valgomaten spesifikt — den generelle rettelsesloggen dekker alt VG gjør
- ⚠️ "Partiene har selv svart" — selvrapportering fra partiene, ingen uavhengig verifisering nevnt

**C. Klarspråk — 5/5**
- ✅ "Om valgomaten" forklarer alt brukeren trenger å vite i korte setninger
- ✅ Scoringsmodellen forklares med konkrete eksempler (25/50/75/100%)
- ✅ "Vi registrerer klikk og resultater anonymt... men vi lagrer ikke noe som kan koble svar eller vinnerparti til deg" — presist og ærlig
- ✅ "VG-journalistene gjengir ikke sine egne meninger" — eksplisitt ansvarsfraskrivelse for videoinnholdet
- ✅ Ingen motstridende tall eller vage formuleringer funnet

**D. Ærlighet om usikkerhet — 3/5**
- ✅ Algoritme-forklaring er mer transparent enn NRK
- ✅ 5-minutters-cache forklart og korrekt implementert (bekreftet ved timestamp-manipulasjonstest)
- ✅ "Du kan se resultatet nå, eller velge flere spørsmål" — ærlig om at 15 svar er minimumsdekket
- ❌ Ingen forklaring av at kun 9 (stortings)partier er med ut fra en redaksjonell beslutning basert på meningsmålinger — forklaring finnes i Om-teksten, men krever "Vis mer"-klikk
- ❌ Ingen markering av usikre partiposisjoner
- ⚠️ 8 sporingsscripts kjører (TCF, Pulse Schibsted x3, Medietall, GDPR-bundle, Sentry, special-tracking) — samtykkeforvaltning via TCF/CMP er implementert, men ingen eksplisitt samtykkebanner ble vist under testing — GDPR-samtykket er sannsynligvis arvet fra VG.no-session

**E. Tilgjengelighet/robusthet — 3/5**
- ✅ Null JS-feil, null 404-feil *(se motsigelse nevnt over)*
- ✅ Sentry feilmonitorering er implementert
- ✅ GDPR TCF API implementert — bedre enn NRKs rene cookie-dropping uten synlig consent
- ✅ Undertekster på video kan slås av/på, lyd kan dempes
- ❌ Ingen `<noscript>`-fallback — blank side uten JS, identisk problem som NRK
- ⚠️ Video-heavy format er utilgjengelig for brukere med nedsatt syn/hørsel uten undertekster i standard åpen posisjon
- ⚠️ Viewport er responsivt, men video + quiz-card side-om-side bryter på <550px

#### 3. Faktiske funn

| # | Funn | Alvorlighet | Reproduksjon |
|---|------|-------------|--------------|
| F1 | "Jonas Gahr Støre bør fortsette som statsminister" er spørsmål 1 — personvalg, ikke politisk sak. AP er strukturelt fordelaktig | Høy | Start quizen — første spørsmål |
| F2 | Ingen numerisk progress-indikator | Middels | Svar på Q1–Q5 og se header |
| F3 | Ingen kildelenker per spørsmål | Høy | Se hvert enkelt spørsmål |
| F4 | Ingen `<noscript>`-fallback | Middels | Deaktiver JS → last siden |
| F5 | 8 sporingsscripts inkl. Sentry, Schibsted Pulse, Medietall — ingen nevnes i personvernforklaringen | Middels | DevTools → Network ved sideload |
| F6 | Ingen temabasert dimensjonsanalyse i resultatet | Lav | Se resultatsiden etter Fullfør |
| F7 | "Logg over rettelser" er generell VG-logg, ikke valgomat-spesifikk | Lav | Klikk "Logg over rettelser VG har gjort" i footer |
| F8 | `lang="en"` på `<html>`-elementet — siden er på norsk men erklærer engelsk som språk | Lav | `document.documentElement.lang` = "en" — **verifisert 2026-09-24** |

#### 4. Topp 3 forbedringer, rangert etter tillitseffekt

**🥇 1. Flytt eller fjern "Jonas Gahr Støre bør fortsette som statsminister".** Som spørsmål 1 er det det mest synlige spørsmålet i hele quizen, og det skjever strukturelt mot AP.

**🥈 2. Legg til numerisk progress og spørsmålsteller.** "Du er på spørsmål 8 av ~20" reduserer frafall dramatisk.

**🥉 3. Legg til en valgomat-spesifikk rettelseslogg.** NRKs endringslogg (med datoer og navngitte partier) er et forbilde; VGs generelle redaksjonslogg gir ikke samme innsikt.

#### 5. Sammenligning NRK vs VG (kort, fra opprinnelig rapport)

| Kriterium | NRK | VG |
|-----------|-----|-----|
| Forside-forklaring FØR start | ⚠️ Gjemt som "her"-lenke | ✅ Full forklaring synlig |
| Algoritme forklart | ⚠️ Metodeartikkel separat | ✅ Direkte på forsiden |
| Antall partier | ✅ 18 (alle nasjonale) | ⚠️ 9 (stortingspartier kun) |
| Nøytral svaropsjon | ❌ Ingen (4-punkts skala) | ✅ "Hverken eller" |
| Kildelenker per spørsmål | ❌ Ingen | ❌ Ingen |
| Partiposisjon-rettelseslogg | ✅ Detaljert endringslogg | ⚠️ Generell VG-logg |
| Noscript-fallback | ❌ Blank side | ❌ Blank side |
| GDPR/samtykke | ⚠️ Cookies uten banner | ✅ TCF API implementert |
| "Støre statsminister"-spørsmål | ⚠️ Adaptivt, langt ute | ❌ Spørsmål 1 |
| Dimensjons-/temaanalyse | ✅ Komplett nedtrekk | ❌ Ikke tilgjengelig |
| `lang`-attributt korrekt | ✅ nb | ❌ "en" |

---

### Rapport 3: Aftenposten Valgomat 2021/2025

**URL:** `aftenposten.no/norge/politikk/i/bnE123/usikker-paa-hva-du-skal-stemme-proev-vaar-valgomat`
**Testet:** 24. september 2026

#### 1. Executive Summary

Aftenpostens valgomat er et re-brukt 2021-produkt som stilles opp på
nytt for Stortingsvalget 2025 under samme URL, med partilederfotografier
og visuell identitet fra 2021 — Erna Solberg er fremtredende i
headerbilde, Jonas Gahr Støre er ikke statsminister. Innholdet er
oppdatert (sist endret 12. august 2025 i metadata), men designet og URL
lyver brukeren rett i ansiktet. Selve quiz-mekanikken er til gjengjeld
metodisk den mest transparente av de tre: 28 påstander med tydelig
nummerering, en "Vet ikke"-opsjon, og en unik to-trinns metode der
brukeren vekter hvilke temaer som faktisk betyr noe for dem.
Resultatsiden viser alle partienes posisjoner per spørsmål. Men
fraværet av dato-oppdatering, feil partiledere, JSON-LD-søkeord
"afghansk politikk" og ingen personvern-forklaring gjør at tilliten
undergraves av rot snarere enn metodiske svakheter.

**Alle fire kjernepåstander i dette avsnittet (2021-banner,
Solberg-bilde, JSON-LD-nøkkelord, datoer) er uavhengig bekreftet, se
"Kritisk gjennomgang" over.**

#### 2. Poengtabell

**A. Funksjonalitet — 3/5**
- ✅ Quiz med 28 spørsmål fungerer, nummerering er tydelig ("Påstand 1 av 28")
- ✅ To-trinns metode: svar på spørsmål → velg viktige temaer — metodisk overlegen NRK og VG
- ✅ Resultatsiden viser alle partiposisjoner per spørsmål, filtrerbart etter parti
- ✅ Expanderbar forklaringstekst ("Les mer") under hvert spørsmål
- ❌ Tema-vektingsskjermen har store UX-problemer: pilknappene er `<input type="checkbox">` skjult i `<label>` — ikke synlig som knapper
- ❌ "Gå videre"-knappen scrollet siden til toppen uten visuell tilbakemelding om ny skjerm
- ⚠️ Partiposisjonene vises ikke under selve quiz-forløpet — kun i resultatsiden

**B. Tillit til metoden — 2/5**
- ✅ To-trinns score-metode (posisjon + viktighet) er metodisk sterkere enn NRK og VG
- ✅ "Om valgomaten" forklarer de to faktorene
- ❌ Forside merket "Stortingsvalget 2021" synlig øverst — **bekreftet**
- ❌ Partilederfotografier er fra 2021, Erna Solberg fremtredende — **bekreftet**
- ❌ Ingen kildelenker per spørsmål
- ❌ "Vis mer" i Om valgomaten forklarer ikke score-formelen, kun de to faktorene i stikkordsform

**C. Klarspråk — 3/5**
- ✅ Spørsmålene er klare og balansert formulert
- ✅ Forklaringstekstene under hvert spørsmål er informative og nøytrale
- ✅ Resultatteksten er konkret og forståelig
- ❌ "Årets valgomat" uten årstall — **bekreftet ord-for-ord**, aktivt misvisende i 2026
- ❌ Ingen forklaring av score-enhet: "33 poeng" — hva er maks, hva er ett poeng
- ⚠️ Tema-vekter-instruksjonstekst refererer et pil-symbol som ikke er synlig/klikkbart som interaktivt element

**D. Ærlighet om usikkerhet — 2/5**
- ✅ "Vet ikke"-alternativ er implementert
- ✅ "Husk at et tema kan være veldig viktig for deg, selv om du er uenig i den konkrete påstanden" — fin nyansering
- ❌ `robots: noarchive` er satt, men brukeropplevelsen sier "Stortingsvalget 2021" uansett
- ❌ JSON-LD keywords `"stortingsvalget 2021,afghansk politikk"` — **bekreftet ord-for-ord**, "afghansk politikk" henger ikke sammen med innholdet
- ⚠️ **Nyansert (se "Kritisk gjennomgang"):** footeren har generiske Personvernpolicy/Informasjonskapsler-lenker, men ingen valgomat-spesifikk forklaring av hva som lagres eller sendes

**E. Tilgjengelighet/robusthet — 3/5**
- ✅ `lang="nb"` korrekt satt
- ✅ 3 noscript-elementer finnes — men innholdet er CSS og bilder, ikke meningsfull fallback-tekst
- ✅ Responsiv layout med korrekt viewport-meta
- ✅ Komplett JSON-LD publisher-struktur med `correctionsPolicy`, `factCheckingPolicy`, `ethicsPolicy` — **bekreftet**, sterk redaksjonell åpenhetsinfrastruktur
- ⚠️ Pilknappene for viktig-vekting er `<input type="checkbox">` skjult bak `<label>` uten ARIA-label
- ⚠️ Rapportens CSP-begrensning på egen JS-feilsjekk er reell for det verktøyet, men en annen tilnærming (playwright `console`-kommando) fant 3 konsollfeil — samme Schibsted SPiD `session-service` 400-feil som på VG (se "Ny oppdagelse" over)

#### 3. Faktiske funn

| # | Funn | Alvorlighet | Reproduksjon |
|---|------|-------------|--------------|
| F1 | "Stortingsvalget 2021" vises øverst på siden | Kritisk | Last forsiden — **verifisert 2026-09-24, skjermbilde** |
| F2 | Partilederfotografier fra 2021, Erna Solberg fremtredende | Kritisk | Se headerbildet — **verifisert 2026-09-24, skjermbilde** |
| F3 | JSON-LD keyword: "afghansk politikk" | Høy | `getMetadata()` → keywords-felt — **verifisert 2026-09-24, rå JSON-LD-dump** |
| F4 | "Årets valgomat" uten årstall | Høy | Les Om valgomaten-teksten — **verifisert ord-for-ord** |
| F5 | Ingen personvern/datalagrings-forklaring | Høy | Les Om valgomaten og footer — **se nyansering over: generisk lenke finnes, valgomat-spesifikk forklaring mangler** |
| F6 | Score-enhet "X poeng" er uforklart | Middels | Se resultatsiden |
| F7 | Viktig-pilknappene er `<input type="checkbox">` uten ARIA-label | Middels | Inspect DOM på tema-vektingsskjermen |
| F8 | Noscript-elementer er CSS/bilder, ikke meningsfull fallback | Middels | Deaktiver JS → last siden |
| F9 | Publishdate 2021-09-08, ModifiedDate 2025-08-12 uten synlig datooppdatering | Middels | Se artikkeldato og metadata — **verifisert eksakt, JSON-LD-dump** |
| F10 | `article:published_time` vises som "Publisert: 23. juni 2021" — SEO/forvirringsrisiko | Lav | Se byline-dato — **verifisert ord-for-ord** |

#### 4. Topp 3 forbedringer, rangert etter tillitseffekt

**🥇 1. Oppdater alle visuelle og tekstlige årstallsmarkører til 2025.** Eliminerer F1, F2, F4, delvis F9 — de fire mest tillits-ødeleggende funnene. Uten disse er det sannsynlig at brukere forlater siden fordi de tror de er på feil valgomat.

**🥈 2. Legg til én setning om personvern og datasporing i Om valgomaten**, spesifikt for valgomaten (ikke bare generisk lenke).

**🥉 3. Forklar score-enheten og vektings-algoritmen i resultatsiden.** "33 poeng" sier ingenting uten kontekst (maks mulig, hvordan det øker).

#### 5. Tre-veis sammenligning: NRK vs VG vs Aftenposten (fra opprinnelig rapport)

| Kriterium | NRK 2025 | VG 2025 | Aftenposten 2021/25 |
|-----------|----------|---------|---------------------|
| Er innholdet fra riktig år? | ✅ | ✅ | ❌ Visuelt 2021 |
| Antall partier | ✅ 18 | ⚠️ 9 | ⚠️ 9 (stortingspartier) |
| "Vet ikke"/nøytral | ❌ Nei | ✅ Ja | ✅ Ja |
| Algoritme forklart | ⚠️ Separat artikkel | ✅ Forsiden | ❌ Kun stikkordsvis |
| Partipos. under quiz | ✅ Direkte | ✅ Direkte | ❌ Kun i resultat |
| Tema-vekting | ❌ Nei | ⚠️ Etterpå | ✅ Aktivt valg |
| Score-enhet tydelig | ✅ % | ✅ % | ❌ "Poeng" (uforklart) |
| Kildelenker | ❌ | ❌ | ❌ |
| Personvern forklart | ⚠️ Ja, m/hull | ✅ Tydelig | ❌ Ingenting *(nyansert over)* |
| Noscript-fallback | ❌ Blank | ❌ Blank | ⚠️ Finnes, men CSS/bilder |
| JS-feil | ✅ 0 *(motsagt, se over)* | ✅ 0 *(motsagt, se over)* | ⚠️ CSP hindrer rapportens egen verifisering, men 3 funnet via annen metode |
| Statsminister-personspørsmål | ⚠️ Adaptivt, sent | ❌ Spørsmål 1 | ✅ Ingen |
| Rettelseslogg | ✅ Detaljert | ⚠️ Generell | ⚠️ "Endringslogg"-lenke |

---

---

### Rapport 4: TV 2 Valgomat 2025 ("Ditt Valg 2025" — Mediekompasset/Elevkanalen)

**URL:** `valgomat.elevkanalen.no` (også `dittvalg2025.no`) · **Testet:** 24. september 2026

**Kontekstavklaring gjort av agenten selv før rapporten:** dette er et
læringsopplegg fra Mediekompasset og TV 2 Skole for elever på
mellomtrinn, ungdomstrinn og VGS, ikke en ordinær redaksjonell
valgomat på linje med de tre andre. Det er verdt å ta med når man
sammenligner — målgruppen og kravene til den er forskjellige (bl.a.
strengere personvernkrav for barn/unge, se punkt 7 under "Mønsteret
som går igjen" over).

#### 1. Executive Summary

Dette er TV 2s valgomat for Stortingsvalget 2025, hostet på
Mediekompassets/Elevkanalens domene som læringsressurs for
skoleelever. Teknisk sett er den den **reneste** av alle valgomatene
testet her: null sporingsscripts, null cookies, null nettverkskall
under quiz-sesjonen (se forbehold om egen verifisering over), all
beregning skjer lokalt i nettleseren og ingenting sendes til server.
Designet er moderne, forklaringstekstene er balanserte, og
score-algoritmen er transparent og verifiserbar (men se "Ny egen
oppdagelse" over — den er faktisk vektet, ikke binær, noe som ikke
kommuniseres). Svakhetene sammenlignet med NRK og VG er at det ikke
finnes noen per-spørsmål partiposisjon underveis (bare sammenligning i
resultatet), metodeartikkelen krever gratis TV 2-konto (ikke betalt
abonnement — se nyansering over), og `lang="en"` er feil på en norsk
side (**bekreftet**). For målgruppen — skoleelever — er dette en svært
god ressurs.

#### 2. Poengtabell

**A. Funksjonalitet — 4/5**
- ✅ Quiz fungerer feilfritt — null JS-feil, null 404-feil gjennom hele sesjonen
- ✅ Kategori-velger (velg 3 av 9 temaer) gir brukeren aktiv kontroll over innholdet — **bekreftet, 9 kategorier synlige i rå HTML**
- ✅ Mellomresultat per kategori — "Du er ferdig med første tema! Mest enig med: AP, Minst enig med: FRP" — pedagogisk og engasjerende
- ✅ Progress-bar viser tydelig de 3 valgte kategoriene som ikoner + sjekkmark
- ✅ "Hva betyr dette?" — expanderbar forklaringstekst på hvert spørsmål, synlig uten klikk
- ✅ Score-verifiserbar: `percentage = score / maxScore × 100` stemmer eksakt med vist prosent — **se nyansering: vektet 1/1.5 per parti per svar, ikke binært, ikke dokumentert for bruker**
- ✅ Bakgrunnsfarge endres per kategori — visuell og pedagogisk effektiv
- ❌ Partiposisjonene vises ikke under quiz-forløpet — man ser ikke hva partiene svarte mens man svarer
- ⚠️ Ingen "vet ikke"/nøytral opsjon — 4-punkts skala tvinger bruker til standpunkt *(**korrigert**: rådata viser faktisk 5 svaralternativer inkl. "Usikker" mellom Enig/Uenig — se `options` i JSON-payload. Rapportens egen poengtabell (D) sier "Ingen 'vet ikke'-opsjon" men egen funn-tabell (F4) sier det samme — begge feil, "Usikker" finnes)*

**B. Tillit til metoden — 3/5**
- ✅ Prosentberegningen er verifiserbar — `score/maxScore × 100`, maks bestemmes av antall besvarte spørsmål (ingen konstant max)
- ✅ "Vi har formulert påstander og plassert partiene ut ifra deres politikk" — tydelig redaksjonelt ansvar — **bekreftet ord-for-ord**
- ✅ Navngitt redaksjon i footer (6 navn + ansvarlig redaktør + nyhetsredaktør) — **bekreftet, 6 navn: Ulrikke Torgersen, Christoffer Sandell, Kristine Malde Lende, Øystein Hammersland, Katarina Gjendem Murphy, Ole Martin Kristiansen**
- ✅ Tilsynsorgan Medietilsynet oppgitt + redaktøransvar-segl — **bekreftet**
- ⚠️ **Nyansert:** metodeartikkelen er IKKE bak et betalt abonnement, men bak et gratis TV 2-kontokrav ("helt uten kostnad"). Fortsatt en reell friksjon, men "paywall" er misvisende presist begrep
- ❌ Ingen partiposisjonsdokumentasjon synlig i quiz — brukeren kan ikke se "AP svarte Helt enig" mens spørsmålet stilles
- ⚠️ Maks 4 hopp-over — grense er kommunisert på forsiden ✅ (**bekreftet ord-for-ord**), men ikke i quizen når man nærmer seg grensen

**C. Klarspråk — 5/5**
- ✅ Forsiden forklarer metoden klart og korrekt — **bekreftet ord-for-ord mot rå HTML**
- ✅ "Du kan kun hoppe over fire spørsmål" — eksplisitt grense, **bekreftet**
- ✅ Forklaringstekstene per spørsmål er nøytrale og balanserte (pro/kontra per standpunkt) — **bekreftet på flere spørsmål i rådata, f.eks. kontantstøtte- og hijab-spørsmålene**
- ✅ Mellomresultat-skjermen er pedagogisk: "Husk at alt fortsatt kan forandre seg"
- ✅ Resultatsiden viser "Her har du og [parti] svart likt/ulikt" med konkrete spørsmål — svært god pedagogikk
- ✅ Ingen selvmotsigelser eller forvirrende tall funnet

**D. Ærlighet om usikkerhet — 4/5**
- ✅ Ingenting lagres eller sendes til TV 2 — teknisk plausibelt for selve besvarelsen (spørsmålsdata leveres inline i første sideoppslag, ikke via separate API-kall — **bekreftet i rå payload**), men se forbehold: ikke fullstendig verifisert av meg for hele sesjonen
- ✅ Mellomresultat advarer: "Husk at alt fortsatt kan forandre seg"
- ✅ Partifoto og navn i resultatet — ingen skjulte partier
- ✅ "Start på nytt" alltid tilgjengelig — brukeren kan endre temavalg
- ❌ Ingen personvernerklæring synlig — ingen tekst om databeskyttelse på selve valgomat-siden
- ⚠️ Metodeartikkelen krever gratis konto — brukere som vil verifisere metodikken må opprette konto

**E. Tilgjengelighet/robusthet — 3/5**
- ✅ Null JS-feil, null 404-feil
- ✅ Korrekt viewport meta, ingen horisontal scroll, responsivt
- ✅ Null sporingsscripts synlige i statisk `<head>` — **bekreftet for statisk markup, ikke for runtime-nettverkstrafikk (se forbehold)**
- ⚠️ Null cookies — ikke uavhengig verifisert av meg (krever DevTools-sesjon, ikke gjort)
- ❌ `lang="en"` på `<html>`-element — **bekreftet eksakt**, `<html lang="en">` i rå HTML fra `valgomat.elevkanalen.no`
- ❌ Ingen `<noscript>`-fallback — blank side uten JavaScript (SvelteKit SPA)
- ❌ Metodelenken åpner i samme fane uten advarsel om innloggingskrav — **bekreftet: lenken går direkte til `tv2.no/nyheter/slik-lagde-vi-valgomaten/17769572/`, som viser innloggingsvegg umiddelbart**

#### 3. Faktiske funn

| # | Funn | Alvorlighet | Reproduksjon |
|---|------|-------------|--------------|
| F1 | `lang="en"` på norsk side — påvirker skjermlesere og søkemotorer | Middels | **Verifisert 2026-09-24**: `document.documentElement.lang` = "en" |
| F2 | Metodeartikkelen krever gratis TV 2-konto (ikke betalt) | Middels | **Verifisert**: `tv2.no/nyheter/slik-lagde-vi-valgomaten/17769572/` viser "Logg inn for å lese denne artikkelen" |
| F3 | Ingen `<noscript>`-fallback — blank side uten JavaScript | Middels | Deaktiver JS og last siden |
| F4 | Score-modellen er vektet (1/1.5 per parti per svaralternativ), ikke binær som "Om valgomaten" antyder | Lav-middels (ny, ikke i original rapport) | Dump rå `survey`-JSON fra sidens `<script>`-payload, se `weightings`-felt per `options`-entry |
| F5 | Ingen personvernerklæring på valgomat-siden | Lav | Se footer på valgomat.elevkanalen.no |
| F6 | Partiposisjonene er usynlige under quiz-forløpet — vises kun i resultat | Lav | Svar på spørsmål og se at partienes svar ikke vises |
| F7 | Mellomresultat viser kun #1 og #sisteplass — ikke full rangering per kategori | Info | Se kategoriskjermen etter ferdig tema |

#### 4. Topp 3 forbedringer

**🥇 1. Gjør metodeartikkelen tilgjengelig uten kontoopprettelse.** Eliminerer F2. Selv med gratis konto er terskelen for høy for en anonym besøkende — lim inn kjerneavsnittene direkte i "Om valgomaten"-boksen på selve valgomat-siden.

**🥈 2. Bytt `lang="en"` til `lang="nb"`.** Eliminerer F1, spesielt viktig for et produkt rettet mot elever som kan bruke hjelpeteknologi.

**🥉 3. Dokumenter vektingsmodellen for brukeren.** Adresserer F4 — "partier som svarer likt som deg får poeng" er en forenkling av en faktisk gradert 1/1.5-vekt. Enten forenkle modellen til å matche beskrivelsen, eller oppdater beskrivelsen til å nevne graderingen.

#### 5. Endelig sammenligning av alle fire valgomater

| Kriterium | NRK 2025 | VG 2025 | Aftenposten 2021/25 | TV2/Elevkanalen |
|-----------|----------|---------|---------------------|-----------------|
| Funksjonalitet | 4/5 | 4/5 | 3/5 | 4/5 |
| Tillit til metode | 2/5 | 3/5 | 2/5 | 3/5 |
| Klarspråk | 4/5 | 5/5 | 3/5 | 5/5 |
| Ærlighet om usikkerhet | 2/5 | 3/5 | 2/5 | 4/5 |
| Tilgjengelighet/robusthet | 3/5 | 3/5 | 3/5 | 3/5 |
| **Total (agentens egen sum, ikke uavhengig kalibrert)** | 15/25 | 18/25 | 13/25 | 19/25 |
| Sporingsscripts | 4 (Snowplow etc.) | 8 (Pulse etc.), + 2 reelle SPiD-400-feil (**egen funn**) | Ukjent (CSP-begrenset), + 3 konsollfeil samme SPiD-mønster (**egen funn**) | 0 synlig i statisk HTML (runtime ikke fullt verifisert) |
| "Vet ikke"-opsjon | ❌ | ✅ | ✅ | ✅ ("Usikker" — **rapporten sa feilaktig ❌, korrigert her**) |
| Partipos. underveis | ✅ | ✅ | ❌ | ❌ |
| Metode tilgjengelig | ⚠️ Separat, gratis | ✅ Forsiden | ❌ | ⚠️ Gratis konto kreves (ikke "paywall") |
| Noscript-fallback | ❌ Blank | ❌ Blank | ⚠️ Finnes, men CSS/bilder | ❌ Blank |
| `lang`-attributt | ✅ nb | ❌ en | ✅ nb | ❌ en |
| Kildelenker per spørsmål | ❌ | ❌ | ❌ | ❌ |

**Merk:** "Vinner"-formuleringer fra originalrapporten ("TV2/Elevkanalen vinner på personvern", "VG vinner på transparent metode") er utelatt her — det er agentens subjektive konklusjon uten kalibrert vekting mellom kriteriene, og hverken NRK/VG/AP-poengsummene eller TV2s er uavhengig kalibrert mot samme skala av meg. Tabellen over holder seg til verifiserbare enkeltfakta.

---

## Videre arbeid

Flere valgomater kan legges til under samme mal (samme fem kriterier,
samme "verifisert av meg"-disiplin) etter hvert som de evalueres —
append som "Rapport N" i dette dokumentet, ikke egne filer, så
tverrsnitts-mønsteret ("Mønsteret som går igjen") forblir ett sted å
oppdatere.
