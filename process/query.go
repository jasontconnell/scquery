package process

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jasontconnell/scquery/process/fields"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

func RunQuery(connstr string, templateId uuid.UUID, path string, queryFields []FieldQuery, resultFields []FieldResult, lang data.Language) ([]ItemResult, error) {
	log.Println("loading items")
	items, err := api.LoadItems(connstr)
	if err != nil {
		return nil, fmt.Errorf("getting items: %w", err)
	}
	_, imap := api.LoadItemMap(items)
	log.Println("loaded", len(items), "items")

	log.Println("loading templates")
	tlist, err := api.LoadTemplates(connstr)
	if err != nil {
		return nil, fmt.Errorf("getting templates: %w", err)
	}
	tmap := api.GetTemplateMap(tlist)
	log.Println("loaded", len(tlist), "templates")

	var fieldIds []uuid.UUID
	if templateId != uuid.Nil {
		t := tmap[templateId]
		flds := t.GetFields()
		for _, f := range flds {
			fieldIds = append(fieldIds, f.GetId())
		}
	}

	log.Println("loading fields for", len(fieldIds), "fields")
	flds, err := api.LoadFilteredFieldValues(connstr, fieldIds, 12)
	if err != nil {
		return nil, fmt.Errorf("getting fields: %w", err)
	}

	api.AssignFieldValues(imap, flds)
	api.SetTemplates(imap, tmap)

	var fmap data.ItemMap

	if templateId != uuid.Nil {
		fmap = api.FilterItemMapCustom(imap, func(node data.ItemNode) bool {
			return node.GetTemplateId() == templateId
		})
	} else {
		fmap = api.FilterItemMap(imap, []string{path})
	}

	allFieldMap := populateTemplateFields(fmap)

	fmapc := api.FilterItemMapCustom(fmap, func(item data.ItemNode) bool {
		if len(queryFields) == 0 {
			return true
		}
		ok := true
		for _, fq := range queryFields {
			tfld := item.GetTemplate().FindField(fq.FieldName)
			if tfld == nil {
				ok = false
				continue
			}

			val := item.GetFieldValue(tfld.GetId(), lang)
			valstr := ""
			if val != nil {
				valstr = val.GetValue()
			}

			switch fq.Op {
			case "=":
				ok = ok && valstr == fq.Value
			case "<>":
				ok = ok && valstr != fq.Value
			case "~":
				ok = ok && strings.Contains(valstr, fq.Value)
			}
		}

		return ok
	})

	resultItems := []data.ItemNode{}
	for _, v := range fmapc {
		resultItems = append(resultItems, v)
	}

	results := []ItemResult{}
	for _, item := range resultItems {
		allvals := item.GetVersionedFieldValues()
		if len(allvals) == 0 {
			continue
		}

		calcFields := []string{}
		fieldIds := []uuid.UUID{}
		tfm, ok := allFieldMap[item.GetTemplateId()]

		if !ok {
			continue
		}

		for _, f := range resultFields {
			fld, ok := tfm[f.Name]
			if ok {
				fieldIds = append(fieldIds, fld)
			} else {
				calcFields = append(calcFields, f.Name)
			}
		}

		vals := item.GetLatestVersionFieldKeys(lang)

		idm := make(map[uuid.UUID]data.FieldValueKey)
		for _, v := range vals {
			idm[v.FieldId] = v
		}
		res := ItemResult{Id: item.GetId(), Fields: []ItemFieldResult{}}

		for _, calc := range calcFields {
			str := ""
			switch calc {
			case "Name":
				str = item.GetName()
			case "Path":
				str = item.GetPath()
			case "ID":
				str = item.GetId().String()
			}

			res.Fields = append(res.Fields, ItemFieldResult{Name: calc, Value: str})
		}

		add := true
		for _, v := range fieldIds {
			fvk := idm[v]
			fv, ok := allvals[fvk]

			if !ok {
				log.Println("couldn't find", v, fvk)
				add = false
				break
			}

			result, err := fields.Process(imap, tmap, fv)
			if err != nil {
				log.Println("error for field", v, "in item", item.GetId(), err)
				add = false
				break
			}

			res.Fields = append(res.Fields, ItemFieldResult{Name: fv.GetName(), Value: result.Value})
		}

		if add {
			results = append(results, res)
		}
	}

	log.Println("len results", len(results))
	return results, nil
}

func populateTemplateFields(m data.ItemMap) map[uuid.UUID]map[string]uuid.UUID {
	results := make(map[uuid.UUID]map[string]uuid.UUID)
	for _, node := range m {
		if _, ok := results[node.GetTemplateId()]; ok {
			continue
		}

		fieldMap := make(map[string]uuid.UUID)

		template := node.GetTemplate()
		tfields := template.GetAllFields()
		for _, f := range tfields {
			fieldMap[f.GetName()] = f.GetId()
		}

		results[node.GetTemplateId()] = fieldMap
	}
	return results
}
