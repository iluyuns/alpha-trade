<template>
  <div class="risk-status">
    <div class="risk-status__item">
      <div class="risk-status__label">连续亏损计数</div>
      <div class="risk-status__value">
        <div class="risk-status__progress">
          <div
            class="risk-status__progress-bar"
            :style="{ width: `${getLossPercentage()}%`, backgroundColor: getLossColor() }"
          ></div>
        </div>
        <div class="risk-status__progress-text">
          {{ riskStatus?.consecutiveLosses || 0 }} / {{ riskStatus?.maxConsecutiveLosses || 5 }}
        </div>
      </div>
    </div>
    
    <div class="risk-status__item">
      <div class="risk-status__label">宏观冷却模式</div>
      <div class="risk-status__value">
        <span
          :class="[
            'risk-status__tag',
            riskStatus?.macroCoolingMode === 'active' ? 'risk-status__tag--warning' : 'risk-status__tag--info'
          ]"
        >
          {{ riskStatus?.macroCoolingMode === 'active' ? 'Active' : 'Inactive' }}
        </span>
        <span v-if="riskStatus?.nextMacroWindow" class="risk-status__hint">
          (下个窗口: {{ riskStatus.nextMacroWindow }})
        </span>
      </div>
    </div>
    
    <div class="risk-status__item">
      <div class="risk-status__label">杠杆限制状态</div>
      <div class="risk-status__value">
        <span
          :class="[
            'risk-status__tag',
            riskStatus?.leverageStatus === 'relaxed' ? 'risk-status__tag--success' : 'risk-status__tag--warning'
          ]"
        >
          {{ riskStatus?.leverageStatus === 'relaxed' ? 'Relaxed' : 'Restricted' }}
        </span>
        <span class="risk-status__hint">
          (允许最大 {{ riskStatus?.maxLeverage || '2.0' }}x, 当前实际 {{ riskStatus?.currentLeverage || '1.0' }}x)
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, computed } from 'vue'
import type { RiskStatus } from '@/stores/dashboard'

const props = defineProps<{
  riskStatus?: RiskStatus
}>()

const getLossPercentage = () => {
  if (!props.riskStatus) return 0
  const { consecutiveLosses, maxConsecutiveLosses } = props.riskStatus
  return Math.min((consecutiveLosses / maxConsecutiveLosses) * 100, 100)
}

const getLossColor = () => {
  const percentage = getLossPercentage()
  if (percentage >= 80) return 'var(--color-danger)'
  if (percentage >= 60) return 'var(--color-warning)'
  return 'var(--color-success)'
}
</script>

<style scoped>
.risk-status {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.risk-status__item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-sm) 0;
  border-bottom: 1px solid var(--color-border-lighter);
}

.risk-status__item:last-child {
  border-bottom: none;
}

.risk-status__label {
  font-weight: 500;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  flex-shrink: 0;
  min-width: 120px;
}

.risk-status__value {
  flex: 1;
  text-align: right;
}

.risk-status__progress {
  width: 100%;
  height: 8px;
  background-color: var(--color-border-lighter);
  border-radius: var(--border-radius-round);
  overflow: hidden;
  margin-bottom: var(--spacing-xs);
}

.risk-status__progress-bar {
  height: 100%;
  transition: all var(--transition-base);
  border-radius: var(--border-radius-round);
}

.risk-status__progress-text {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.risk-status__tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--border-radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.risk-status__tag--success {
  background-color: rgba(103, 194, 58, 0.1);
  color: var(--color-success);
}

.risk-status__tag--warning {
  background-color: rgba(230, 162, 60, 0.1);
  color: var(--color-warning);
}

.risk-status__tag--info {
  background-color: rgba(144, 147, 153, 0.1);
  color: var(--color-info);
}

.risk-status__hint {
  margin-left: var(--spacing-xs);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

@media (max-width: 767px) {
  .risk-status__item {
    flex-direction: column;
    gap: var(--spacing-xs);
  }
  
  .risk-status__value {
    text-align: left;
  }
  
  .risk-status__label {
    min-width: auto;
  }
}
</style>
