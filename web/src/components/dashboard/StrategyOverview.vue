<template>
  <div class="strategy-overview">
    <Table :data="strategies" :columns="columns">
      <template #cell-direction="{ row }">
        <span :class="['strategy-overview__tag', `strategy-overview__tag--${getDirectionClass(row.direction)}`]">
          {{ row.direction }}
        </span>
      </template>
      <template #cell-status="{ row }">
        <span :class="['strategy-overview__tag', `strategy-overview__tag--${getStatusClass(row.status)}`]">
          {{ getStatusText(row.status) }}
        </span>
      </template>
      <template #cell-actions="{ row }">
        <div class="strategy-overview__actions">
          <Button
            v-if="row.status === 'running'"
            type="warning"
            size="small"
            @click="handlePause(row)"
          >
            暂停
          </Button>
          <Button
            v-else-if="row.status === 'stopped'"
            type="success"
            size="small"
            @click="handleStart(row)"
          >
            启动
          </Button>
          <Button
            v-if="row.reason"
            type="info"
            size="small"
            @click="showReason(row)"
          >
            查看原因
          </Button>
        </div>
      </template>
    </Table>
    
    <Dialog v-model="showReasonDialog" title="状态原因" @close="showReasonDialog = false">
      <p>{{ selectedReason }}</p>
      <template #footer>
        <Button @click="showReasonDialog = false" class="ui-dialog__single-button">确定</Button>
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { StrategyOverview } from '@/stores/dashboard'
import Table from '@/components/ui/Table.vue'
import Button from '@/components/ui/Button.vue'
import Dialog from '@/components/ui/Dialog.vue'

const props = defineProps<{
  strategies: StrategyOverview[]
}>()

const showReasonDialog = ref(false)
const selectedReason = ref('')

const columns = [
  { prop: 'id', label: 'ID' },
  { prop: 'name', label: '策略名' },
  { prop: 'symbol', label: '标的' },
  { prop: 'direction', label: '方向' },
  { prop: 'status', label: '状态' },
  { prop: 'winRate', label: '胜率' },
  { prop: 'actions', label: '操作' },
]

const getDirectionClass = (direction: string) => {
  switch (direction) {
    case 'Long':
      return 'success'
    case 'Short':
      return 'danger'
    default:
      return 'info'
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'running':
      return 'success'
    case 'stopped':
      return 'danger'
    case 'cooling':
      return 'warning'
    default:
      return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'running':
      return '运行中'
    case 'stopped':
      return '已停止'
    case 'cooling':
      return '冷却中'
    default:
      return status
  }
}

const handlePause = (strategy: StrategyOverview) => {
  if (confirm(`确定要暂停策略 ${strategy.name} 吗？`)) {
    // TODO: 调用暂停 API
    if ((window as any).$toast) {
      ;(window as any).$toast.success('策略已暂停')
    }
  }
}

const handleStart = (strategy: StrategyOverview) => {
  if (confirm(`确定要启动策略 ${strategy.name} 吗？`)) {
    // TODO: 调用启动 API
    if ((window as any).$toast) {
      ;(window as any).$toast.success('策略已启动')
    }
  }
}

const showReason = (strategy: StrategyOverview) => {
  selectedReason.value = strategy.reason || '无原因'
  showReasonDialog.value = true
}
</script>

<style scoped>
.strategy-overview {
  padding: var(--spacing-sm) 0;
}

.strategy-overview__tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--border-radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.strategy-overview__tag--success {
  background-color: rgba(103, 194, 58, 0.1);
  color: var(--color-success);
}

.strategy-overview__tag--danger {
  background-color: rgba(245, 108, 108, 0.1);
  color: var(--color-danger);
}

.strategy-overview__tag--warning {
  background-color: rgba(230, 162, 60, 0.1);
  color: var(--color-warning);
}

.strategy-overview__tag--info {
  background-color: rgba(144, 147, 153, 0.1);
  color: var(--color-info);
}

.strategy-overview__actions {
  display: flex;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}
</style>
