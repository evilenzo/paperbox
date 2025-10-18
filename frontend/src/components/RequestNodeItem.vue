<script setup lang="ts">
import { ref } from 'vue'
import { main } from '@/lib/wailsjs/go/models'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Button } from '@/components/ui/button'
import { ChevronRight, Folder } from 'lucide-vue-next'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuShortcut,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'

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
    <ContextMenu>
      <ContextMenuTrigger as-child>
        <CollapsibleTrigger as-child>
          <Button variant="ghost" class="w-full justify-start gap-2 !px-2">
            <Folder class="size-4" />
            <span class="font-normal">{{ node.name }}</span>
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

  <ContextMenu v-else>
    <ContextMenuTrigger as-child>
      <Button class="w-full justify-start !px-2" variant="ghost">
        <span :class="`text-xs font-semibold ${getMethodColor(node.method)}`">{{
          node.method
        }}</span>
        <span class="font-normal">{{ node.name }}</span>
      </Button>
    </ContextMenuTrigger>
    <ContextMenuContent>
      <ContextMenuItem>
        <span>Save</span>
      </ContextMenuItem>
      <ContextMenuItem>
        <span>Rename</span>
      </ContextMenuItem>
      <ContextMenuItem>
        <span class="text-destructive">Delete</span>
      </ContextMenuItem>
    </ContextMenuContent>
  </ContextMenu>
</template>
