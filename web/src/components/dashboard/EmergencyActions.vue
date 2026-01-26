<template>
  <div class="emergency-actions">
    <div class="emergency-actions__buttons">
      <Button type="danger" @click="handleKillSwitch">
        <template #icon>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </template>
        Kill Switch (清仓)
      </Button>
      <Button type="warning" @click="handleResetCounters">
        <template #icon>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
        </template>
        Reset Daily Counters (重置计数)
      </Button>
    </div>
    
    <Dialog v-model="showKillDialog" title="Kill Switch" @close="showKillDialog = false">
      <p>确定要执行 Kill Switch 吗？这将清空所有持仓。</p>
      <template #footer>
        <Button @click="showKillDialog = false">取消</Button>
        <Button type="danger" @click="confirmKillSwitch">清仓</Button>
      </template>
    </Dialog>
    
    <Dialog v-model="showResetDialog" title="重置计数器" @close="showResetDialog = false">
      <p>确定要重置每日计数器吗？这将重置连续亏损计数等指标。</p>
      <template #footer>
        <Button @click="showResetDialog = false">取消</Button>
        <Button type="primary" @click="confirmReset">重置</Button>
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import Dialog from '@/components/ui/Dialog.vue'

const showKillDialog = ref(false)
const showResetDialog = ref(false)

const handleKillSwitch = () => {
  showKillDialog.value = true
}

const confirmKillSwitch = async () => {
  // TODO: 调用 Kill Switch API
  showKillDialog.value = false
  if ((window as any).$toast) {
    ;(window as any).$toast.success('Kill Switch 指令已发送')
  }
}

const handleResetCounters = () => {
  showResetDialog.value = true
}

const confirmReset = async () => {
  // TODO: 调用重置 API
  showResetDialog.value = false
  if ((window as any).$toast) {
    ;(window as any).$toast.success('计数器已重置')
  }
}
</script>

<style scoped>
.emergency-actions {
  padding: var(--spacing-sm) 0;
}

.emergency-actions__buttons {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

@media (max-width: 767px) {
  .emergency-actions__buttons {
    flex-direction: column;
  }
  
  .emergency-actions__buttons button {
    width: 100%;
  }
}
</style>
