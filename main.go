package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jasontconnell/scquery/conf"
	"github.com/jasontconnell/scquery/process"
)

func main() {
	c := flag.String("c", "config.json", "config filename with connectionString")
	tid := flag.String("template", "", "template id to search")
	path := flag.String("path", "", "path to search or include")
	flds := flag.String("fields", "", "csv of fields and values, like Field=Value to search")
	output := flag.String("output", "", "csv of fields to output per line")
	order := flag.String("order", "", "order by field")
	flag.Parse()

	if *tid == "" {
		flag.PrintDefaults()
		return
	}

	start := time.Now()
	cfg := conf.LoadConfig(*c)

	log.Println(*tid, *path, *flds, *output, cfg.ConnectionString)

	templateId := uuid.Must(uuid.Parse(*tid))

	resultFields := []process.FieldResult{}
	if output != nil {
		sp := strings.Split(*output, ",")
		for _, s := range sp {
			resultFields = append(resultFields, process.FieldResult{Name: strings.Trim(s, " ")})
		}
	}

	res, err := process.RunQuery(cfg.ConnectionString, templateId, nil, resultFields)
	if err != nil {
		log.Fatal(err)
	}

	if *order != "" && len(res) > 0 {
		orderIndex := -1
		fst := res[0]

		for i, f := range fst.Fields {
			if f.Name == *order {
				orderIndex = i
				break
			}
		}

		log.Println(*order, "order index", orderIndex, fst.Fields[orderIndex])

		sort.Slice(res, func(i, j int) bool {
			iv, jv := res[i].Id.String(), res[j].Id.String()
			if orderIndex != -1 && orderIndex < len(res[i].Fields) {
				iv, jv = res[i].Fields[orderIndex].Value, res[j].Fields[orderIndex].Value
			}
			return iv < jv
		})
	}

	for _, r := range res {
		line := r.Id.String()
		for _, f := range r.Fields {
			line += "," + f.Value
		}
		fmt.Println(line)
	}

	log.Println("finished", time.Since(start))
}
