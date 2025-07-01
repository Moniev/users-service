package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserDetails struct {
	ent.Schema
}

func (UserDetails) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty().
			StructTag(`json:"name"`),
		field.String("first_name").
			NotEmpty().
			StructTag(`json:"first_name"`),
		field.String("last_name").
			NotEmpty().
			StructTag(`json:"last_name"`),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (UserDetails) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("user_details").
			Unique().
			Required().
			StructTag(`json:"owner"`),
	}
}
