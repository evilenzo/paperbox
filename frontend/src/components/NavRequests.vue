<script setup lang="ts">
import type { LucideIcon } from 'lucide-vue-next'
import { onMounted, onUnmounted, ref, computed } from 'vue'
import RequestNodeItem from '@/components/RequestNodeItem.vue'
import { FolderPlus, Search, RefreshCw } from 'lucide-vue-next'

import { models, requests } from '@/lib/wailsjs/go/models'
import { GetRequests, SetRequestsPatch, AddRequest, AddFolder } from '@/lib/wailsjs/go/main/App'
import { EventsOn, EventsOff, LogInfo, LogError } from '@/lib/wailsjs/runtime/runtime'
import Button from './ui/button/Button.vue'

// Type for requests:updated event data
interface RequestsUpdatedEvent {
  version?: number
  values?: Record<string, requests.Item>
  Values?: Record<string, requests.Item>
}

const requestsData = ref<models.Requests | null>(null)
const error = ref<string | null>(null)

// Load requests from backend
async function loadRequests() {
  try {
    requestsData.value = await GetRequests()
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
  return Object.entries(map)
    .filter(([id, item]: [string, requests.Item]) => {
      // Only show folders at root level
      return !allChildIds.has(id) && item.type === 'folder'
    })
    .map(([id, item]: [string, requests.Item]) => ({ id, item }))
})

// Update requests config
async function updateRequests(newValues: Record<string, requests.Item>) {
  try {
    const patch = { values: newValues }
    LogInfo(`Calling SetRequestsPatch with patch containing ${Object.keys(newValues).length} items`)
    LogInfo(
      `Patch structure: ${JSON.stringify(Object.keys(patch))}, values keys: ${Object.keys(newValues).slice(0, 3).join(', ')}...`,
    )

    const result = await SetRequestsPatch(patch)
    LogInfo(`SetRequestsPatch completed successfully, result: ${result}`)
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : String(err)
    const errorStack = err instanceof Error ? err.stack : 'No stack trace'
    LogError(`Failed to update requests: ${errorMessage}`)
    LogError(`Error stack: ${errorStack}`)
    error.value = errorMessage
    throw err
  }
}

// Add new item (request or folder) to a parent folder
async function addItem(parentId: string, type: 'request' | 'folder', name: string) {
  try {
    if (type === 'folder') {
      await AddFolder(parentId, name)
    } else {
      await AddRequest(parentId, name, 'GET', '')
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
  if (!requestsData.value || !requestsData.value.values) return

  const newValues = { ...requestsData.value.values }
  const item = newValues[itemId]

  // Remove from parent's children
  if (item) {
    Object.keys(newValues).forEach((parentId) => {
      const parent = newValues[parentId]
      if (parent && parent.type === 'folder' && parent.children) {
        const children = Array.isArray(parent.children)
          ? parent.children.filter((id: string) => id !== itemId)
          : []
        newValues[parentId] = {
          ...parent,
          children,
        }
      }
    })

    // If it's a folder, also delete all children recursively
    if (item.type === 'folder' && item.children) {
      const childrenIds = Array.isArray(item.children)
        ? item.children.map((id: unknown) => String(id))
        : []
      childrenIds.forEach((childId) => {
        delete newValues[childId]
      })
    }

    // Delete the item itself
    delete newValues[itemId]
  }

  await updateRequests(newValues)
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
      <Button variant="ghost" class="flex-1 justify-center !px-2">
        <FolderPlus class="size-4" />
      </Button>
      <Button variant="ghost" class="flex-1 justify-center !px-2">
        <Search class="size-4" />
      </Button>
      <Button variant="ghost" class="flex-1 justify-center !px-2" @click="loadRequests">
        <RefreshCw class="size-4" />
      </Button>
    </div>
    <RequestNodeItem
      v-for="rootItem in rootItems"
      :key="rootItem.id"
      :item="rootItem.item"
      :item-id="rootItem.id"
      :items-map="itemsMap"
      @add-request="(parentId: string) => addItem(parentId, 'request', 'New Request')"
      @add-folder="(parentId: string) => addItem(parentId, 'folder', 'New Folder')"
      @rename="(itemId: string, newName: string) => renameItem(itemId, newName)"
      @delete="(itemId: string) => deleteItem(itemId)"
    />
    <div v-if="error" class="rounded-md bg-destructive/10 p-2 text-xs text-destructive mt-2">
      {{ error }}
    </div>
  </div>
</template>
