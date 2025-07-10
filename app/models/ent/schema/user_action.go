package schema

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserAction struct {
	ent.Schema
}

func (UserAction) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("type").
			StructTag(`json:"type"`),
		field.String("action").
			StructTag(`json:"action"`),
		field.String("details").
			StructTag(`json:"details"`),
		field.Time("created_at").
			Immutable().
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`),
	}
}

func (UserAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("user_actions").
			Required().
			StructTag(`json:"author"`),
	}
}

func (UserAction) Hooks() []ent.Hook {
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
