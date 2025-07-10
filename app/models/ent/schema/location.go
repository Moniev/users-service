package schema

import (
	"context"
	"time"

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
		field.Time("created_at").
			Immutable().
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`),
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

func (Location) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if m.Op().Is(ent.OpCreate) {
					if err := m.SetField("created_at", time.Now().UTC()); err != nil {
						return nil, err
					}

					if err := m.SetField("updated_at", time.Now().UTC()); err != nil {
						return nil, err
					}

				} else if m.Op().Is(ent.OpUpdate) || m.Op().Is(ent.OpUpdateOne) {
					if err := m.SetField("updated_at", time.Now().UTC()); err != nil {
						return nil, err
					}
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}
