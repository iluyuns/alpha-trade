<template>
  <div class="dashboard">
    <PullRefresh @refresh="handleRefresh">
      <!-- 核心指标 -->
      <div class="dashboard__metrics">
        <Card
          v-for="metric in coreMetrics"
          :key="metric.label"
          class="dashboard__metric-card"
          :class="metric.status"
        >
          <div class="metric-content">
            <div class="metric-label">{{ metric.label }}</div>
            <div class="metric-value" :key="metric.value">{{ metric.value }}</div>
            <div class="metric-change" v-if="metric.change !== null">
              <svg
                v-if="metric.change > 0"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
              >
                <polyline points="18 15 12 9 6 15"></polyline>
              </svg>
              <svg
                v-else-if="metric.change < 0"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
              >
                <polyline points="6 9 12 15 18 9"></polyline>
              </svg>
              {{ Math.abs(metric.change).toFixed(2) }}%
            </div>
          </div>
        </Card>
      </div>

      <!-- 系统健康与风控状态 -->
      <div class="dashboard__status">
        <Card class="dashboard__status-card">
          <template #header>系统健康 (System Health)</template>
          <SystemHealth :health="dashboardData?.systemHealth || []" />
        </Card>
        <Card class="dashboard__status-card">
          <template #header>风控状态 (Risk Status)</template>
          <RiskStatus :risk-status="dashboardData?.riskStatus" />
        </Card>
      </div>

      <!-- 策略概览 -->
      <Card class="dashboard__strategies">
        <template #header>策略概览 (Active Strategies)</template>
        <StrategyOverview :strategies="dashboardData?.strategies || []" />
      </Card>

      <!-- 紧急操作 -->
      <Card class="dashboard__emergency">
        <template #header>紧急操作 (Emergency Actions)</template>
        <EmergencyActions />
      </Card>
    </PullRefresh>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useDashboardStore } from '@/stores/dashboard'
import { useAuthStore } from '@/stores/auth'
import Card from '@/components/ui/Card.vue'
import SystemHealth from '@/components/dashboard/SystemHealth.vue'
import RiskStatus from '@/components/dashboard/RiskStatus.vue'
import StrategyOverview from '@/components/dashboard/StrategyOverview.vue'
import EmergencyActions from '@/components/dashboard/EmergencyActions.vue'
import PullRefresh from '@/components/ui/PullRefresh.vue'

const dashboardStore = useDashboardStore()
const authStore = useAuthStore()

const dashboardData = computed(() => dashboardStore.data)

const coreMetrics = computed(() => {
  if (!dashboardData.value) return []
  
  const data = dashboardData.value
  return [
    {
      label: '今日盈亏',
      value: `$${parseFloat(data.pnlDaily || '0').toFixed(2)}`,
      change: parseFloat(data.pnlPercent || '0'),
      status: parseFloat(data.pnlDaily || '0') >= 0 ? 'positive' : 'negative',
    },
    {
      label: '当前权益',
      value: `$${parseFloat(data.totalEquity || '0').toFixed(2)}`,
      change: null,
      status: 'neutral',
    },
    {
      label: '风险敞口',
      value: `${parseFloat(data.riskExposure || '0').toFixed(1)}%`,
      change: null,
      status: parseFloat(data.riskExposure || '0') > 80 ? 'warning' : 'normal',
    },
    {
      label: '日内回撤',
      value: `${parseFloat(data.dailyDrawdown || '0').toFixed(2)}%`,
      change: null,
      status: parseFloat(data.dailyDrawdown || '0') > 5 ? 'danger' : 'normal',
    },
  ]
})

const handleRefresh = async () => {
  await dashboardStore.fetchDashboard()
}

onMounted(() => {
  dashboardStore.fetchDashboard()
  
  if (authStore.token) {
    dashboardStore.connectWebSocket(authStore.token)
  }
})

onUnmounted(() => {
  dashboardStore.disconnectWebSocket()
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.dashboard__metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.dashboard__metric-card {
  transition: all var(--transition-base);
}

.dashboard__metric-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.metric-content {
  text-align: center;
}

.metric-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-sm);
}

.metric-value {
  font-size: var(--font-size-xl);
  font-weight: bold;
  margin-bottom: var(--spacing-xs);
  transition: all var(--transition-base);
}

.dashboard__metric-card.positive .metric-value {
  color: var(--color-success);
}

.dashboard__metric-card.negative .metric-value {
  color: var(--color-danger);
}

.dashboard__metric-card.warning .metric-value {
  color: var(--color-warning);
}

.dashboard__metric-card.danger .metric-value {
  color: var(--color-danger);
}

.metric-change {
  font-size: var(--font-size-xs);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: var(--color-text-secondary);
}

.dashboard__status {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-md);
}

.dashboard__status-card {
  min-height: 200px;
}

.dashboard__strategies,
.dashboard__emergency {
  width: 100%;
}

@media (max-width: 767px) {
  .dashboard__metrics {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-sm);
  }
  
  .dashboard__status {
    grid-template-columns: 1fr;
  }
  
  .metric-value {
    font-size: var(--font-size-lg);
  }
}
</style>
