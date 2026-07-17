// 请求日志类型定义

export interface RequestLog {
  id: number
  user_id: number
  api_key_id: number
  account_id: number
  client_request_id: string  // 内部请求ID，由网关生成
  request_id?: string | null // 上游API返回的请求ID（可能为空）
  model: string
  group_id?: number

  // 请求信息
  request_body: string
  request_method: string
  request_path: string

  // 响应信息
  response_body?: string
  response_status: number

  // 元数据
  stream: boolean
  duration_ms?: number
  user_agent?: string

  // 错误信息
  is_error: boolean
  error_message?: string
  error_type?: string

  created_at: string

  // 关联实体
  user?: {
    id: number
    email: string
  }
  api_key?: {
    id: number
    name: string
    key: string
  }
  group?: {
    id: number
    name: string
  }
}

// 管理员请求日志（包含额外字段）
export interface AdminRequestLog extends RequestLog {
  ip_address?: string
  account?: {
    id: number
    name: string
  }
}

// 请求日志查询参数
export interface RequestLogQueryParams {
  page?: number
  page_size?: number
  api_key_id?: number
  model?: string
  stream?: boolean
  is_error?: boolean
  start_date?: string
  end_date?: string
  timezone?: string
}

// 管理员请求日志查询参数
export interface AdminRequestLogQueryParams extends RequestLogQueryParams {
  user_id?: number
  account_id?: number
  group_id?: number
}
