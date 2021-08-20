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

	foundCandidateNames := make(map[string]bool, 0)

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

				// Dependent keys may be duplicated where one
				// key is labels-only and other one contains
				// labels + attributes.
				//
				// Specifically in Terraform this applies to
				// a resource type depending on 'provider' attribute.
				//
				// We do need such dependent keys elsewhere
				// to know how to do completion within a block
				// but this doesn't matter when completing the label itself
				// unless/until we're also completing the dependent attributes.
				if _, ok := foundCandidateNames[label.Value]; ok {
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
					// TODO: AdditionalTextEdits:
					// - prefill required fields if body is empty
					// - prefill dependent attribute(s)
					Detail:      bodySchema.Detail,
					Description: bodySchema.Description,
				})
				foundCandidateNames[label.Value] = true
				count++
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
