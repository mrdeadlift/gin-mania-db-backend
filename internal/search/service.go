package search

import "strings"

// Gin represents a gin entry in the catalogue.
type Gin struct {
	Name        string   `json:"name"`
	Country     string   `json:"country"`
	Botanicals  []string `json:"botanicals"`
	Description string   `json:"description"`
}

var catalogue = []Gin{
	{
		Name:        "Hendrick's",
		Country:     "Scotland",
		Botanicals:  []string{"cucumber", "rose"},
		Description: "Known for its delicate infusion of cucumber and rose petals.",
	},
	{
		Name:        "Tanqueray No. Ten",
		Country:     "England",
		Botanicals:  []string{"grapefruit", "lime", "chamomile"},
		Description: "A citrus-forward gin crafted in small batches with fresh fruits.",
	},
	{
		Name:        "Four Pillars Rare Dry",
		Country:     "Australia",
		Botanicals:  []string{"orange", "pepperberry", "lemon myrtle"},
		Description: "Combines native Australian botanicals with traditional gin notes.",
	},
	{
		Name:        "Ki No Bi Kyoto Dry",
		Country:     "Japan",
		Botanicals:  []string{"yuzu", "green tea", "ginger"},
		Description: "A dry gin that showcases Japanese botanicals from the Kyoto region.",
	},
}

// Search returns gins that match the given query.
func Search(query string) []Gin {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return catalogue
	}

	needle := strings.ToLower(trimmed)
	results := make([]Gin, 0, len(catalogue))

	for _, gin := range catalogue {
		if matches(gin, needle) {
			results = append(results, gin)
		}
	}

	return results
}

func matches(g Gin, needle string) bool {
	if strings.Contains(strings.ToLower(g.Name), needle) {
		return true
	}

	if strings.Contains(strings.ToLower(g.Country), needle) {
		return true
	}

	if strings.Contains(strings.ToLower(g.Description), needle) {
		return true
	}

	for _, botanical := range g.Botanicals {
		if strings.Contains(strings.ToLower(botanical), needle) {
			return true
		}
	}

	return false
}
