package schema

import (
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
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
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
