<script setup lang="ts">
import type { LucideIcon } from 'lucide-vue-next'
import { onMounted, onUnmounted, ref, computed } from 'vue'
import RequestNodeItem from '@/components/RequestNodeItem.vue'
import { FolderPlus, Search, RefreshCw } from 'lucide-vue-next'

import { models, requests } from '@/lib/wailsjs/go/models'
import {
  GetRequests,
  SetRequestsPatch,
  AddRequest,
  AddFolder,
  AddRootFolder,
  DeleteItem,
} from '@/lib/wailsjs/go/main/App'
import { EventsOn, EventsOff, LogInfo, LogError } from '@/lib/wailsjs/runtime/runtime'
import Button from './ui/button/Button.vue'
import { Input } from './ui/input'

// Type for requests:updated event data
interface RequestsUpdatedEvent {
  version?: number
  values?: Record<string, requests.Item>
  Values?: Record<string, requests.Item>
  rootOrder?: string[]
  RootOrder?: string[]
}

const requestsData = ref<models.Requests | null>(null)
const rootOrder = ref<string[]>([])
const error = ref<string | null>(null)
const addingRequestTo = ref<string | null>(null)
const addingFolderTo = ref<string | null>(null)
const addingRootFolder = ref(false)
const newRootFolderInput = ref('')

// Load requests from backend
async function loadRequests() {
  try {
    const data = await GetRequests()
    requestsData.value = data
    rootOrder.value = data.rootOrder || []
    error.value = null
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load requests'
    LogError('Failed to load requests: ' + (err instanceof Error ? err.message : String(err)))
  }
}

// Setup event listeners
function setupEventListeners() {
  // Listen for requests:updated events (optimistic update)
  EventsOn('requests:updated', (updatedConfig: RequestsUpdatedEvent) => {
    LogInfo('Received requests:updated event')
    LogInfo(
      `Event data type: ${typeof updatedConfig}, keys: ${updatedConfig ? Object.keys(updatedConfig).join(', ') : 'null'}`,
    )

    // Convert to Requests format using the model's createFrom method
    if (updatedConfig) {
      const values = updatedConfig.values || updatedConfig.Values || {}
      const order = updatedConfig.rootOrder || updatedConfig.RootOrder || []
      LogInfo(`Updating requestsData with ${Object.keys(values).length} items`)

      // Log the item being renamed to verify it's in the update
      if (values && typeof values === 'object') {
        const itemKeys = Object.keys(values)
        if (itemKeys.length > 0 && itemKeys[0]) {
          const sampleKey = itemKeys[0]
          const sampleItem = values[sampleKey]
          LogInfo(`Sample item in event: ${sampleKey}, name="${sampleItem?.name || 'unknown'}"`)
        }
      }

      // Create a deep copy to ensure reactivity works correctly
      const valuesCopy = JSON.parse(JSON.stringify(values))
      const newRequests = models.Requests.createFrom({ values: valuesCopy })

      // Force reactivity by creating a new object reference
      requestsData.value = newRequests
      rootOrder.value = Array.isArray(order) ? [...order] : []

      LogInfo(
        `requestsData updated, ${requestsData.value?.values ? Object.keys(requestsData.value.values).length : 0} items`,
      )

      // Log specific item to verify it's updated
      if (requestsData.value?.values) {
        const itemIds = Object.keys(requestsData.value.values)
        if (itemIds.length > 0 && itemIds[0]) {
          const firstItemId = itemIds[0]
          const firstItem = requestsData.value.values[firstItemId]
          if (firstItem) {
            LogInfo(`Sample item ${firstItemId}: name="${firstItem.name}"`)
          }
        }
      }
    } else {
      LogError('updatedConfig is null or undefined!')
    }
    error.value = null
  })

  // Listen for requests:error events
  EventsOn('requests:error', (data: { message: string }) => {
    error.value = data.message
    LogError('Requests error: ' + data.message)
  })
}

// Cleanup event listeners
function cleanupEventListeners() {
  EventsOff('requests:updated')
  EventsOff('requests:error')
}

onMounted(async () => {
  setupEventListeners()
  await loadRequests()
})

onUnmounted(() => {
  cleanupEventListeners()
})

const itemsMap = computed(() => {
  if (!requestsData.value || !requestsData.value.values) {
    return {}
  }
  // Create a new object with new item references to ensure reactivity
  const map: Record<string, requests.Item> = {}
  for (const [id, item] of Object.entries(requestsData.value.values)) {
    // Create a new object for each item to ensure reactivity
    map[id] = { ...item }
  }
  return map
})

const rootItems = computed(() => {
  if (!requestsData.value || !requestsData.value.values) return []
  const map = itemsMap.value
  // Find root items (items that are not children of any folder)
  const allChildIds = new Set<string>()
  Object.values(map).forEach((item: requests.Item) => {
    if (item.children) {
      // children is array of UUID strings (after Wails regenerates, this will be string[])
      const childrenIds = Array.isArray(item.children)
        ? item.children.flat().map((id: unknown) => String(id))
        : []
      childrenIds.forEach((id: string) => allChildIds.add(id))
    }
  })
  const items = Object.entries(map)
    .filter(([id, item]: [string, requests.Item]) => {
      // Only show folders at root level
      return !allChildIds.has(id) && item.type === 'folder'
    })
    .map(([id, item]: [string, requests.Item]) => ({ id, item }))

  // Sort by rootOrder if available, otherwise by name
  if (rootOrder.value && rootOrder.value.length > 0) {
    const orderMap = new Map<string, number>()
    rootOrder.value.forEach((id, index) => {
      orderMap.set(id, index)
    })
    return items.sort((a, b) => {
      const aOrder = orderMap.get(a.id) ?? Infinity
      const bOrder = orderMap.get(b.id) ?? Infinity
      return aOrder - bOrder
    })
  }

  // Fallback: sort by name in reverse alphabetical order
  return items.sort((a, b) => b.item.name.localeCompare(a.item.name))
})

// Update requests config
async function updateRequests(newValues: Record<string, requests.Item>) {
  try {
    const patch = models.RequestsPatch.createFrom({ values: newValues })
    LogInfo(`Calling SetRequestsPatch with patch containing ${Object.keys(newValues).length} items`)

    await SetRequestsPatch(patch)
    LogInfo('SetRequestsPatch completed successfully')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : String(err)
    LogError(`Failed to update requests: ${errorMessage}`)
    error.value = errorMessage
    throw err
  }
}

// Start adding a new request (show input field)
function startAddingRequest(parentId: string) {
  addingRequestTo.value = parentId
}

// Cancel adding request
function cancelAddingRequest() {
  addingRequestTo.value = null
}

// Start adding a new folder (show input field)
function startAddingFolder(parentId: string) {
  addingFolderTo.value = parentId
}

// Cancel adding folder
function cancelAddingFolder() {
  addingFolderTo.value = null
}

// Start adding a root-level folder
function startAddingRootFolder() {
  addingRootFolder.value = true
  newRootFolderInput.value = ''
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const input = document.querySelector('#new-root-folder-input') as HTMLInputElement
      if (input) {
        input.focus()
        input.select()
      }
    })
  })
}

// Cancel adding root folder
function cancelAddingRootFolder() {
  addingRootFolder.value = false
  newRootFolderInput.value = ''
}

// Create root folder
async function createRootFolder() {
  const name = newRootFolderInput.value.trim()
  if (name) {
    try {
      await AddRootFolder(name)
      addingRootFolder.value = false
      newRootFolderInput.value = ''
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create folder'
      LogError(
        'Failed to create root folder: ' + (err instanceof Error ? err.message : String(err)),
      )
    }
  } else {
    cancelAddingRootFolder()
  }
}

// Handle root folder input blur
function handleRootFolderBlur() {
  setTimeout(() => {
    if (addingRootFolder.value) {
      createRootFolder()
    }
  }, 150)
}

// Add new item (request or folder) to a parent folder
async function addItem(parentId: string, type: 'request' | 'folder', name: string) {
  try {
    if (type === 'folder') {
      await AddFolder(parentId, name)
    } else {
      await AddRequest(parentId, name, 'GET', '')
    }
    // Clear adding state after successful creation
    if (type === 'request') {
      addingRequestTo.value = null
    } else if (type === 'folder') {
      addingFolderTo.value = null
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to add item'
    LogError('Failed to add item: ' + (err instanceof Error ? err.message : String(err)))
    throw err
  }
}

// Rename an item
async function renameItem(itemId: string, newName: string) {
  if (!requestsData.value || !requestsData.value.values) {
    LogError('Cannot rename: requestsData is not loaded')
    return
  }

  const newValues = { ...requestsData.value.values }
  const item = newValues[itemId]
  if (!item) {
    LogError(`Cannot rename: item ${itemId} not found`)
    return
  }

  LogInfo(`Renaming item ${itemId}: "${item.name}" -> "${newName}"`)

  newValues[itemId] = {
    ...item,
    name: newName,
  }

  try {
    LogInfo(`Calling updateRequests with ${Object.keys(newValues).length} items`)
    await updateRequests(newValues)
    LogInfo('updateRequests completed successfully')
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : String(err)
    LogError(`Failed to rename item: ${errorMessage}`)
    LogError(`Error stack: ${err instanceof Error ? err.stack : 'No stack trace'}`)
    error.value = errorMessage
  }
}

// Delete an item
async function deleteItem(itemId: string) {
  try {
    await DeleteItem(itemId)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete item'
    LogError('Failed to delete item: ' + (err instanceof Error ? err.message : String(err)))
    throw err
  }
}

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
    <div class="flex gap-1">
      <Button variant="ghost" class="flex-1 justify-center !px-2" @click="startAddingRootFolder">
        <FolderPlus class="size-4" />
      </Button>
      <Button variant="ghost" class="flex-1 justify-center !px-2">
        <Search class="size-4" />
      </Button>
      <Button variant="ghost" class="flex-1 justify-center !px-2" @click="loadRequests">
        <RefreshCw class="size-4" />
      </Button>
    </div>
    <!-- New root folder input field -->
    <div v-if="addingRootFolder" class="w-full">
      <Input
        id="new-root-folder-input"
        v-model="newRootFolderInput"
        class="h-9 text-sm"
        placeholder="Enter folder name..."
        @keyup.enter="createRootFolder"
        @keyup.esc="cancelAddingRootFolder"
        @blur="handleRootFolderBlur"
      />
    </div>
    <RequestNodeItem
      v-for="rootItem in rootItems"
      :key="rootItem.id"
      :item="rootItem.item"
      :item-id="rootItem.id"
      :items-map="itemsMap"
      :level="0"
      :adding-request-to="addingRequestTo"
      :adding-folder-to="addingFolderTo"
      @add-request="(parentId: string) => startAddingRequest(parentId)"
      @add-folder="(parentId: string) => startAddingFolder(parentId)"
      @create-request="(parentId: string, name: string) => addItem(parentId, 'request', name)"
      @create-folder="(parentId: string, name: string) => addItem(parentId, 'folder', name)"
      @cancel-add-request="cancelAddingRequest"
      @cancel-add-folder="cancelAddingFolder"
      @rename="(itemId: string, newName: string) => renameItem(itemId, newName)"
      @delete="(itemId: string) => deleteItem(itemId)"
    />
    <div v-if="error" class="rounded-md bg-destructive/10 p-2 text-xs text-destructive mt-2">
      {{ error }}
    </div>
  </div>
</template>
