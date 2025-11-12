<script setup lang="ts">
import { ref, computed } from 'vue'
import { requests } from '@/lib/wailsjs/go/models'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Button } from '@/components/ui/button'
import { ChevronRight, Folder } from 'lucide-vue-next'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'

const props = defineProps<{
  item: requests.Item
  itemId: string
  itemsMap: Record<string, requests.Item>
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

function getMethodColor(method?: string) {
  if (!method) return 'text-muted-foreground'
  return methodColors[method.toUpperCase()] ?? 'text-muted-foreground'
}

const childItems = computed((): Array<{ id: string; item: requests.Item }> => {
  if (props.item.type !== 'folder' || !props.item.children) return []
  // children is array of UUID strings (after Wails regenerates, this will be string[])
  const childrenIds = Array.isArray(props.item.children)
    ? props.item.children.flat().map((id: unknown) => String(id))
    : []
  return childrenIds
    .map((id: string) => ({ id, item: props.itemsMap[id] }))
    .filter(
      (entry: {
        id: string
        item: requests.Item | undefined
      }): entry is { id: string; item: requests.Item } => entry.item !== undefined,
    )
})
</script>

<template>
  <Collapsible v-if="item.type === 'folder' && childItems.length > 0" v-model:open="isOpen">
    <ContextMenu>
      <ContextMenuTrigger as-child>
        <CollapsibleTrigger as-child>
          <Button variant="ghost" class="w-full justify-start gap-2 !px-2">
            <Folder class="size-4" />
            <span class="font-normal">{{ item.name }}</span>
            <ChevronRight
              :class="`ml-auto transition-transform duration-200 ${isOpen ? 'rotate-90' : ''}`"
            />
          </Button>
        </CollapsibleTrigger>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuSub>
          <ContextMenuSubTrigger>Add</ContextMenuSubTrigger>
          <ContextMenuSubContent class="w-48">
            <ContextMenuItem>Request</ContextMenuItem>
            <ContextMenuItem>Folder</ContextMenuItem>
          </ContextMenuSubContent>
        </ContextMenuSub>
        <ContextMenuItem>
          <span>Rename</span>
        </ContextMenuItem>
        <ContextMenuItem>
          <span class="text-destructive">Delete</span>
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
    <CollapsibleContent>
      <div :style="{ marginLeft: `${(level || 1) * 24}px` }" class="space-y-1">
        <RequestNodeItem
          v-for="child in childItems"
          :key="child.id"
          :item="child.item"
          :item-id="child.id"
          :items-map="itemsMap"
          :level="(level || 0) + 1"
        />
      </div>
    </CollapsibleContent>
  </Collapsible>

  <ContextMenu v-else>
    <ContextMenuTrigger as-child>
      <Button class="w-full justify-start gap-2 !px-2" variant="ghost">
        <Folder v-if="item.type === 'folder'" class="size-4" />
        <span v-if="item.method" :class="`text-xs font-semibold ${getMethodColor(item.method)}`">{{
          item.method
        }}</span>
        <span class="font-normal">{{ item.name }}</span>
      </Button>
    </ContextMenuTrigger>
    <ContextMenuContent>
      <ContextMenuSub v-if="item.type === 'folder'">
        <ContextMenuSubTrigger>Add</ContextMenuSubTrigger>
        <ContextMenuSubContent class="w-48">
          <ContextMenuItem>Request</ContextMenuItem>
          <ContextMenuItem>Folder</ContextMenuItem>
        </ContextMenuSubContent>
      </ContextMenuSub>
      <ContextMenuItem>
        <span>Rename</span>
      </ContextMenuItem>
      <ContextMenuItem>
        <span class="text-destructive">Delete</span>
      </ContextMenuItem>
    </ContextMenuContent>
  </ContextMenu>
</template>
