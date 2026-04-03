package agent

import (
	"github.com/gophergolang/public-charity/internal/manifest"
)

type Match struct {
	Users   []*manifest.User
	Rule    manifest.BusinessRule
	BizName string
	Overlap []string // shared interests
}

func FindMatches(rule manifest.BusinessRule, candidates []*manifest.User) []Match {
	// Filter by minimum score
	var qualified []*manifest.User
	for _, u := range candidates {
		if u.NeedScores.Get(rule.Category) >= rule.MinScore {
			qualified = append(qualified, u)
		}
	}

	if len(qualified) == 0 {
		return nil
	}

	switch rule.MatchType {
	case "pair":
		return findPairs(rule, qualified)
	case "broadcast":
		return broadcastMatches(rule, qualified)
	default:
		return broadcastMatches(rule, qualified)
	}
}

func findPairs(rule manifest.BusinessRule, users []*manifest.User) []Match {
	var matches []Match
	seen := make(map[string]bool)

	for i := 0; i < len(users); i++ {
		if seen[users[i].Username] {
			continue
		}
		for j := i + 1; j < len(users); j++ {
			if seen[users[j].Username] {
				continue
			}
			overlap := interestOverlap(users[i].Interests, users[j].Interests)
			if len(overlap) >= rule.InterestOverlapMin {
				matches = append(matches, Match{
					Users:   []*manifest.User{users[i], users[j]},
					Rule:    rule,
					Overlap: overlap,
				})
				seen[users[i].Username] = true
				seen[users[j].Username] = true
				break
			}
		}
	}
	return matches
}

func broadcastMatches(rule manifest.BusinessRule, users []*manifest.User) []Match {
	var matches []Match
	for _, u := range users {
		matches = append(matches, Match{
			Users: []*manifest.User{u},
			Rule:  rule,
		})
	}
	return matches
}

func interestOverlap(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[v] = true
	}
	var overlap []string
	for _, v := range b {
		if set[v] {
			overlap = append(overlap, v)
		}
	}
	return overlap
}
