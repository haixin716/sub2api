// Package schema 定义 Ent ORM 的数据库 schema。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RequestLog 定义请求日志实体的 schema。
//
// 请求日志记录每次 API 调用的完整请求和响应内容，用于调试、审计和问题排查。
// 这是一个只追加的表，不支持更新和删除。
type RequestLog struct {
	ent.Schema
}

// Annotations 返回 schema 的注解配置。
func (RequestLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "request_logs"},
	}
}

// Fields 定义请求日志实体的所有字段。
func (RequestLog) Fields() []ent.Field {
	return []ent.Field{
		// 关联字段
		field.Int64("user_id"),
		field.Int64("api_key_id"),
		field.Int64("account_id"),
		field.String("client_request_id").
			MaxLen(64).
			NotEmpty().
			Comment("内部请求ID，由网关生成，用于关联 usage_logs 和 request_logs"),
		field.String("request_id").
			MaxLen(64).
			Optional().
			Nillable().
			Comment("上游 API 返回的请求ID（x-request-id响应头，可能为空）"),
		field.String("model").
			MaxLen(100).
			NotEmpty().
			Comment("模型名称"),
		field.Int64("group_id").
			Optional().
			Nillable().
			Comment("分组ID"),

		// 请求信息
		field.Text("request_body").
			NotEmpty().
			Comment("完整请求体JSON"),
		field.String("request_method").
			MaxLen(10).
			Default("POST").
			Comment("HTTP方法"),
		field.String("request_path").
			MaxLen(255).
			Default("/v1/messages").
			Comment("请求路径"),

		// 响应信息
		field.Text("response_body").
			Optional().
			Nillable().
			Comment("完整响应体JSON（非流式）或聚合内容（流式）"),
		field.Int("response_status").
			Default(200).
			Comment("HTTP状态码"),

		// 元数据
		field.Bool("stream").
			Default(false).
			Comment("是否流式响应"),
		field.Int("duration_ms").
			Optional().
			Nillable().
			Comment("请求耗时（毫秒）"),
		field.String("ip_address").
			MaxLen(45). // 支持 IPv6
			Optional().
			Nillable().
			Comment("客户端IP地址"),
		field.String("user_agent").
			MaxLen(512).
			Optional().
			Nillable().
			Comment("User-Agent头信息"),

		// 错误信息
		field.Bool("is_error").
			Default(false).
			Comment("是否错误请求"),
		field.Text("error_message").
			Optional().
			Nillable().
			Comment("错误信息"),
		field.String("error_type").
			MaxLen(50).
			Optional().
			Nillable().
			Comment("错误类型"),

		// 时间戳（只有 created_at，日志不可修改）
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("创建时间"),
	}
}

// Edges 定义请求日志实体的关联关系。
func (RequestLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("request_logs").
			Field("user_id").
			Required().
			Unique(),
		edge.From("api_key", APIKey.Type).
			Ref("request_logs").
			Field("api_key_id").
			Required().
			Unique(),
		edge.From("account", Account.Type).
			Ref("request_logs").
			Field("account_id").
			Required().
			Unique(),
		edge.From("group", Group.Type).
			Ref("request_logs").
			Field("group_id").
			Unique(),
	}
}

// Indexes 定义数据库索引，优化查询性能。
func (RequestLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("api_key_id"),
		index.Fields("account_id"),
		index.Fields("group_id"),
		index.Fields("client_request_id"), // 主关联键，用于关联 usage_logs
		index.Fields("request_id"),        // 辅助，用于上游追踪
		index.Fields("created_at"),
		index.Fields("model"),
		index.Fields("is_error"),
		// 复合索引用于时间范围查询
		index.Fields("user_id", "created_at"),
		index.Fields("api_key_id", "created_at"),
		index.Fields("is_error", "created_at"),
	}
}
