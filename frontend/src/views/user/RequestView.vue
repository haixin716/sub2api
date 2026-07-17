<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Statistics Cards -->
      <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <!-- Total Requests -->
        <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
                <Icon name="document" size="md" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  总请求数
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ totalRequests.toLocaleString() }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  所选时间范围内
                </p>
              </div>
            </div>
          </div>

          <!-- Error Requests -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-red-100 p-2 dark:bg-red-900/30">
                <Icon name="exclamationCircle" size="md" class="text-red-600 dark:text-red-400" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  错误请求
                </p>
                <p class="text-xl font-bold text-red-600 dark:text-red-400">
                  {{ errorRequests.toLocaleString() }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ errorRate }}% 错误率
                </p>
              </div>
            </div>
          </div>

          <!-- Stream Requests -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
                <Icon name="bolt" size="md" class="text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  流式请求
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ streamRequests.toLocaleString() }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ streamRate }}% 流式率
                </p>
              </div>
            </div>
          </div>

          <!-- Average Duration -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
                <Icon name="clock" size="md" class="text-amber-600 dark:text-amber-400" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  平均耗时
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatDuration(avgDuration) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">每个请求</p>
              </div>
            </div>
          </div>
        </div>

      <!-- Filters Card -->
      <div class="card">
        <div class="px-6 py-4">
          <div class="flex flex-wrap items-end gap-4">
              <!-- API Key Filter -->
              <div class="min-w-[180px]">
                <label class="input-label">API 密钥</label>
                <Select
                  v-model="filters.api_key_id"
                  :options="apiKeyOptions"
                  placeholder="所有 API 密钥"
                  @change="applyFilters"
                />
              </div>

              <!-- Date Range Filter -->
              <div>
                <label class="input-label">时间范围</label>
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>

              <!-- Actions -->
              <div class="ml-auto flex items-center gap-3">
                <button @click="resetFilters" class="btn btn-secondary">
                  重置
                </button>
                <button @click="handleExport" :disabled="exporting" class="btn btn-primary">
                  <Icon v-if="exporting" name="refresh" size="sm" class="mr-2 animate-spin" />
                  {{ exporting ? '导出中...' : '导出 CSV' }}
                </button>
              </div>
            </div>
          </div>
        </div>

      <!-- Table Card -->
      <div class="card overflow-hidden">
        <div class="overflow-auto">
          <DataTable :columns="columns" :data="requests" :loading="loading">
          <template #cell-request_id="{ row }">
            <code class="text-xs text-gray-600 dark:text-gray-400">{{ row.client_request_id.substring(0, 12) }}...</code>
          </template>

          <template #cell-api_key="{ row }">
            <span v-if="row.api_key" class="text-sm text-gray-900 dark:text-white">{{ row.api_key.name }}</span>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>

          <template #cell-model="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-method_path="{ row }">
            <div class="flex flex-col gap-1">
              <span class="inline-flex items-center gap-1">
                <span class="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                  {{ row.request_method }}
                </span>
                <code class="text-xs text-gray-600 dark:text-gray-400">{{ row.request_path }}</code>
              </span>
            </div>
          </template>

          <template #cell-status="{ row }">
            <span
              class="inline-flex items-center rounded px-2 py-1 text-xs font-medium"
              :class="getStatusClass(row.response_status, row.is_error)"
            >
              {{ row.response_status }}
            </span>
          </template>

          <template #cell-stream="{ row }">
            <span class="inline-flex items-center gap-1">
              <Icon
                :name="row.stream ? 'bolt' : 'document'"
                size="sm"
                :class="row.stream ? 'text-purple-500' : 'text-gray-400'"
              />
              <span :class="row.stream ? 'text-purple-600 dark:text-purple-400' : 'text-gray-500 dark:text-gray-400'" class="text-xs">
                {{ row.stream ? '流式' : '普通' }}
              </span>
            </span>
          </template>

          <template #cell-duration_ms="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">
              {{ row.duration_ms ? formatDuration(row.duration_ms) : '-' }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-400">
              {{ formatDateTime(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <button
              @click="viewDetails(row)"
              class="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
            >
              查看详情
            </button>
          </template>
        </DataTable>
        </div>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
      />
    </div>

    <!-- Detail Modal -->
    <RequestDetailModal
      v-if="selectedRequest"
      :request="selectedRequest"
      @close="selectedRequest = null"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { list, exportToCSV } from '@/api/request'
import { keysAPI } from '@/api'
import type { RequestLog } from '@/types/request'
import type { ApiKey } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import RequestDetailModal from '@/components/request/RequestDetailModal.vue'

// State
const requests = ref<RequestLog[]>([])
const loading = ref(false)
const exporting = ref(false)
const selectedRequest = ref<RequestLog | null>(null)

// Filters
const filters = ref({
  api_key_id: undefined as number | undefined,
})

// API Keys
const apiKeys = ref<ApiKey[]>([])
const apiKeyOptions = computed(() => [
  { value: null, label: '所有 API 密钥' },
  ...apiKeys.value.map(key => ({ value: key.id, label: key.name }))
])

// Initialize date range to last 7 days
const getDefaultDateRange = () => {
  const now = new Date()
  const end = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
  const startD = new Date()
  startD.setDate(startD.getDate() - 6)
  const start = `${startD.getFullYear()}-${String(startD.getMonth() + 1).padStart(2, '0')}-${String(startD.getDate()).padStart(2, '0')}`
  return { start, end }
}
const defaultRange = getDefaultDateRange()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

// Pagination
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0,
})

// Options removed (model, stream, error filters no longer needed)

// Computed
const totalRequests = computed(() => pagination.value.total)
const errorRequests = computed(() => requests.value.filter(r => r.is_error).length)
const streamRequests = computed(() => requests.value.filter(r => r.stream).length)
const errorRate = computed(() =>
  totalRequests.value > 0 ? ((errorRequests.value / totalRequests.value) * 100).toFixed(1) : '0.0'
)
const streamRate = computed(() =>
  totalRequests.value > 0 ? ((streamRequests.value / totalRequests.value) * 100).toFixed(1) : '0.0'
)
const avgDuration = computed(() => {
  const durations = requests.value.filter(r => r.duration_ms).map(r => r.duration_ms!)
  return durations.length > 0 ? durations.reduce((a, b) => a + b, 0) / durations.length : 0
})

// Table columns
const columns = [
  { key: 'request_id', label: '请求 ID', width: '120px' },
  { key: 'api_key', label: 'API 密钥', width: '150px' },
  { key: 'model', label: '模型', width: '180px' },
  { key: 'method_path', label: '方法 & 路径', width: '250px' },
  { key: 'status', label: '状态码', width: '80px' },
  { key: 'stream', label: '类型', width: '60px' },
  { key: 'duration_ms', label: '耗时', width: '100px' },
  { key: 'created_at', label: '创建时间', width: '160px' },
  { key: 'actions', label: '操作', width: '100px' },
]

// Methods
async function loadRequests() {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      ...filters.value,
      start_date: startDate.value,
      end_date: endDate.value,
    }

    const response = await list(params)
    requests.value = response.items
    pagination.value.total = response.total
    pagination.value.pages = response.pages
  } catch (error) {
    console.error('Failed to load requests:', error)
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.value.page = 1
  loadRequests()
}

function resetFilters() {
  filters.value = {
    api_key_id: undefined,
  }
  const range = getDefaultDateRange()
  startDate.value = range.start
  endDate.value = range.end
  applyFilters()
}

function onDateRangeChange() {
  applyFilters()
}

function onPageChange(page: number) {
  pagination.value.page = page
  loadRequests()
}

function onPageSizeChange(pageSize: number) {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  loadRequests()
}

function viewDetails(request: RequestLog) {
  selectedRequest.value = request
}

async function handleExport() {
  exporting.value = true
  try {
    // Load all data for export
    const params = {
      page: 1,
      page_size: 1000,
      ...filters.value,
      start_date: startDate.value,
      end_date: endDate.value,
    }
    const response = await list(params)
    exportToCSV(response.items)
  } catch (error) {
    console.error('Export failed:', error)
  } finally {
    exporting.value = false
  }
}

function getStatusClass(status: number, isError: boolean) {
  if (isError || status >= 400) {
    return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  }
  if (status >= 300) {
    return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
  }
  return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

// Lifecycle
onMounted(async () => {
  loadRequests()
  // Load API key options
  try {
    const response = await keysAPI.list(1, 100)
    apiKeys.value = response.items
  } catch (error) {
    console.error('Failed to load API keys:', error)
  }
})
</script>
