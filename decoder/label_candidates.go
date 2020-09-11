package decoder

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func (d *Decoder) labelCandidatesFromDependentSchema(idx int, db map[schema.SchemaKey]*schema.BodySchema, prefixRng, editRng hcl.Range) (lang.Candidates, error) {
	candidates := lang.NewCandidates()
	count := 0

	prefix, _ := d.bytesFromRange(prefixRng)

	for schemaKey, bodySchema := range db {
		depKeys, err := decodeSchemaKey(schemaKey)
		if err != nil {
			// key undecodable
			continue
		}

		for _, label := range depKeys.Labels {
			if label.Index == idx {
				if uint(count) >= d.maxCandidates {
					// reached maximum no of candidates
					return candidates, nil
				}
				if len(prefix) > 0 && !strings.HasPrefix(label.Value, string(prefix)) {
					continue
				}
				candidates.List = append(candidates.List, lang.Candidate{
					Label:        label.Value,
					Kind:         lang.LabelCandidateKind,
					IsDeprecated: bodySchema.IsDeprecated,
					TextEdit: lang.TextEdit{
						NewText: label.Value,
						Snippet: label.Value,
						Range:   editRng,
					},
					// TODO: AdditionalTextEdits (required fields if body is empty)
					Detail:      bodySchema.Detail,
					Description: bodySchema.Description,
				})
			}
		}
	}

	candidates.IsComplete = true

	// TODO: sort by more metadata, such as IsDeprecated
	sort.Slice(candidates.List, func(i, j int) bool {
		return candidates.List[i].Label < candidates.List[j].Label
	})

	return candidates, nil
}

func decodeSchemaKey(key schema.SchemaKey) (schema.DependencyKeys, error) {
	var dk schema.DependencyKeys
	err := json.Unmarshal([]byte(key), &dk)
	return dk, err
}
