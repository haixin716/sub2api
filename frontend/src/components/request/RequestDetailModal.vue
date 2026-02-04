<template>
  <BaseDialog
    :show="true"
    title="请求详情"
    width="full"
    @close="emit('close')"
  >
    <div class="space-y-6">
      <!-- Basic Info -->
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">请求 ID</label>
          <div class="mt-1 flex items-center gap-2">
            <code class="text-sm text-gray-900 dark:text-white">{{ request.request_id }}</code>
            <button
              @click="copyToClipboard(request.request_id)"
              class="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              复制
            </button>
          </div>
        </div>

        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">模型</label>
          <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ request.model }}</p>
        </div>

        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">请求方法 & 路径</label>
          <div class="mt-1 flex items-center gap-2">
            <span class="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-200">
              {{ request.request_method }}
            </span>
            <code class="text-xs text-gray-600 dark:text-gray-400">{{ request.request_path }}</code>
          </div>
        </div>

        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">响应状态</label>
          <div class="mt-1">
            <span
              class="inline-flex items-center rounded px-2 py-1 text-xs font-medium"
              :class="getStatusClass(request.response_status, request.is_error)"
            >
              {{ request.response_status }}
            </span>
          </div>
        </div>

        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">请求类型</label>
          <div class="mt-1 flex items-center gap-2">
            <Icon
              :name="request.stream ? 'bolt' : 'document'"
              size="sm"
              :class="request.stream ? 'text-purple-500' : 'text-gray-400'"
            />
            <span class="text-sm text-gray-900 dark:text-white">
              {{ request.stream ? '流式' : '非流式' }}
            </span>
          </div>
        </div>

        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">耗时</label>
          <p class="mt-1 text-sm text-gray-900 dark:text-white">
            {{ request.duration_ms ? formatDuration(request.duration_ms) : '-' }}
          </p>
        </div>

        <div v-if="isAdmin && 'ip_address' in request && request.ip_address">
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">IP 地址</label>
          <code class="mt-1 block text-sm text-gray-900 dark:text-white">{{ request.ip_address }}</code>
        </div>

        <div>
          <label class="text-xs font-medium text-gray-500 dark:text-gray-400">创建时间</label>
          <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatDateTime(request.created_at) }}</p>
        </div>
      </div>

      <!-- Error Message -->
      <div v-if="request.is_error && request.error_message" class="rounded-lg bg-red-50 p-4 dark:bg-red-900/20">
        <label class="text-xs font-medium text-red-800 dark:text-red-300">错误信息</label>
        <p class="mt-1 text-sm text-red-700 dark:text-red-400">{{ request.error_message }}</p>
        <p v-if="request.error_type" class="mt-1 text-xs text-red-600 dark:text-red-500">
          类型: {{ request.error_type }}
        </p>
      </div>

      <!-- User Agent -->
      <div v-if="request.user_agent">
        <label class="text-xs font-medium text-gray-500 dark:text-gray-400">User Agent</label>
        <p class="mt-1 text-xs text-gray-600 dark:text-gray-400 break-all">{{ request.user_agent }}</p>
      </div>

      <!-- Tabs for Request/Response -->
      <div>
        <div class="border-b border-gray-200 dark:border-gray-700">
          <nav class="-mb-px flex space-x-6">
            <button
              @click="activeTab = 'request'"
              :class="[
                'whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm',
                activeTab === 'request'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
              ]"
            >
              请求体
            </button>
            <button
              @click="activeTab = 'response'"
              :class="[
                'whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm',
                activeTab === 'response'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
              ]"
            >
              响应体
            </button>
          </nav>
        </div>

        <!-- Request Body -->
        <div v-show="activeTab === 'request'" class="mt-4">
          <div class="relative">
            <div class="absolute right-2 top-2 z-10">
              <button
                @click="copyToClipboard(request.request_body)"
                class="rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-gray-700"
              >
                复制
              </button>
            </div>
            <pre class="overflow-auto rounded-lg bg-gray-900 p-4 text-xs text-gray-100 dark:bg-gray-950"
                 style="max-height: 500px"><code>{{ formatJSON(request.request_body) }}</code></pre>
          </div>
        </div>

        <!-- Response Body -->
        <div v-show="activeTab === 'response'" class="mt-4">
          <div v-if="request.response_body" class="relative">
            <div class="absolute right-2 top-2 z-10">
              <button
                @click="copyToClipboard(request.response_body)"
                class="rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-gray-700"
              >
                复制
              </button>
            </div>
            <pre class="overflow-auto rounded-lg bg-gray-900 p-4 text-xs text-gray-100 dark:bg-gray-950"
                 style="max-height: 500px"><code>{{ formatJSON(request.response_body) }}</code></pre>
          </div>
          <div v-else class="rounded-lg bg-gray-50 p-8 text-center dark:bg-gray-800">
            <p class="text-sm text-gray-500 dark:text-gray-400">无响应内容</p>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { RequestLog, AdminRequestLog } from '@/types/request'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

interface Props {
  request: RequestLog | AdminRequestLog
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'close'): void
}>()

const activeTab = ref<'request' | 'response'>('request')

const isAdmin = computed(() => {
  return 'ip_address' in props.request
})

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

function formatJSON(jsonStr: string): string {
  try {
    const parsed = JSON.parse(jsonStr)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return jsonStr
  }
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    // You might want to show a toast notification here
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}
</script>
