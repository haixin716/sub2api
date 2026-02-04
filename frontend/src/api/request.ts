// 用户端请求日志 API

import { apiClient } from './client'
import type { RequestLog, RequestLogQueryParams } from '@/types/request'
import type { PaginatedResponse } from '@/types'

/**
 * 获取请求日志列表（分页）
 */
export async function list(
  params: RequestLogQueryParams = {},
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<RequestLog>> {
  const { data } = await apiClient.get<PaginatedResponse<RequestLog>>('/requests', {
    params,
    signal: options?.signal,
  })
  return data
}

/**
 * 获取单条请求日志详情
 */
export async function getById(id: number): Promise<RequestLog> {
  const { data } = await apiClient.get<RequestLog>(`/requests/${id}`)
  return data
}

/**
 * 导出为 CSV（前端实现）
 */
export function exportToCSV(logs: RequestLog[]): void {
  const headers = [
    'ID',
    'Request ID',
    'Model',
    'Method',
    'Path',
    'Status',
    'Stream',
    'Is Error',
    'Duration (ms)',
    'Created At',
  ]

  const rows = logs.map(log => [
    log.id,
    log.request_id,
    log.model,
    log.request_method,
    log.request_path,
    log.response_status,
    log.stream ? 'Yes' : 'No',
    log.is_error ? 'Yes' : 'No',
    log.duration_ms || '',
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
  link.setAttribute('download', `request_logs_${Date.now()}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
