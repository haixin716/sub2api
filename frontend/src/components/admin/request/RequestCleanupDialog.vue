<template>
  <BaseDialog :show="show" title="请求记录清理" width="wide" @close="handleClose">
    <div class="space-y-4">
      <!-- Filters -->
      <div class="flex flex-wrap items-end gap-4">
        <!-- User Search -->
        <div ref="userSearchRef" class="usage-filter-dropdown relative min-w-[200px]">
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
            v-if="localFilters.user_id"
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
        <div ref="apiKeySearchRef" class="usage-filter-dropdown relative min-w-[200px]">
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
            v-if="localFilters.api_key_id"
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
        <div class="min-w-[180px]">
          <label class="input-label">模型</label>
          <Select v-model="localFilters.model" :options="modelOptions" searchable />
        </div>

        <!-- Stream Filter -->
        <div class="min-w-[120px]">
          <label class="input-label">请求类型</label>
          <Select
            v-model="localFilters.stream"
            :options="streamOptions"
          />
        </div>

        <!-- Error Filter -->
        <div class="min-w-[120px]">
          <label class="input-label">状态</label>
          <Select
            v-model="localFilters.is_error"
            :options="errorOptions"
          />
        </div>

        <!-- Date Range Filter -->
        <div>
          <label class="input-label">时间范围</label>
          <DateRangePicker
            v-model:start-date="localStartDate"
            v-model:end-date="localEndDate"
          />
        </div>
      </div>

      <!-- Warning -->
      <div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200">
        ⚠️ 清理操作将永久删除符合条件的请求记录，此操作不可撤销。请确认时间范围和过滤条件。
      </div>

      <!-- Recent Tasks -->
      <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
        <div class="flex items-center justify-between">
          <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-200">
            最近的清理任务
          </h4>
          <button type="button" class="btn btn-ghost btn-sm" @click="loadTasks">
            刷新
          </button>
        </div>

        <div class="mt-3 space-y-2">
          <div v-if="tasksLoading" class="text-sm text-gray-500 dark:text-gray-400">
            加载中...
          </div>
          <div v-else-if="tasks.length === 0" class="text-sm text-gray-500 dark:text-gray-400">
            暂无清理任务
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="task in tasks"
              :key="task.id"
              class="flex flex-col gap-2 rounded-lg border border-gray-100 px-3 py-2 text-sm text-gray-600 dark:border-dark-700 dark:text-gray-300"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span :class="statusClass(task.status)" class="rounded-full px-2 py-0.5 text-xs font-semibold">
                    {{ statusLabel(task.status) }}
                  </span>
                  <span class="text-xs text-gray-400">#{{ task.id }}</span>
                  <button
                    v-if="canCancel(task)"
                    type="button"
                    class="btn btn-ghost btn-xs text-rose-600 hover:text-rose-700 dark:text-rose-300"
                    @click="openCancelConfirm(task)"
                  >
                    取消
                  </button>
                </div>
                <div class="text-xs text-gray-400">
                  {{ formatDateTime(task.created_at) }}
                </div>
              </div>
              <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                <span>范围: {{ formatRange(task) }}</span>
                <span>已删除: {{ task.deleted_rows.toLocaleString() }}</span>
              </div>
              <div v-if="task.error_message" class="text-xs text-rose-500">
                {{ task.error_message }}
              </div>
            </div>
          </div>
        </div>

        <Pagination
          v-if="tasksTotal > tasksPageSize"
          class="mt-4"
          :total="tasksTotal"
          :page="tasksPage"
          :page-size="tasksPageSize"
          :page-size-options="[5]"
          :show-page-size-selector="false"
          :show-jump="true"
          @update:page="handleTaskPageChange"
          @update:pageSize="handleTaskPageSizeChange"
        />
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          关闭
        </button>
        <button type="button" class="btn btn-danger" :disabled="submitting" @click="openConfirm">
          {{ submitting ? '提交中...' : '创建清理任务' }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <ConfirmDialog
    :show="confirmVisible"
    title="确认创建清理任务"
    message="此操作将永久删除符合条件的请求记录，不可恢复。确定要继续吗？"
    confirm-text="确认清理"
    danger
    @confirm="submitCleanup"
    @cancel="confirmVisible = false"
  />

  <ConfirmDialog
    :show="cancelConfirmVisible"
    title="确认取消任务"
    message="确定要取消此清理任务吗？"
    confirm-text="确认取消"
    danger
    @confirm="cancelTask"
    @cancel="cancelConfirmVisible = false"
  />
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import {
  listCleanupTasks,
  createCleanupTask,
  cancelCleanupTask,
  searchUsers,
  searchAPIKeys,
} from '@/api/admin/request'
import type {
  RequestLogCleanupTask,
  CreateRequestLogCleanupTaskRequest,
} from '@/api/admin/request'

interface Props {
  show: boolean
  startDate: string
  endDate: string
}

const props = defineProps<Props>()
const emit = defineEmits(['close'])

const appStore = useAppStore()

const localFilters = ref<{
  user_id?: number
  api_key_id?: number
  model?: string | null
  stream?: boolean
  is_error?: boolean
}>({})
const localStartDate = ref('')
const localEndDate = ref('')

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

// Options (same as the list page)
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

const tasks = ref<RequestLogCleanupTask[]>([])
const tasksLoading = ref(false)
const tasksPage = ref(1)
const tasksPageSize = ref(5)
const tasksTotal = ref(0)
const submitting = ref(false)
const confirmVisible = ref(false)
const cancelConfirmVisible = ref(false)
const canceling = ref(false)
const cancelTarget = ref<RequestLogCleanupTask | null>(null)
let pollTimer: number | null = null

const resetFilters = () => {
  localFilters.value = {}
  localStartDate.value = props.startDate
  localEndDate.value = props.endDate
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
  tasksPage.value = 1
  tasksTotal.value = 0
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
  localFilters.value.user_id = u.id
  clearApiKey()
  try {
    apiKeyResults.value = await searchAPIKeys({ user_id: u.id })
  } catch {
    apiKeyResults.value = []
  }
}

const clearUser = () => {
  userKeyword.value = ''
  userResults.value = []
  showUserDropdown.value = false
  localFilters.value.user_id = undefined
  clearApiKey()
}

// API Key search
const debounceApiKeySearch = () => {
  if (apiKeySearchTimeout) clearTimeout(apiKeySearchTimeout)
  apiKeySearchTimeout = setTimeout(async () => {
    try {
      apiKeyResults.value = await searchAPIKeys({
        user_id: localFilters.value.user_id,
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
  localFilters.value.api_key_id = k.id
}

const clearApiKey = () => {
  apiKeyKeyword.value = ''
  apiKeyResults.value = []
  showApiKeyDropdown.value = false
  localFilters.value.api_key_id = undefined
}

const onClearApiKey = () => {
  clearApiKey()
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

const startPolling = () => {
  stopPolling()
  pollTimer = window.setInterval(() => {
    loadTasks()
  }, 10000)
}

const stopPolling = () => {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

const handleClose = () => {
  stopPolling()
  confirmVisible.value = false
  cancelConfirmVisible.value = false
  canceling.value = false
  cancelTarget.value = null
  submitting.value = false
  emit('close')
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending: '等待中',
    running: '执行中',
    succeeded: '已完成',
    failed: '失败',
    canceled: '已取消'
  }
  return map[status] || status
}

const statusClass = (status: string) => {
  const map: Record<string, string> = {
    pending: 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-200',
    running: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-200',
    succeeded: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-200',
    failed: 'bg-rose-100 text-rose-700 dark:bg-rose-500/20 dark:text-rose-200',
    canceled: 'bg-gray-200 text-gray-600 dark:bg-dark-600 dark:text-gray-300'
  }
  return map[status] || 'bg-gray-100 text-gray-600'
}

const formatDateTime = (value?: string | null) => {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

const formatRange = (task: RequestLogCleanupTask) => {
  const start = formatDateTime(task.filters.start_time)
  const end = formatDateTime(task.filters.end_time)
  return `${start} ~ ${end}`
}

const getUserTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

const loadTasks = async () => {
  if (!props.show) return
  tasksLoading.value = true
  try {
    const res = await listCleanupTasks({
      page: tasksPage.value,
      page_size: tasksPageSize.value
    })
    tasks.value = res.items || []
    tasksTotal.value = res.total || 0
    if (res.page) {
      tasksPage.value = res.page
    }
    if (res.page_size) {
      tasksPageSize.value = res.page_size
    }
  } catch (error) {
    console.error('Failed to load cleanup tasks:', error)
    appStore.showError('加载清理任务失败')
  } finally {
    tasksLoading.value = false
  }
}

const handleTaskPageChange = (page: number) => {
  tasksPage.value = page
  loadTasks()
}

const handleTaskPageSizeChange = (size: number) => {
  if (!Number.isFinite(size) || size <= 0) return
  tasksPageSize.value = size
  tasksPage.value = 1
  loadTasks()
}

const openConfirm = () => {
  confirmVisible.value = true
}

const canCancel = (task: RequestLogCleanupTask) => {
  return task.status === 'pending' || task.status === 'running'
}

const openCancelConfirm = (task: RequestLogCleanupTask) => {
  cancelTarget.value = task
  cancelConfirmVisible.value = true
}

const buildPayload = (): CreateRequestLogCleanupTaskRequest | null => {
  if (!localStartDate.value || !localEndDate.value) {
    appStore.showError('请选择开始和结束日期')
    return null
  }

  const payload: CreateRequestLogCleanupTaskRequest = {
    start_date: localStartDate.value,
    end_date: localEndDate.value,
    timezone: getUserTimezone()
  }

  if (localFilters.value.user_id && localFilters.value.user_id > 0) {
    payload.user_id = localFilters.value.user_id
  }
  if (localFilters.value.api_key_id && localFilters.value.api_key_id > 0) {
    payload.api_key_id = localFilters.value.api_key_id
  }
  if (localFilters.value.model) {
    payload.model = localFilters.value.model
  }
  if (localFilters.value.stream !== null && localFilters.value.stream !== undefined) {
    payload.stream = localFilters.value.stream
  }
  if (localFilters.value.is_error !== null && localFilters.value.is_error !== undefined) {
    payload.is_error = localFilters.value.is_error
  }

  return payload
}

const submitCleanup = async () => {
  const payload = buildPayload()
  if (!payload) {
    confirmVisible.value = false
    return
  }
  submitting.value = true
  confirmVisible.value = false
  try {
    await createCleanupTask(payload)
    appStore.showSuccess('清理任务已创建')
    loadTasks()
  } catch (error) {
    console.error('Failed to create cleanup task:', error)
    appStore.showError('创建清理任务失败')
  } finally {
    submitting.value = false
  }
}

const cancelTask = async () => {
  const task = cancelTarget.value
  if (!task) {
    cancelConfirmVisible.value = false
    return
  }
  canceling.value = true
  cancelConfirmVisible.value = false
  try {
    await cancelCleanupTask(task.id)
    appStore.showSuccess('清理任务已取消')
    loadTasks()
  } catch (error) {
    console.error('Failed to cancel cleanup task:', error)
    appStore.showError('取消清理任务失败')
  } finally {
    canceling.value = false
    cancelTarget.value = null
  }
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      resetFilters()
      loadTasks()
      startPolling()
      document.addEventListener('click', onDocumentClick)
      // Load model options
      adminAPI.dashboard.getModelStats({ start_date: props.startDate, end_date: props.endDate }).then((ms) => {
        const uniqueModels = new Set<string>()
        ms.models?.forEach((s: any) => s.model && uniqueModels.add(s.model))
        modelOptions.value = [
          { value: null, label: '全部模型' },
          ...Array.from(uniqueModels).sort().map((m) => ({ value: m, label: m }))
        ]
      }).catch(() => {})
    } else {
      stopPolling()
      document.removeEventListener('click', onDocumentClick)
    }
  }
)

onUnmounted(() => {
  stopPolling()
  document.removeEventListener('click', onDocumentClick)
})
</script>
