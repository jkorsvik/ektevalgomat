# Kan programpunktene kobles til utsagnene? (måling, 2026-10-13)

## Spørsmålet

Steget **Løfte** i kjeden er ment å vise hva partiet sier i sitt eget program,
for *dette* utsagnet. `ProgramPoint` har et felt for det:

```yaml
statements: [<statement_id>, ...]   # valgfritt: utsagn løftet berører
```

Feltet står tomt på alle 316 programpunkter. Spørsmålet var om alle 316 kan
kodes, slik at løftesteget fylles per utsagn.

## Hva som ble gjort (2026-10-13)

Målingen under ble gjort først, og den endret svaret. Resultatet ble:

- **59 av 316 programpunkter** har fått en kobling, med **64 lenker** til
  sammen (et punkt kan svare på mer enn ett utsagn), hver med en skrevet
  begrunnelse i YAML-filen.
- **32 av 60 utsagn** har minst ett programpunkt som navngir dem.
- **7 av lenkene krysser dimensjoner.** De to taksonomiene overlapper uten å
  være like: datasentrenes styring ligger i «Teknologisk kapasitet» i
  programmene, men utsagn 11 regner det som «Norsk handlefrihet». Hver av de
  sju står oppført med begrunnelse i `program_links_test.go`, så en åttende
  ikke kan snike seg inn.
- **Resten står tomt med vilje.** Det er ikke en manglende jobb.

Stegene på utsagnssiden viser nå forskjellen: **Løftetrinnet teller bare
punkter som navngir utsagnet**, og alt partiet ellers sier i samme tema
ligger bak et utvidelsesfelt merket som tema. Et punkt som bare deler tema
gjør altså ingen påstand om utsagnet.

## Målingen

Materialet:

- **316** programpunkter, fordelt på **9** dimensjoner.
- **60** utsagn, fordelt på **12** dimensjoner.

Tre av utsagnsdimensjonene (`Finansiering av fellesskapet`, `Tilgang,
integrering og rettsstat`, `Tillegg (utenfor de 42)`) har **ingen**
programpunkter i det hele tatt. Der kan koblingen ikke finnes.

Mekanisk måling: hvor mange programpunkter har et **tema** som navngir
utsagnets emne, i samme dimensjon?

**4 av 316.** Og én av de fire er feil ved etterlesing:

```
mdg[0]  Lobbyregister                   ->  1  Lobbyregister
v[0]    Lobbyregister                   ->  1  Lobbyregister
sv[8]   Statlig mineralselskap          -> 17  Statlig mineralselskap for letevirksomhet
sp[22]  Gratis SFO for lavinntektsfamilier -> 34 Gratis barnehageplass for lavinntektsfamilier
```

Den siste er ikke en kobling: SFO og barnehage er to forskjellige ting.
Mekanisk matching alene er altså ikke trygg nok, selv på fire treff.

En generøs manuell lesning av to dimensjoner (`Demokratisk åpenhet`,
`Norsk handlefrihet`) finner omtrent **ett av seks** punkter som med rimelighet
berører et bestemt utsagn — «ett lobbyregister», «valg mot utenlandsk
påvirkning», «nei til EUs fjerde energimarkedspakke». Det gir i størrelsesorden
**50 av 316**. Resten er tematiske, ikke utsagnsspesifikke: «åpenhet om skatt,
selskapsstrukturer og eierforhold» hører til et tema, ikke til ett av de fem
åpenhetsutsagnene.

## Funn

Å kode alle 316 ville bety å finne opp rundt **260 koblinger**. Ingen
mekanisk regel finner dem, og hver kobling er en vurdering som ikke kan
verifiseres maskinelt. På et nettsted hvis ene poeng er at hvert standpunkt
skal kunne etterprøves, er en oppdiktet kobling verre enn en tom.

Feltets egen dokumentasjon sier det samme: det er *valgfritt*, og beskrives
som «et utsagn løftet berører» — ikke som noe hvert punkt har.

## Anbefaling

To ærlige veier, og de utelukker hverandre ikke:

1. **Dimensjonskobling (ingen ny data).** Løftesteget viser partiets punkter i
   utsagnets *dimensjon*, merket som tema: «Partiet sier i dette temaet».
   Dette er hva programmene faktisk inneholder, det dekker alle 60 utsagn, og
   det krever ingen tolkning. Det er den koblingen materialet støtter.
2. **Presis kobling der den finnes (manuell, gradvis).** Kod `statement:` bare
   der punktet navngir samme konkrete spørsmål, med en begrunnelse per kobling,
   og la resten stå tomt. Rundt 50 punkter i første runde, ikke 316.

Det som ikke bør gjøres er å fylle feltet for å fylle det.
