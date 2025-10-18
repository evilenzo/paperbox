<script setup lang="ts">
import { ref } from 'vue'
import LeftSidebar from '@/components/LeftSidebar.vue'
// import { useColorMode } from '@vueuse/core'
import MainContent from '@/components/MainContent.vue'
import type { Header } from '@/components/MainContent.vue'
import { ResizablePanel, ResizablePanelGroup, ResizableHandle } from '@/components/ui/resizable'

// TODO: Add color mode switcher
// const mode = useColorMode()

const headers = ref<Header[]>([
  { name: 'Content-Type', value: 'application/json', inactive: true },
  { name: 'Authorization', value: 'Bearer token', inactive: true },
])

const updateHeaders = (newHeaders: { name: string; value: string; inactive: boolean }[]) => {
  headers.value = newHeaders
}
</script>

<template>
  <ResizablePanelGroup direction="horizontal" class="min-h-screen h-screen w-screen">
    <ResizablePanel :default-size="22" :min-size="22">
      <LeftSidebar />
    </ResizablePanel>
    <ResizableHandle />
    <ResizablePanel :default-size="75" :min-size="50">
      <ResizablePanelGroup direction="vertical" class="grow">
        <ResizablePanel :default-size="75" :min-size="25">
          <MainContent :headers="headers" @update:headers="updateHeaders" />
        </ResizablePanel>
        <ResizableHandle />
        <ResizablePanel :default-size="25" :min-size="25">
          <div class="flex h-full items-center justify-center p-6">
            <span class="font-semibold">Results</span>
          </div>
        </ResizablePanel>
      </ResizablePanelGroup>
      <div class="flex h-full items-center justify-center p-6">
        <span class="font-semibold">Content</span>
      </div>
    </ResizablePanel>
    <ResizableHandle />
    <ResizablePanel :default-size="5" :min-size="5">
      <div class="flex h-full items-center justify-center p-6">
        <span class="font-semibold">Sidebar</span>
      </div>
    </ResizablePanel>
  </ResizablePanelGroup>
</template>
