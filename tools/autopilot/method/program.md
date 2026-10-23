# Method: reading a party programme

Used by a worker when a queue item asks for one party's programme. The goal is
a list of what the party says it will do, each line a verbatim quote from the
party's own programme PDF, so a reader can compare promise against vote.

## The nine canonical programmes (2025-2029)

Fetch the current PDF from the party's own site; do not reuse a cached URL
blindly. The known-good URLs as of 2025-10-03 (see the commits that first
fetched them) are in `tools/harvester/topics.yaml`'s neighbourhood and in the
`partiprogram` evidence files already in `internal/data/content/evidence/`.
Rødt's programme is behind a Vercel bot challenge and needs a real-browser UA.

| party_id | party | fetch from |
|---|---|---|
| ap | Arbeiderpartiet | arbeiderpartiet.no (fl_attachment PDF) |
| frp | Fremskrittspartiet | frp.no/files/Program/2025/... |
| h | Høyre | res.cloudinary.com/hoyre/... |
| krf | Kristelig Folkeparti | krf.no/content/uploads/2025/05/... |
| mdg | Miljøpartiet De Grønne | mdg.no/_service/.../download |
| rodt | Rødt | roedt.no/fil/... (needs browser UA) |
| sp | Senterpartiet | senterpartiet.no/.../attachment/inline |
| sv | Sosialistisk Venstreparti | sv.no/wp-content/uploads/... |
| v | Venstre | venstre.no/politikk/partiprogram/... |

## What to extract

For each of the nine dimensions the site already uses (see
`docs/research-memo.md`), find the sentences in that party's programme that
speak to it. One programme point per concrete promise:

- the **quote**, verbatim, short (one sentence / one bullet), the party's own
  words, no paraphrase;
- the **topic** in a few words;
- the **dimension** it maps to;
- the **page or chapter** it sits in.

A programme point is not a position. It is the party's stated intent, tier C,
same class as the `partiprogram` evidence already on file. It never becomes a
coded position on a statement by itself — that needs a vote or a merknad
(tier A). The programme page shows the promise; the statement page shows, from
votes, whether the party acted on it.

## The bar

Quote exactly. If the programme says "redusere satsene" do not write "fjerne".
If a party says nothing on a dimension, that is a result: record it as an empty
list for that dimension, not a guessed point. A reader must be able to find
every quote by opening the PDF and searching for it.
