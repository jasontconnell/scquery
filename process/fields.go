package process

import (
	"regexp"
)

var freg *regexp.Regexp = regexp.MustCompile(`(.+)(=|<>|~)'(.*)'`)

func GetFieldsQuery(fields []string) []FieldQuery {
	list := []FieldQuery{}
	for _, f := range fields {
		g := freg.FindStringSubmatch(f)
		if len(g) == 0 {
			continue
		}

		fq := FieldQuery{FieldName: g[1], Op: g[2], Value: g[3]}
		list = append(list, fq)
	}
	return list
}
