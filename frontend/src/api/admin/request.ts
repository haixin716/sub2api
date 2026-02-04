// 管理员端请求日志 API

import { apiClient } from '../client'
import type { AdminRequestLog, AdminRequestLogQueryParams } from '@/types/request'
import type { PaginatedResponse } from '@/types'

/**
 * 获取所有请求日志列表（分页）
 */
export async function list(
  params: AdminRequestLogQueryParams = {},
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<AdminRequestLog>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminRequestLog>>(
    '/admin/requests',
    {
      params,
      signal: options?.signal,
    }
  )
  return data
}

/**
 * 获取单条请求日志详情
 */
export async function getById(id: number): Promise<AdminRequestLog> {
  const { data } = await apiClient.get<AdminRequestLog>(`/admin/requests/${id}`)
  return data
}

/**
 * 搜索用户
 */
export async function searchUsers(keyword: string): Promise<Array<{ id: number; email: string }>> {
  const { data } = await apiClient.get<Array<{ id: number; email: string }>>(
    '/admin/requests/search-users',
    {
      params: { keyword },
    }
  )
  return data
}

/**
 * 搜索 API Keys
 */
export async function searchAPIKeys(
  params: { user_id?: number; keyword?: string }
): Promise<Array<{ id: number; name: string; key: string }>> {
  const { data } = await apiClient.get<Array<{ id: number; name: string; key: string }>>(
    '/admin/requests/search-api-keys',
    {
      params,
    }
  )
  return data
}

/**
 * 导出为 CSV（前端实现）
 */
export function exportToCSV(logs: AdminRequestLog[]): void {
  const headers = [
    'ID',
    'User ID',
    'User Email',
    'API Key ID',
    'API Key Name',
    'Account ID',
    'Account Name',
    'Request ID',
    'Model',
    'Method',
    'Path',
    'Status',
    'Stream',
    'Is Error',
    'Error Message',
    'Duration (ms)',
    'IP Address',
    'User Agent',
    'Created At',
  ]

  const rows = logs.map(log => [
    log.id,
    log.user_id,
    log.user?.email || '',
    log.api_key_id,
    log.api_key?.name || '',
    log.account_id,
    log.account?.name || '',
    log.request_id,
    log.model,
    log.request_method,
    log.request_path,
    log.response_status,
    log.stream ? 'Yes' : 'No',
    log.is_error ? 'Yes' : 'No',
    log.error_message || '',
    log.duration_ms || '',
    log.ip_address || '',
    log.user_agent || '',
    log.created_at,
  ])

  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.map(cell => `"${cell}"`).join(',')),
  ].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', `admin_request_logs_${Date.now()}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
