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
              <!-- User Search -->
              <div ref="userSearchRef" class="usage-filter-dropdown relative min-w-[220px]">
                <label class="input-label">用户</label>
                <input
                  v-model="userKeyword"
                  type="text"
                  class="input pr-8"
                  placeholder="搜索用户邮箱"
                  @input="debounceUserSearch"
                  @focus="showUserDropdown = true"
                />
                <button
                  v-if="filters.user_id"
                  type="button"
                  @click="clearUser"
                  class="absolute right-2 top-9 text-gray-400"
                  aria-label="Clear user filter"
                >
                  ✕
                </button>
                <div
                  v-if="showUserDropdown && (userResults.length > 0 || userKeyword)"
                  class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border bg-white shadow-lg dark:bg-gray-800"
                >
                  <button
                    v-for="u in userResults"
                    :key="u.id"
                    type="button"
                    @click="selectUser(u)"
                    class="w-full px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    <span>{{ u.email }}</span>
                    <span class="ml-2 text-xs text-gray-400">#{{ u.id }}</span>
                  </button>
                </div>
              </div>

              <!-- API Key Search -->
              <div ref="apiKeySearchRef" class="usage-filter-dropdown relative min-w-[220px]">
                <label class="input-label">API 密钥</label>
                <input
                  v-model="apiKeyKeyword"
                  type="text"
                  class="input pr-8"
                  placeholder="搜索 API 密钥名称"
                  @input="debounceApiKeySearch"
                  @focus="onApiKeyFocus"
                />
                <button
                  v-if="filters.api_key_id"
                  type="button"
                  @click="onClearApiKey"
                  class="absolute right-2 top-9 text-gray-400"
                  aria-label="Clear API key filter"
                >
                  ✕
                </button>
                <div
                  v-if="showApiKeyDropdown && apiKeyResults.length > 0"
                  class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border bg-white shadow-lg dark:bg-gray-800"
                >
                  <button
                    v-for="k in apiKeyResults"
                    :key="k.id"
                    type="button"
                    @click="selectApiKey(k)"
                    class="w-full px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    <span class="truncate">{{ k.name || `#${k.id}` }}</span>
                    <span class="ml-2 text-xs text-gray-400">#{{ k.id }}</span>
                  </button>
                </div>
              </div>

              <!-- Model Filter -->
              <div class="min-w-[200px]">
                <label class="input-label">模型</label>
                <Select v-model="filters.model" :options="modelOptions" searchable @change="applyFilters" />
              </div>

              <!-- Stream Filter -->
              <div class="min-w-[120px]">
                <label class="input-label">请求类型</label>
                <Select
                  v-model="filters.stream"
                  :options="streamOptions"
                  @change="applyFilters"
                />
              </div>

              <!-- Error Filter -->
              <div class="min-w-[120px]">
                <label class="input-label">状态</label>
                <Select
                  v-model="filters.is_error"
                  :options="errorOptions"
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
                <button @click="showCleanupDialog = true" class="btn btn-danger">
                  清理
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

          <template #cell-user="{ row }">
            <div v-if="row.user" class="text-sm">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.user.email }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ row.user_id }}</div>
            </div>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>

          <template #cell-api_key="{ row }">
            <div v-if="row.api_key" class="text-sm">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.api_key.name }}</div>
              <code class="text-xs text-gray-500 dark:text-gray-400">{{ row.api_key.key.substring(0, 20) }}...</code>
            </div>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>

          <template #cell-account="{ row }">
            <div v-if="row.account" class="text-sm">
              <span class="font-medium text-gray-900 dark:text-white">{{ row.account.name }}</span>
            </div>
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

          <template #cell-ip_address="{ row }">
            <code class="text-xs text-gray-600 dark:text-gray-400">{{ row.ip_address || '-' }}</code>
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

    <!-- Cleanup Dialog -->
    <RequestCleanupDialog
      :show="showCleanupDialog"
      :start-date="startDate"
      :end-date="endDate"
      @close="showCleanupDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { list, exportToCSV, searchUsers, searchAPIKeys } from '@/api/admin/request'
import { adminAPI } from '@/api/admin'
import type { AdminRequestLog } from '@/types/request'
import type { SelectOption } from '@/components/common/Select.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import RequestDetailModal from '@/components/request/RequestDetailModal.vue'
import RequestCleanupDialog from '@/components/admin/request/RequestCleanupDialog.vue'

// State
const requests = ref<AdminRequestLog[]>([])
const loading = ref(false)
const exporting = ref(false)
const selectedRequest = ref<AdminRequestLog | null>(null)
const showCleanupDialog = ref(false)

// Filters
const filters = ref({
  user_id: undefined as number | undefined,
  api_key_id: undefined as number | undefined,
  model: '' as string | null,
  stream: null as boolean | null,
  is_error: null as boolean | null,
})

// User search state
const userSearchRef = ref<HTMLElement | null>(null)
const userKeyword = ref('')
const userResults = ref<Array<{ id: number; email: string }>>([])
const showUserDropdown = ref(false)
let userSearchTimeout: ReturnType<typeof setTimeout> | null = null

// API Key search state
const apiKeySearchRef = ref<HTMLElement | null>(null)
const apiKeyKeyword = ref('')
const apiKeyResults = ref<Array<{ id: number; name: string; key: string }>>([])
const showApiKeyDropdown = ref(false)
let apiKeySearchTimeout: ReturnType<typeof setTimeout> | null = null

// Model options
const modelOptions = ref<SelectOption[]>([{ value: null, label: '全部模型' }])

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

// Options
const streamOptions = [
  { label: '全部类型', value: null },
  { label: '流式', value: true },
  { label: '同步', value: false },
]
const errorOptions = [
  { label: '全部状态', value: null },
  { label: '成功', value: false },
  { label: '错误', value: true },
]

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
  { key: 'user', label: '用户', width: '180px' },
  { key: 'api_key', label: 'API 密钥', width: '180px' },
  { key: 'account', label: '账号', width: '120px' },
  { key: 'model', label: '模型', width: '180px' },
  { key: 'method_path', label: '方法 & 路径', width: '250px' },
  { key: 'status', label: '状态码', width: '80px' },
  { key: 'stream', label: '类型', width: '60px' },
  { key: 'duration_ms', label: '耗时', width: '100px' },
  { key: 'ip_address', label: 'IP 地址', width: '140px' },
  { key: 'created_at', label: '创建时间', width: '160px' },
  { key: 'actions', label: '操作', width: '100px' },
]

// Methods
async function loadRequests() {
  loading.value = true
  try {
    const { model, stream, is_error, ...restFilters } = filters.value
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      ...restFilters,
      model: model || undefined,
      stream: stream ?? undefined,
      is_error: is_error ?? undefined,
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
    user_id: undefined,
    api_key_id: undefined,
    model: '',
    stream: null,
    is_error: null,
  }
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
  const range = getDefaultDateRange()
  startDate.value = range.start
  endDate.value = range.end
  applyFilters()
}

// User search
const debounceUserSearch = () => {
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  userSearchTimeout = setTimeout(async () => {
    if (!userKeyword.value) {
      userResults.value = []
      return
    }
    try {
      userResults.value = await searchUsers(userKeyword.value)
    } catch {
      userResults.value = []
    }
  }, 300)
}

const selectUser = async (u: { id: number; email: string }) => {
  userKeyword.value = u.email
  showUserDropdown.value = false
  filters.value.user_id = u.id
  clearApiKey()
  // Auto-load API keys for this user
  try {
    apiKeyResults.value = await searchAPIKeys({ user_id: u.id })
  } catch {
    apiKeyResults.value = []
  }
  applyFilters()
}

const clearUser = () => {
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  filters.value.user_id = undefined
  clearApiKey()
  applyFilters()
}

// API Key search
const debounceApiKeySearch = () => {
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  apiKeySearchTimeout = setTimeout(async () => {
    try {
      apiKeyResults.value = await searchAPIKeys({
        user_id: filters.value.user_id,
        keyword: apiKeyKeyword.value || ''
      })
    } catch {
      apiKeyResults.value = []
    }
  }, 300)
}

const selectApiKey = (k: { id: number; name: string; key: string }) => {
  apiKeyKeyword.value = k.name || String(k.id)
  showApiKeyDropdown.value = false
  filters.value.api_key_id = k.id
  applyFilters()
}

const clearApiKey = () => {
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
  filters.value.api_key_id = undefined
}

const onClearApiKey = () => {
  clearApiKey()
  applyFilters()
}

const onApiKeyFocus = () => {
  showApiKeyDropdown.value = true
  if (apiKeyResults.value.length === 0) {
    debounceApiKeySearch()
  }
}

// Click outside handler
const onDocumentClick = (e: MouseEvent) => {
  const target = e.target as Node | null
  if (!target) return
  if (!(userSearchRef.value?.contains(target) ?? false)) showUserDropdown.value = false
  if (!(apiKeySearchRef.value?.contains(target) ?? false)) showApiKeyDropdown.value = false
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

function viewDetails(request: AdminRequestLog) {
  selectedRequest.value = request
}

async function handleExport() {
  exporting.value = true
  try {
    // Load all data for export
    const { model, stream, is_error, ...restFilters } = filters.value
    const params = {
      page: 1,
      page_size: 1000,
      ...restFilters,
      model: model || undefined,
      stream: stream ?? undefined,
      is_error: is_error ?? undefined,
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
  document.addEventListener('click', onDocumentClick)
  loadRequests()
  // Load model options
  try {
    const ms = await adminAPI.dashboard.getModelStats({ start_date: startDate.value, end_date: endDate.value })
    const uniqueModels = new Set<string>()
    ms.models?.forEach((s: any) => s.model && uniqueModels.add(s.model))
    modelOptions.value.push(
      ...Array.from(uniqueModels)
        .sort()
        .map((m) => ({ value: m, label: m }))
    )
  } catch {
    // Ignore model options loading errors
  }
})

onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
})
</script>
