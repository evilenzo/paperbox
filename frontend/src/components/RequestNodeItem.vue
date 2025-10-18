<!-- RequestNodeItem.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import { main } from '@/lib/wailsjs/go/models'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Button } from '@/components/ui/button'
import { ChevronRight, Folder } from 'lucide-vue-next'

defineProps<{
  node: main.RequestNode
  level?: number
}>()

const isOpen = ref(false)

const methodColors: Record<string, string> = {
  GET: 'text-blue-700',
  POST: 'text-green-700',
  PUT: 'text-yellow-700',
  DELETE: 'text-red-700',
  PATCH: 'text-purple-700',
  OPTIONS: 'text-cyan-700',
  HEAD: 'text-gray-700',
}

function getMethodColor(method: string) {
  return methodColors[method.toUpperCase()] ?? 'text-muted-foreground'
}
</script>

<template>
  <Collapsible v-if="node.children && node.children.length > 0" v-model:open="isOpen">
    <CollapsibleTrigger as-child>
      <Button variant="ghost" class="w-full justify-start gap-2 !px-2">
        <Folder class="size-4" />
        <span class="font-normal">{{ node.name }}</span>
        <ChevronRight
          :class="`ml-auto transition-transform duration-200 ${isOpen ? 'rotate-90' : ''}`"
        />
      </Button>
    </CollapsibleTrigger>
    <CollapsibleContent>
      <div :class="`ml-${(level || 1) * 6}`">
        <RequestNodeItem
          v-for="child in node.children"
          :key="child.name"
          :node="child"
          :level="(level || 0) + 1"
        />
      </div>
    </CollapsibleContent>
  </Collapsible>

  <Button v-else class="w-full justify-start !px-2" variant="ghost">
    <span :class="`text-xs font-semibold ${getMethodColor(node.method)}`">{{ node.method }}</span>
    <span class="font-normal">{{ node.name }}</span>
  </Button>
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
