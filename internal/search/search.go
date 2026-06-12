// Package search provides fast fuzzy filtering over flattened hotkeys.
package search

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// Item is one searchable hotkey, flattened out of its cheatsheet/section.
type Item struct {
	Sheet   string
	Section string
	Keys    string
	Desc    string
}

// fields are the individually-searchable texts of an item. Matching happens
// per field (not over a concatenation) so a fuzzy subsequence can't bleed
// across field boundaries and produce surprising hits.
func (i Item) fields() []string {
	return []string{
		strings.ToLower(i.Keys),
		strings.ToLower(i.Desc),
		strings.ToLower(i.Section),
		strings.ToLower(i.Sheet),
	}
}

// Filter returns the items matching query, best matches first. Every
// whitespace-separated token must fuzzy-match at least one field of an item
// for it to be included. An empty/whitespace-only query returns items
// unchanged, in their original order.
func Filter(items []Item, query string) []Item {
	tokens := strings.Fields(strings.ToLower(query))
	if len(tokens) == 0 {
		return items
	}

	type scored struct {
		item  Item
		score int
		order int
	}
	var hits []scored

	for idx, it := range items {
		fields := it.fields()
		total, ok := 0, true
		for _, tok := range tokens {
			best, found := bestTokenScore(tok, fields)
			if !found {
				ok = false
				break
			}
			total += best
		}
		if ok {
			hits = append(hits, scored{item: it, score: total, order: idx})
		}
	}

	// Highest score first; stable on original order for ties.
	sort.SliceStable(hits, func(a, b int) bool {
		if hits[a].score != hits[b].score {
			return hits[a].score > hits[b].score
		}
		return hits[a].order < hits[b].order
	})

	out := make([]Item, len(hits))
	for i, h := range hits {
		out[i] = h.item
	}
	return out
}

// bestTokenScore returns the best fuzzy score of tok across fields, and whether
// it matched any field at all.
func bestTokenScore(tok string, fields []string) (int, bool) {
	matches := fuzzy.Find(tok, fields)
	if len(matches) == 0 {
		return 0, false
	}
	return matches[0].Score, true // fuzzy.Find returns best match first
}
