<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { requests } from '@/lib/wailsjs/go/models'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  addingRequestTo?: string | null
  addingFolderTo?: string | null
}>()

const emit = defineEmits<{
  'add-request': [parentId: string]
  'add-folder': [parentId: string]
  'create-request': [parentId: string, name: string]
  'create-folder': [parentId: string, name: string]
  'cancel-add-request': []
  'cancel-add-folder': []
  rename: [itemId: string, newName: string]
  delete: [itemId: string]
}>()

const isOpen = ref(false)
const isRenaming = ref(false)
const renameInput = ref('')
const renameStartTime = ref(0)
const newRequestInput = ref('')
const newRequestStartTime = ref(0)
const newFolderInput = ref('')
const newFolderStartTime = ref(0)

// Use computed to get the current item from itemsMap to ensure reactivity
const currentItem = computed(() => {
  return props.itemsMap[props.itemId] || props.item
})

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
  const item = currentItem.value
  if (item.type !== 'folder' || !item.children) return []
  // children is array of UUID strings (after Wails regenerates, this will be string[])
  const childrenIds = Array.isArray(item.children)
    ? item.children.flat().map((id: unknown) => String(id))
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

// Check if we can add a folder (max 3 levels: root(0) -> nested(1) -> nested(2) -> request)
// level >= 2 means we're at the third folder level, so we can't add more folders
const canAddFolder = computed(() => {
  const currentLevel = props.level ?? 0
  return currentLevel < 2
})

// Check if we're currently adding a request to this folder
const isAddingRequest = computed(() => {
  return props.addingRequestTo === props.itemId && currentItem.value.type === 'folder'
})

// Check if we're currently adding a folder to this folder
const isAddingFolder = computed(() => {
  return props.addingFolderTo === props.itemId && currentItem.value.type === 'folder'
})

// Auto-open folder when adding request or folder
watch(isAddingRequest, (newValue) => {
  if (newValue && currentItem.value.type === 'folder') {
    isOpen.value = true
  }
})

watch(isAddingFolder, (newValue) => {
  if (newValue && currentItem.value.type === 'folder') {
    isOpen.value = true
  }
})

function handleAddRequest() {
  if (currentItem.value.type === 'folder') {
    emit('add-request', props.itemId)
    // Reset input and focus it
    newRequestInput.value = ''
    newRequestStartTime.value = Date.now()
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        const input = document.querySelector(
          `#new-request-input-${props.itemId}`,
        ) as HTMLInputElement
        if (input) {
          input.focus()
          input.select()
        }
      })
    })
  }
}

function handleCreateRequest() {
  const name = newRequestInput.value.trim()
  if (name) {
    emit('create-request', props.itemId, name)
    newRequestInput.value = ''
  } else {
    emit('cancel-add-request')
  }
}

function handleCancelAddRequest() {
  emit('cancel-add-request')
  newRequestInput.value = ''
  // Close folder if it's empty
  if (currentItem.value.type === 'folder' && childItems.value.length === 0) {
    isOpen.value = false
  }
}

function handleNewRequestBlur(event: FocusEvent) {
  // Don't submit if focus is moving to another element in the same component
  const relatedTarget = event.relatedTarget as HTMLElement | null
  if (relatedTarget && relatedTarget.closest(`[data-new-request-id="${props.itemId}"]`)) {
    return
  }

  // Don't submit if input just started (user might be moving mouse after clicking)
  const timeSinceStart = Date.now() - newRequestStartTime.value
  if (timeSinceStart < 300) {
    // If blur happened too quickly, refocus the input
    requestAnimationFrame(() => {
      const input = document.querySelector(`#new-request-input-${props.itemId}`) as HTMLInputElement
      if (input && isAddingRequest.value) {
        input.focus()
        input.select()
      }
    })
    return
  }

  // Submit on blur with small delay
  setTimeout(() => {
    if (isAddingRequest.value) {
      handleCreateRequest()
    }
  }, 150)
}

function handleAddFolder() {
  if (currentItem.value.type === 'folder' && canAddFolder.value) {
    emit('add-folder', props.itemId)
    // Reset input and focus it
    newFolderInput.value = ''
    newFolderStartTime.value = Date.now()
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        const input = document.querySelector(
          `#new-folder-input-${props.itemId}`,
        ) as HTMLInputElement
        if (input) {
          input.focus()
          input.select()
        }
      })
    })
  }
}

function handleCreateFolder() {
  const name = newFolderInput.value.trim()
  if (name) {
    emit('create-folder', props.itemId, name)
    newFolderInput.value = ''
  } else {
    emit('cancel-add-folder')
  }
}

function handleCancelAddFolder() {
  emit('cancel-add-folder')
  newFolderInput.value = ''
  // Close folder if it's empty
  if (currentItem.value.type === 'folder' && childItems.value.length === 0) {
    isOpen.value = false
  }
}

function handleNewFolderBlur(event: FocusEvent) {
  // Don't submit if focus is moving to another element in the same component
  const relatedTarget = event.relatedTarget as HTMLElement | null
  if (relatedTarget && relatedTarget.closest(`[data-new-folder-id="${props.itemId}"]`)) {
    return
  }

  // Don't submit if input just started (user might be moving mouse after clicking)
  const timeSinceStart = Date.now() - newFolderStartTime.value
  if (timeSinceStart < 300) {
    // If blur happened too quickly, refocus the input
    requestAnimationFrame(() => {
      const input = document.querySelector(`#new-folder-input-${props.itemId}`) as HTMLInputElement
      if (input && isAddingFolder.value) {
        input.focus()
        input.select()
      }
    })
    return
  }

  // Submit on blur with small delay
  setTimeout(() => {
    if (isAddingFolder.value) {
      handleCreateFolder()
    }
  }, 150)
}

function handleRename() {
  isRenaming.value = true
  renameInput.value = currentItem.value.name
  renameStartTime.value = Date.now()
  // Use requestAnimationFrame to ensure DOM is updated before focusing
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const input = document.querySelector(`#rename-input-${props.itemId}`) as HTMLInputElement
      if (input) {
        input.focus()
        input.select()
      }
    })
  })
}

async function handleRenameSubmit() {
  const newName = renameInput.value.trim()
  // Always close rename mode first to show the button again
  isRenaming.value = false
  renameInput.value = ''

  // Only emit if name actually changed
  if (newName && newName !== currentItem.value.name) {
    emit('rename', props.itemId, newName)
  }
}

function handleRenameCancel() {
  isRenaming.value = false
  renameInput.value = ''
}

function handleRenameBlur(event: FocusEvent) {
  // Don't submit if focus is moving to another element in the same component
  const relatedTarget = event.relatedTarget as HTMLElement | null
  if (relatedTarget && relatedTarget.closest(`[data-item-id="${props.itemId}"]`)) {
    return
  }

  // Don't submit if rename just started (user might be moving mouse after clicking rename)
  const timeSinceStart = Date.now() - renameStartTime.value
  if (timeSinceStart < 300) {
    // If blur happened too quickly, refocus the input
    requestAnimationFrame(() => {
      const input = document.querySelector(`#rename-input-${props.itemId}`) as HTMLInputElement
      if (input && isRenaming.value) {
        input.focus()
        input.select()
      }
    })
    return
  }

  // Don't submit on blur if Enter was just pressed (it will handle submission)
  // Submit on blur with small delay only if Enter wasn't pressed
  setTimeout(() => {
    if (isRenaming.value) {
      handleRenameSubmit()
    }
  }, 150)
}

function handleDelete() {
  emit('delete', props.itemId)
}
</script>

<template>
  <Collapsible v-if="currentItem.type === 'folder'" v-model:open="isOpen">
    <ContextMenu>
      <ContextMenuTrigger as-child>
        <CollapsibleTrigger as-child>
          <div :data-item-id="itemId" class="w-full">
            <Input
              v-if="isRenaming"
              :id="`rename-input-${itemId}`"
              v-model="renameInput"
              class="h-9 text-sm"
              @keyup.enter="handleRenameSubmit"
              @keyup.esc="handleRenameCancel"
              @blur="handleRenameBlur"
            />
            <Button v-else variant="ghost" class="w-full justify-start gap-2 !px-2">
              <Folder class="size-4" />
              <span class="font-normal">{{ currentItem.name }}</span>
              <ChevronRight
                v-if="childItems.length > 0 || isAddingRequest || isAddingFolder"
                :class="`ml-auto transition-transform duration-200 ${isOpen ? 'rotate-90' : ''}`"
              />
            </Button>
          </div>
        </CollapsibleTrigger>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuSub>
          <ContextMenuSubTrigger>Add</ContextMenuSubTrigger>
          <ContextMenuSubContent class="w-48">
            <ContextMenuItem @click="handleAddRequest">Request</ContextMenuItem>
            <ContextMenuItem v-if="canAddFolder" @click="handleAddFolder">Folder</ContextMenuItem>
          </ContextMenuSubContent>
        </ContextMenuSub>
        <ContextMenuItem @click="handleRename">
          <span>Rename</span>
        </ContextMenuItem>
        <ContextMenuItem @click="handleDelete">
          <span class="text-destructive">Delete</span>
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
    <CollapsibleContent>
      <div class="ml-6 space-y-1">
        <RequestNodeItem
          v-for="child in childItems"
          :key="child.id"
          :item="child.item"
          :item-id="child.id"
          :items-map="itemsMap"
          :level="(level || 0) + 1"
          :adding-request-to="addingRequestTo"
          :adding-folder-to="addingFolderTo"
          @add-request="(parentId: string) => $emit('add-request', parentId)"
          @add-folder="(parentId: string) => $emit('add-folder', parentId)"
          @create-request="
            (parentId: string, name: string) => $emit('create-request', parentId, name)
          "
          @create-folder="
            (parentId: string, name: string) => $emit('create-folder', parentId, name)
          "
          @cancel-add-request="$emit('cancel-add-request')"
          @cancel-add-folder="$emit('cancel-add-folder')"
          @rename="(itemId: string, newName: string) => $emit('rename', itemId, newName)"
          @delete="(itemId: string) => $emit('delete', itemId)"
        />
        <!-- New request input field -->
        <div v-if="isAddingRequest" :data-new-request-id="itemId" class="w-full">
          <Input
            :id="`new-request-input-${itemId}`"
            v-model="newRequestInput"
            class="h-9 text-sm"
            placeholder="Enter request name..."
            @keyup.enter="handleCreateRequest"
            @keyup.esc="handleCancelAddRequest"
            @blur="handleNewRequestBlur"
          />
        </div>
        <!-- New folder input field -->
        <div v-if="isAddingFolder" :data-new-folder-id="itemId" class="w-full">
          <Input
            :id="`new-folder-input-${itemId}`"
            v-model="newFolderInput"
            class="h-9 text-sm"
            placeholder="Enter folder name..."
            @keyup.enter="handleCreateFolder"
            @keyup.esc="handleCancelAddFolder"
            @blur="handleNewFolderBlur"
          />
        </div>
      </div>
    </CollapsibleContent>
  </Collapsible>

  <div v-else-if="currentItem.type !== 'folder'">
    <ContextMenu>
      <ContextMenuTrigger as-child>
        <div :data-item-id="itemId" class="w-full">
          <Input
            v-if="isRenaming"
            :id="`rename-input-${itemId}`"
            v-model="renameInput"
            class="h-9 text-sm"
            @keyup.enter="handleRenameSubmit"
            @keyup.esc="handleRenameCancel"
            @blur="handleRenameBlur"
          />
          <Button v-else class="w-full justify-start gap-2 !px-2" variant="ghost">
            <Folder v-if="currentItem.type === 'folder'" class="size-4" />
            <span
              v-if="currentItem.method"
              :class="`text-xs font-semibold ${getMethodColor(currentItem.method)}`"
              >{{ currentItem.method }}</span
            >
            <span class="font-normal">{{ currentItem.name }}</span>
          </Button>
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuSub v-if="currentItem.type === 'folder'">
          <ContextMenuSubTrigger>Add</ContextMenuSubTrigger>
          <ContextMenuSubContent class="w-48">
            <ContextMenuItem @click="handleAddRequest">Request</ContextMenuItem>
            <ContextMenuItem v-if="canAddFolder" @click="handleAddFolder">Folder</ContextMenuItem>
          </ContextMenuSubContent>
        </ContextMenuSub>
        <ContextMenuItem @click="handleRename">
          <span>Rename</span>
        </ContextMenuItem>
        <ContextMenuItem @click="handleDelete">
          <span class="text-destructive">Delete</span>
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  </div>
</template>
