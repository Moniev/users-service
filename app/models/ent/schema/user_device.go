package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserDevice struct {
	ent.Schema
}

func (UserDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			StructTag(`json:"id"`).
			Unique().
			Immutable(),

		field.String("name").
			StructTag(`json:"name"`).
			Optional().
			Comment("Name of the device"),

		field.String("type").
			StructTag(`json:"type"`).
			Optional().
			Comment("type of the device"),

		field.String("token").
			StructTag(`json:"token"`).
			Comment("The unique fingerprint visitorId from FingerprintJS.").
			NotEmpty(),

		field.String("ip_address").
			StructTag(`json:"ip_address,omitempty"`).
			Comment("The last known IP address of the device.").
			Optional(),

		field.String("user_agent").
			StructTag(`json:"user_agent,omitempty"`).
			Comment("The full User-Agent string from the last request.").
			Optional(),

		field.String("os_name").
			StructTag(`json:"os_name,omitempty"`).
			Comment("Operating system name (e.g., Windows, macOS).").
			Optional(),

		field.String("os_version").
			StructTag(`json:"os_version,omitempty"`).
			Comment("Operating system version (e.g., 11, 14.5).").
			Optional(),

		field.String("browser_name").
			StructTag(`json:"browser_name,omitempty"`).
			Comment("Browser name (e.g., Chrome, Firefox).").
			Optional(),

		field.String("browser_version").
			StructTag(`json:"browser_version,omitempty"`).
			Comment("Browser version (e.g., 126.0).").
			Optional(),

		field.Time("last_seen_at").
			StructTag(`json:"last_seen_at"`).
			Comment("Timestamp of the last time this device was used for authentication.").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("created_at").
			StructTag(`json:"created_at"`).
			Default(time.Now),
		field.Time("updated_at").
			StructTag(`json:"updated_at"`).
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (UserDevice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("user_devices").
			Unique().
			Required().
			StructTag(`json:"owner"`),
		edge.From("user_settings_2fa", UserSettings.Type).
			Ref("second_factor_target").
			Unique().
			StructTag(`json:"user_settings_2fa,omitempty"`),
		edge.From("second_factor_codes", SecondFactorCode.Type).
			Ref("target_user_device").
			Unique().
			StructTag(`json:"second_factor_codes"`),
	}
}
