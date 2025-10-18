<script setup lang="ts">
import type { LucideIcon } from 'lucide-vue-next'
import { ChevronRight, Folder } from 'lucide-vue-next'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Button } from '@/components/ui/button'
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
  <!-- <div class="space-y-2">
    <h2 class="px-2 text-xs font-medium text-muted-foreground">Requests</h2>
    <div class="space-y-1">
      <Collapsible
        v-for="item in items"
        :key="item.title"
        as-child
        :default-open="item.isActive"
        class="group/collapsible"
      >
        <div>
          <CollapsibleTrigger as-child>
            <Button variant="ghost" class="w-full justify-start gap-2 !px-2">
              <component :is="Folder" class="size-4" />
              <span class="font-normal">{{ item.title }}</span>
              <ChevronRight
                class="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
              />
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="ml-6 space-y-1">
              <Button
                v-for="subItem in item.items"
                :key="subItem.title"
                variant="ghost"
                size="sm"
                class="w-full justify-start"
                as-child
              >
                <a :href="subItem.url">
                  <span class="font-normal">{{ subItem.title }}</span>
                </a>
              </Button>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    </div>
  </div> -->
</template>
