package manifest

const ScoreThreshold = 0.4

var Categories = []string{
	"company",
	"tools",
	"transport",
	"grocery",
	"errands",
	"admin",
	"skills",
	"wellness",
	"digital",
	"community",
}

type NeedScores struct {
	Company   float64 `json:"company"`
	Tools     float64 `json:"tools"`
	Transport float64 `json:"transport"`
	Grocery   float64 `json:"grocery"`
	Errands   float64 `json:"errands"`
	Admin     float64 `json:"admin"`
	Skills    float64 `json:"skills"`
	Wellness  float64 `json:"wellness"`
	Digital   float64 `json:"digital"`
	Community float64 `json:"community"`
}

func (s NeedScores) Get(category string) float64 {
	switch category {
	case "company":
		return s.Company
	case "tools":
		return s.Tools
	case "transport":
		return s.Transport
	case "grocery":
		return s.Grocery
	case "errands":
		return s.Errands
	case "admin":
		return s.Admin
	case "skills":
		return s.Skills
	case "wellness":
		return s.Wellness
	case "digital":
		return s.Digital
	case "community":
		return s.Community
	}
	return 0
}

func (s NeedScores) AboveThreshold(category string) bool {
	return s.Get(category) > ScoreThreshold
}

func (s NeedScores) ActiveCategories() []string {
	var active []string
	for _, c := range Categories {
		if s.AboveThreshold(c) {
			active = append(active, c)
		}
	}
	return active
}
