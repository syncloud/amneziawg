package peers

import "strings"

type Resolver struct {
	Servers []string
	Domain  string
}

var PublicResolver = Resolver{
	Servers: []string{"8.8.8.8", "2001:4860:4860::8888"},
}

func (r Resolver) String() string {
	entries := r.Servers
	if r.Domain != "" {
		entries = append(append([]string{}, r.Servers...), r.Domain)
	}
	return strings.Join(entries, ", ")
}
