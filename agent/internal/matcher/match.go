// Package matcher finds complementary user pairs for physical meetups.
package matcher

import (
	"fmt"
	"math"
	"sort"

	"github.com/alexbreadman/public-charity/agent/internal/client"
)

const ScoreThreshold = 0.4

type Match struct {
	UserA           *client.User
	UserB           *client.User
	Score           float64
	SharedSlots     []string // overlapping availability
	SharedInterests []string
	NeedOfferPairs  []NeedOffer // what A needs that B offers, and vice versa
}

type NeedOffer struct {
	Needer   string // display name or email
	Category string
	Score    float64
	Offer    string // description from the offerer
}

// FindMatches takes all users and returns scored, ranked match candidates.
func FindMatches(users []client.User, maxResults int, isRecent func(a, b string) bool) []Match {
	// Group by cell cluster.
	cells := groupByCells(users)

	var candidates []Match
	seen := make(map[string]bool)

	for _, cluster := range cells {
		for i := 0; i < len(cluster); i++ {
			for j := i + 1; j < len(cluster); j++ {
				a, b := &cluster[i], &cluster[j]

				pairKey := pairKey(a.ID, b.ID)
				if seen[pairKey] {
					continue
				}
				seen[pairKey] = true

				if isRecent(a.ID, b.ID) {
					continue
				}

				m := scorePair(a, b)
				if m != nil {
					candidates = append(candidates, *m)
				}
			}
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > maxResults {
		candidates = candidates[:maxResults]
	}
	return candidates
}

func scorePair(a, b *client.User) *Match {
	// Availability overlap — required.
	shared := intersect(a.Avail, b.Avail)
	if len(shared) == 0 {
		return nil
	}

	// Need↔Offer matching.
	var score float64
	var pairs []NeedOffer

	bOfferCats := offerCategories(b)
	aOfferCats := offerCategories(a)

	for cat, s := range a.NeedScores {
		if s > ScoreThreshold {
			if desc, ok := bOfferCats[cat]; ok {
				score += s
				pairs = append(pairs, NeedOffer{
					Needer: displayName(a), Category: cat, Score: s, Offer: desc,
				})
			}
		}
	}
	for cat, s := range b.NeedScores {
		if s > ScoreThreshold {
			if desc, ok := aOfferCats[cat]; ok {
				score += s
				pairs = append(pairs, NeedOffer{
					Needer: displayName(b), Category: cat, Score: s, Offer: desc,
				})
			}
		}
	}

	if score == 0 {
		return nil
	}

	// Interest overlap — bonus.
	sharedInterests := intersect(a.Interests, b.Interests)
	score += float64(len(sharedInterests)) * 0.1

	return &Match{
		UserA: a, UserB: b,
		Score: score, SharedSlots: shared,
		SharedInterests: sharedInterests, NeedOfferPairs: pairs,
	}
}

func offerCategories(u *client.User) map[string]string {
	m := make(map[string]string)
	for _, o := range u.Offers {
		m[o.Category] = o.Description
	}
	return m
}

func displayName(u *client.User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Email
}

func intersect(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[v] = true
	}
	var out []string
	for _, v := range b {
		if set[v] {
			out = append(out, v)
		}
	}
	return out
}

// groupByCells builds one cluster per occupied cell. Each cluster contains
// all users in that cell plus all users in the 8 adjacent cells.
// Clusters intentionally overlap — a user may appear in multiple clusters.
// The pair dedup in FindMatches (via seen map) prevents duplicate match scoring.
func groupByCells(users []client.User) [][]client.User {
	byCell := make(map[string][]int) // cellID → user indices
	for i, u := range users {
		if u.CellID == "" {
			continue
		}
		byCell[u.CellID] = append(byCell[u.CellID], i)
	}

	var clusters [][]client.User
	for cellID := range byCell {
		adj := adjacentCells(cellID)
		var cluster []client.User
		for _, c := range adj {
			for _, idx := range byCell[c] {
				cluster = append(cluster, users[idx])
			}
		}
		if len(cluster) >= 2 {
			clusters = append(clusters, cluster)
		}
	}
	return clusters
}

func adjacentCells(cellID string) []string {
	var cy, cx int
	fmt.Sscanf(cellID, "%d_%d", &cy, &cx)
	cells := make([]string, 0, 9)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cells = append(cells, fmt.Sprintf("%d_%d", cy+dy, cx+dx))
		}
	}
	return cells
}

func pairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + ":" + b
}

// Haversine distance in miles (unused for now, but available for future refinement).
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 3959.0 // earth radius in miles
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
