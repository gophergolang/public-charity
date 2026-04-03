package site

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/gophergolang/public-charity/internal/manifest"
	"github.com/gophergolang/public-charity/internal/storage"
)

//go:embed templates/*.html
var templateFS embed.FS

type ScoreDisplay struct {
	Label   string
	Value   float64
	Percent int
}

type ProfileData struct {
	Username  string
	Bio       string
	ScoreList []ScoreDisplay
	Interests []string
	Offers    []manifest.Offer
	Surplus   []manifest.Surplus
	Timeline  []manifest.TimelineEntry
}

func Generate(u *manifest.User) error {
	tmpl, err := template.ParseFS(templateFS, "templates/base.html", "templates/profile.html")
	if err != nil {
		return err
	}

	data := ProfileData{
		Username:  u.Username,
		Bio:       u.Bio,
		Interests: u.Interests,
		Offers:    u.Offers,
		Surplus:   u.Surplus,
		Timeline:  u.Timeline,
		ScoreList: buildScoreList(u.NeedScores),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	path := "sites/" + u.Username + "/index.html"
	return storage.WriteRaw(path, buf.Bytes())
}

func buildScoreList(scores manifest.NeedScores) []ScoreDisplay {
	list := make([]ScoreDisplay, 0, len(manifest.Categories))
	for _, cat := range manifest.Categories {
		val := scores.Get(cat)
		list = append(list, ScoreDisplay{
			Label:   capitalize(cat),
			Value:   val,
			Percent: int(val * 100),
		})
	}
	return list
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
