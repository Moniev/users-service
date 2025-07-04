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
			StructTag(`json:"-"`).
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

		field.JSON("subscription_ids", []int{}).
			Default([]int{}).
			StructTag(`json:"subscription_ids"`),
		field.JSON("team_ids", []int{}).
			Default([]int{}).
			StructTag(`json:"team_ids"`),
		field.JSON("organization_ids", []int{}).
			Default([]int{}).
			StructTag(`json:"organization_ids"`),
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
			StructTag(`json:"-"`).
			Unique(),
		edge.To("verification_code", VerificationCode.Type).
			StructTag(`json:"-"`).
			Unique(),
		edge.To("second_factor_code", SecondFactorCode.Type).
			StructTag(`json:"-"`).
			Unique(),
		edge.To("reset_code", ResetCode.Type).
			StructTag(`json:"-"`).
			Unique(),
		edge.To("user_devices", UserDevice.Type).
			StructTag(`json:"user_devices"`),
		edge.To("user_actions", UserAction.Type).
			StructTag(`json:"-"`),
		edge.To("user_roles", UserRole.Type).
			StructTag(`json:"user_roles"`),
	}
}
