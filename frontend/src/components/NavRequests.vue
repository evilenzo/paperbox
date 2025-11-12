<script setup lang="ts">
import type { LucideIcon } from 'lucide-vue-next'
import { onMounted, ref, computed } from 'vue'
import RequestNodeItem from '@/components/RequestNodeItem.vue'

import { main } from '@/lib/wailsjs/go/models'
import { GetRequests } from '@/lib/wailsjs/go/main/App'
import Button from './ui/button/Button.vue'

const requestsData = ref<main.Requests | null>(null)

onMounted(async () => {
  requestsData.value = await GetRequests()
  console.log('Requests data:', requestsData.value)
  console.log('Items map:', itemsMap.value)
  console.log('Root items:', rootItems.value)
})

const itemsMap = computed(() => {
  if (!requestsData.value || !requestsData.value.values) return {}
  return requestsData.value.values as Record<string, main.Item>
})

const rootItems = computed(() => {
  if (!requestsData.value || !requestsData.value.values) return []
  const map = itemsMap.value
  // Find root items (items that are not children of any folder)
  const allChildIds = new Set<string>()
  Object.values(map).forEach((item: main.Item) => {
    if (item.children) {
      // children is array of UUID strings (after Wails regenerates, this will be string[])
      const childrenIds = Array.isArray(item.children)
        ? item.children.flat().map((id: unknown) => String(id))
        : []
      childrenIds.forEach((id: string) => allChildIds.add(id))
    }
  })
  return Object.entries(map)
    .filter(([id]: [string, main.Item]) => !allChildIds.has(id))
    .map(([id, item]: [string, main.Item]) => ({ id, item }))
})

defineProps<{
  items: {
    title: string
    url: string
    icon?: LucideIcon
    isActive?: boolean
    items?: {
      title: string
      url: string
    }[]
  }[]
}>()
</script>

<template>
  <div class="space-y-1 flex-1 flex-col">
    <RequestNodeItem
      v-for="rootItem in rootItems"
      :key="rootItem.id"
      :item="rootItem.item"
      :item-id="rootItem.id"
      :items-map="itemsMap"
    />
    <Button variant="ghost" size="sm">Hello</Button>
    <Button variant="ghost" class="w-full justify-center gap-2 !px-2">Hii</Button>
  </div>
</template>
