package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RolePermission struct {
	ent.Schema
}

func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("name").
			Unique().
			StructTag(`json:"name"`),
		field.Bool("view").
			StructTag(`json:"view"`),
		field.Bool("add").
			StructTag(`json:"add"`),
		field.Bool("edit").
			StructTag(`json:"edit"`),
		field.Bool("delete").
			StructTag(`json:"delete"`),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (RolePermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_role", UserRole.Type).
			StructTag(`json:"user_role"`),
	}
}
