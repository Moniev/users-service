package schema

import (
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
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
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
