package schema

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserSettings struct {
	ent.Schema
}

func (UserSettings) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.Text("UUID").
			NotEmpty().
			StructTag(`json:"uuid"`),
		field.Bool("two_factor").
			Default(false).
			StructTag(`json:"two_factor"`),
		field.Bool("night_mode").
			Default(false).
			StructTag(`json:"night_mode"`),
		field.Time("created_at").
			Immutable().
			StructTag(`json:"created_at"`),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`),
	}
}

func (UserSettings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("user_settings").
			Unique().
			Required().
			StructTag(`json:"owner"`),
		edge.To("second_factor_target", UserDevice.Type).
			Unique().
			StructTag(`json:"second_factor_target,omitempty"`),
		edge.To("notification_target_devices", UserDevice.Type).
			StructTag(`json:"notification_target_devices,omitempty"`),
	}
}

func (UserSettings) Hooks() []ent.Hook {
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
