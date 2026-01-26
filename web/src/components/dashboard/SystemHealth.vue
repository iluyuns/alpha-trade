<template>
  <div class="system-health">
    <div
      v-for="item in health"
      :key="item.name"
      class="system-health__item"
    >
      <div class="system-health__label">{{ item.name }}</div>
      <div class="system-health__value">
        <span :class="['system-health__status', `system-health__status--${item.status}`]">
          {{ getStatusText(item.status) }}
        </span>
        <span v-if="item.latency" class="system-health__latency">
          ({{ item.latency }}ms)
        </span>
        <div v-if="item.message" class="system-health__message">
          {{ item.message }}
        </div>
        <div v-if="item.lastHeartbeat" class="system-health__heartbeat">
          心跳: {{ formatTime(item.lastHeartbeat) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
import type { SystemHealthItem } from '@/stores/dashboard'

const props = defineProps<{
  health: SystemHealthItem[]
}>()

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    normal: '正常',
    warning: '警告',
    error: '错误',
  }
  return map[status] || status
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000)
  
  if (diff < 60) {
    return `${diff}秒前`
  } else if (diff < 3600) {
    return `${Math.floor(diff / 60)}分钟前`
  } else {
    return `${Math.floor(diff / 3600)}小时前`
  }
}
</script>

<style scoped>
.system-health {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.system-health__item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-sm) 0;
  border-bottom: 1px solid var(--color-border-lighter);
}

.system-health__item:last-child {
  border-bottom: none;
}

.system-health__label {
  font-weight: 500;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  flex-shrink: 0;
  min-width: 120px;
}

.system-health__value {
  flex: 1;
  text-align: right;
}

.system-health__status {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--border-radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.system-health__status--normal {
  background-color: rgba(103, 194, 58, 0.1);
  color: var(--color-success);
}

.system-health__status--warning {
  background-color: rgba(230, 162, 60, 0.1);
  color: var(--color-warning);
}

.system-health__status--error {
  background-color: rgba(245, 108, 108, 0.1);
  color: var(--color-danger);
}

.system-health__latency {
  margin-left: var(--spacing-xs);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

.system-health__message {
  margin-top: var(--spacing-xs);
  color: var(--color-text-regular);
  font-size: var(--font-size-xs);
}

.system-health__heartbeat {
  margin-top: var(--spacing-xs);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

@media (max-width: 767px) {
  .system-health__item {
    flex-direction: column;
    gap: var(--spacing-xs);
  }
  
  .system-health__value {
    text-align: left;
  }
  
  .system-health__label {
    min-width: auto;
  }
}
</style>
