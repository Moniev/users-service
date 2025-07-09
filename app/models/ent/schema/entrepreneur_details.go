package schema

import "entgo.io/ent"

type EntrepreneurDetails struct {
	ent.Schema
}

func (EntrepreneurDetails) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (EntrepreneurDetails) Fields() []ent.Field {
	return []ent.Field{}
}
