package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type VerificationCode struct {
	ent.Schema
}

func (VerificationCode) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("code").
			Unique().
			NotEmpty().
			StructTag(`json:"code"`),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("expires_at").
			StructTag(`json:"expires_at"`).
			Default(time.Now().UTC().Add(time.Hour * 1)),
	}
}

func (VerificationCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("verification_code").
			Unique().
			Required().
			StructTag(`json:"owner"`),
	}
}
