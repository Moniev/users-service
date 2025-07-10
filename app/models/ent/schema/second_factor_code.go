package schema

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type SecondFactorCode struct {
	ent.Schema
}

func (SecondFactorCode) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.Bool("used").
			Default(false).
			StructTag(`json:"used"`),
		field.String("code").
			Unique().
			StructTag(`json:"code"`),
		field.Time("created_at").
			Immutable().
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`),
		field.Time("expires_at").
			StructTag(`json:"expires_at"`).
			Default(time.Now().UTC().Add(time.Minute * 5)),
	}
}

func (SecondFactorCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("second_factor_code").
			Unique().
			Required().
			StructTag(`json:"owner"`),
		edge.To("target_user_device", UserDevice.Type).
			Unique().
			Required().
			StructTag(`json:"target_user_device"`),
	}
}

func (SecondFactorCode) Hooks() []ent.Hook {
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
