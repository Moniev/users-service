package schema

import (
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
			StructTag(`json:"uuid"`),
		field.Bool("two_factor").
			Default(false).
			StructTag(`json:"two_factor"`),
		field.Bool("night_mode").
			Default(false).
			StructTag(`json:"night_mode"`),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
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
