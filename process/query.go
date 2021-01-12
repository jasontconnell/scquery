package process

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jasontconnell/scquery/process/fields"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

func RunQuery(connstr string, templateId uuid.UUID, queryFields []FieldQuery, resultFields []FieldResult) ([]ItemResult, error) {
	items, err := api.LoadItems(connstr)
	if err != nil {
		return nil, fmt.Errorf("getting items: %w", err)
	}

	tlist, err := api.LoadTemplates(connstr)
	if err != nil {
		return nil, fmt.Errorf("getting templates: %w", err)
	}

	flds, err := api.LoadFieldsParallel(connstr, 12)
	if err != nil {
		return nil, fmt.Errorf("getting fields: %w", err)
	}

	tmap := api.GetTemplateMap(tlist)

	_, imap := api.LoadItemMap(items)

	api.AssignFieldValues(imap, flds)
	api.SetTemplates(imap, tmap)

	fmap := api.FilterItemMapCustom(imap, func(node data.ItemNode) bool {
		inc := false
		if node.GetTemplateId() == templateId {
			inc = true
		}
		return inc
	})

	template := tmap[templateId]

	tfields := template.GetAllFields()
	tfm := make(map[string]uuid.UUID)
	for _, f := range tfields {
		tfm[f.GetName()] = f.GetId()
	}

	fieldIds := []uuid.UUID{}
	for _, f := range resultFields {
		fld, ok := tfm[f.Name]
		if ok {
			fieldIds = append(fieldIds, fld)
		}
	}

	results := []ItemResult{}
	for _, item := range fmap {
		allvals := item.GetVersionedFieldValues()
		if len(allvals) == 0 {
			continue
		}

		vals := item.GetLatestVersionFieldKeys(data.English)
		idm := make(map[uuid.UUID]data.FieldValueKey)
		for _, v := range vals {
			idm[v.FieldId] = v
		}
		res := ItemResult{Id: item.GetId(), Fields: []ItemFieldResult{}}
		for _, v := range fieldIds {
			fvk := idm[v]
			fv, ok := allvals[fvk]
			if !ok {
				continue
			}

			result, err := fields.Process(imap, tmap, fv)
			if err != nil {
				log.Println("error for field", v, "in item", item.GetId(), err)
				continue
			}

			res.Fields = append(res.Fields, ItemFieldResult{Name: fv.GetName(), Value: result.Value})
		}

		results = append(results, res)
	}

	return results, nil
}
