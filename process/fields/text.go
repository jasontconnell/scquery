package fields

import "github.com/jasontconnell/sitecore/data"

type textProcessor struct {
	value string
}

func (t textProcessor) Process(m data.ItemMap, s string) (FieldProcessResult, error) {
	return FieldProcessResult{Value: s, Raw: s}, nil
}
