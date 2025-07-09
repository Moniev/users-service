package schema

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type EntrepreneurDetails struct {
	ent.Schema
}

func (EntrepreneurDetails) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable().
			StructTag(`json:"id"`),
		field.String("business_name").
			Optional().
			StructTag(`json:"business_name"`),
		field.String("nip").
			Optional().
			Unique().
			StructTag(`json:"nip"`),
		field.String("krs").
			Optional().
			Unique().
			StructTag(`json:"krs"`),
		field.Text("description").
			Nillable().
			Optional().
			StructTag(`json:"description"`),
		field.Text("offer").
			Nillable().
			Optional().
			StructTag(`json:"offer"`),
		field.Float("income").
			Optional().
			Positive().
			StructTag(`json:"income"`),
		field.Float("costs").
			Optional().
			Positive().
			StructTag(`json:"costs"`),
		field.Float("funding_capital").
			Optional().
			Nillable().
			StructTag(`json:"funding_capital"`),
		field.String("industry").
			Optional().
			StructTag(`json:"industry"`),
		field.JSON("management_council_members", []string{}).
			Optional().
			StructTag(`json:"management_council_members"`),
		field.JSON("decision_makers", []string{}).
			Optional().
			StructTag(`json:"decision_makers"`),
		field.String("business_phone_number").
			Optional().
			StructTag(`json:"business_phone_number"`),
		field.String("business_mail").
			Optional().
			Nillable().
			StructTag(`json:"business_mail"`),
		field.String("website_address").
			Optional().
			Nillable().
			StructTag(`json:"website_address"`),
		field.Time("created_at").
			Immutable().
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`),
	}
}

func (EntrepreneurDetails) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_details", UserDetails.Type).
			Ref("entrepreneur_details").
			Unique().
			StructTag(`json:"user_details"`),
	}
}

func (EntrepreneurDetails) Hooks() []ent.Hook {
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
