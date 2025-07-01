package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type ActivationCode struct {
	ent.Schema
}

func (ActivationCode) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("code").
			Unique().
			StructTag(`json:"code"`),
		field.Bool("used").
			Default(false).
			StructTag(`json:"used"`),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("expires_at").
			StructTag(`json:"expires_at"`).
			Default(time.Now().UTC().Add(time.Hour * 12)),
	}
}

func (ActivationCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("activation_code").
			Unique().
			Required().
			StructTag(`json:"owner"`),
	}
}
