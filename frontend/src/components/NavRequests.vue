<script setup lang="ts">
import type { LucideIcon } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import RequestNodeItem from '@/components/RequestNodeItem.vue'

import { main } from '@/lib/wailsjs/go/models'
import { Greet } from '@/lib/wailsjs/go/main/App'

const requests = ref<main.RequestNode[]>([])

onMounted(async () => {
  requests.value = await Greet()
  console.log(requests.value)
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
  <div class="space-y-2 flex-1 flex-col">
    <RequestNodeItem v-for="request in requests" :key="request.name" :node="request" />
  </div>
</template>
