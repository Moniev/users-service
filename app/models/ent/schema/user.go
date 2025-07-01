package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),
		field.String("mail").
			NotEmpty().
			Unique().
			StructTag(`json:"mail"`),
		field.String("phone").
			Optional().
			Unique().
			StructTag(`json:"phone"`),
		field.String("password").
			NotEmpty().
			Unique(),
		field.Bool("active").
			Default(false).
			StructTag(`json:"active"`),
		field.Bool("verified").
			Default(false).
			StructTag(`json:"verified"`),
		field.Bool("blacklisted").
			Default(false).
			StructTag(`json:"blacklisted"`),
		field.Bool("removed").
			Default(false).
			StructTag(`json:"removed"`),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_details", UserDetails.Type).
			Unique().
			StructTag(`json:"user_details"`),
		edge.To("user_settings", UserSettings.Type).
			Unique().
			StructTag(`json:"user_settings"`),
		edge.To("activation_code", ActivationCode.Type).
			Unique().
			StructTag(`json:"activation_code"`),
		edge.To("verification_code", VerificationCode.Type).
			Unique().
			StructTag(`json:"verification_code"`),
		edge.To("second_factor_code", SecondFactorCode.Type).
			Unique().
			StructTag(`json:"second_factor_code"`),
		edge.To("reset_code", ResetCode.Type).
			Unique().
			StructTag(`json:"reset_code"`),
		edge.To("user_devices", UserDevice.Type).
			StructTag(`json:"user_devices"`),
		edge.To("user_actions", UserAction.Type).
			StructTag(`json:"user_actions"`),
		edge.To("user_roles", UserRole.Type).
			StructTag(`json:"user_roles"`),
	}
}
