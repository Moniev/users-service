package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Location struct {
	ent.Schema
}

func (Location) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("country").
			NotEmpty().
			StructTag(`json:"country"`),
		field.String("province").
			NotEmpty().
			StructTag(`json:"province"`),
		field.String("city").
			NotEmpty().
			StructTag(`json:"city"`),
		field.String("postal_code").
			NotEmpty().
			StructTag(`json:"postal_code"`),
		field.String("street").
			Optional().
			StructTag(`json:"street"`),
		field.Int("building_number").
			Positive().
			StructTag(`json:"building_number"`),
		field.Int("apartment_number").
			NonNegative().
			StructTag(`json:"apartment_number"`),
	}
}

func (Location) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_details", UserDetails.Type).
			Ref("locations").
			Unique().
			Required().
			StructTag(`json:"user_details"`),
	}
}
