<template>
  <div :class="['ui-table', { 'ui-table--mobile': isMobile }]">
    <div v-if="isMobile" class="ui-table__mobile">
      <div
        v-for="(row, index) in data"
        :key="index"
        class="ui-table__mobile-card"
      >
        <div
          v-for="column in columns"
          :key="column.prop"
          class="ui-table__mobile-cell"
        >
          <div class="ui-table__mobile-label">{{ column.label }}</div>
          <div class="ui-table__mobile-value">
            <slot
              :name="`cell-${column.prop}`"
              :row="row"
              :column="column"
              :value="row[column.prop]"
            >
              {{ row[column.prop] }}
            </slot>
          </div>
        </div>
      </div>
    </div>
    <table v-else class="ui-table__desktop">
      <thead>
        <tr>
          <th v-for="column in columns" :key="column.prop">
            {{ column.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, index) in data" :key="index">
          <td v-for="column in columns" :key="column.prop">
            <slot
              :name="`cell-${column.prop}`"
              :row="row"
              :column="column"
              :value="row[column.prop]"
            >
              {{ row[column.prop] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

export interface TableColumn {
  prop: string
  label: string
  width?: string
}

interface Props {
  data: any[]
  columns: TableColumn[]
}

defineProps<Props>()

const isMobile = ref(window.innerWidth < 768)

const updateMobile = () => {
  isMobile.value = window.innerWidth < 768
}

onMounted(() => {
  window.addEventListener('resize', updateMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateMobile)
})
</script>

<style scoped>
.ui-table {
  width: 100%;
  background-color: var(--color-bg);
  border-radius: var(--border-radius);
  overflow: hidden;
}

.ui-table__desktop {
  width: 100%;
  border-collapse: collapse;
}

.ui-table__desktop thead {
  background-color: var(--color-bg-page);
}

.ui-table__desktop th {
  padding: var(--spacing-md);
  text-align: left;
  font-weight: 500;
  color: var(--color-text-regular);
  border-bottom: 1px solid var(--color-border-light);
}

.ui-table__desktop td {
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--color-border-lighter);
  color: var(--color-text-primary);
}

.ui-table__desktop tbody tr:hover {
  background-color: var(--color-bg-page);
}

.ui-table__mobile {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
}

.ui-table__mobile-card {
  background-color: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--border-radius);
  padding: var(--spacing-md);
}

.ui-table__mobile-cell {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-sm) 0;
  border-bottom: 1px solid var(--color-border-lighter);
}

.ui-table__mobile-cell:last-child {
  border-bottom: none;
}

.ui-table__mobile-label {
  font-weight: 500;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.ui-table__mobile-value {
  color: var(--color-text-primary);
  text-align: right;
  flex: 1;
  margin-left: var(--spacing-md);
}
</style>
