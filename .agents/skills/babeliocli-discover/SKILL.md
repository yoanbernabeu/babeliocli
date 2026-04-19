---
name: babeliocli-discover
description: "Search the Babelio catalogue, open a book's detail page, and read its reader critiques using babeliocli. Use this skill whenever a user wants to find a book on Babelio, look up a book's synopsis / rating / publisher / page count, pull reviews about a specific title, or compare reader opinions programmatically."
---

# babeliocli — Search, Book Details, Reviews

## Search

```bash
babeliocli search "becky chambers" -f text
babeliocli search "1984 orwell"
```

Result JSON:

```json
{
  "query": "becky chambers",
  "count": 5,
  "results": [
    {
      "book_id": "1409270",
      "title": "Un psaume pour les recyclés sauvages",
      "author": "Becky Chambers",
      "author_id": "274210",
      "url": "https://www.babelio.com/livres/Chambers-Un-psaume-pour-les-recycles-sauvages/1409270",
      "avg_rating": 0
    }
  ]
}
```

Babelio ships two search layouts (rich cards and author-mosaic); the CLI handles both transparently. `avg_rating` is the viewer's own rating when logged in — treat it as hinting, not as the global average. Use `babeliocli book` for the real average.

## Book details

```bash
babeliocli book /livres/Simmons-Les-Cantos-dHyperion-tome-1--Hyperion-1/5603 -f text
```

The argument can be any of:

- full URL (`https://www.babelio.com/livres/SLUG/ID`)
- absolute path (`/livres/SLUG/ID`)
- bare `SLUG/ID`

JSON:

```json
{
  "book_id": "5603",
  "title": "Les Cantos d'Hypérion, tome 1 : Hypérion 1",
  "author": "Dan Simmons",
  "author_id": "2347",
  "url": "https://www.babelio.com/livres/Simmons-Les-Cantos-dHyperion-tome-1--Hyperion-1/5603",
  "synopsis": "Sur Hypérion, lointaine planète de l'Hégémonie …",
  "publisher": "Pocket",
  "pages": 282,
  "avg_rating": 4.15,
  "nb_ratings": 2418,
  "nb_reviews": 131
}
```

## Reviews / critiques

```bash
babeliocli reviews /livres/Simmons-Les-Cantos-dHyperion-tome-1--Hyperion-1/5603 --limit 20 -f text
```

Flags:

- `--limit N` (default `20`): stop after N reviews.
- `--max-pages N` (default `20`): cap on HTTP pagination.

JSON:

```json
{
  "book_url": "https://www.babelio.com/livres/Simmons-Les-Cantos-dHyperion-tome-1--Hyperion-1/5603",
  "count": 20,
  "reviews": [
    {
      "author": "HordeDuContre…",
      "author_url": "https://www.babelio.com/monprofil.php?id_user=…",
      "rating": 5,
      "date": "17 février 2022",
      "date_iso": "2022-02-17",
      "body": "C'est ce genre de livre. …"
    }
  ]
}
```

`rating` is an integer 0–5. `date_iso` is the best-effort conversion of the French date; may be empty if Babelio renders an unusual format.

## Agent recipes

### Summarize the mood around a book

```bash
babeliocli reviews /livres/.../5603 --limit 30 |
  jq -r '.reviews[] | "★\(.rating): \(.body[0:200])"' |
  head
```

### Compare ratings across editions

```bash
babeliocli search "hyperion simmons" |
  jq -r '.results[] | "\(.avg_rating)\t\(.title)"' |
  sort -rn
```

### Pick a random positive critique

```bash
babeliocli reviews /livres/.../5603 |
  jq '.reviews | map(select(.rating >= 4)) | .[].body'
```

## Caveats

- Reviews pages are HTML-rendered; counts match Babelio exactly but the body is whatever text the reviewer wrote (expect typos, emoji, inline quotes).
- The CLI de-duplicates by `book_id` within a single search, but two distinct editions of the same novel are separate `book_id`s.
- `nb_reviews` on the book page reflects Babelio's tab counter; reviews actually fetchable may be fewer if some were hidden by moderation.
