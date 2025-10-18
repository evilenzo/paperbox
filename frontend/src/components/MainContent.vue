<script setup lang="ts">
import { ref } from 'vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import Tabs from '@/components/ui/tabs/Tabs.vue'
import TabsList from '@/components/ui/tabs/TabsList.vue'
import TabsTrigger from '@/components/ui/tabs/TabsTrigger.vue'
import { Trash } from 'lucide-vue-next'

export type Header = {
  name: string
  value: string
  inactive: boolean
}

const props = defineProps<{
  headers: Header[]
}>()

const emit = defineEmits<{
  'update:headers': [headers: Header[]]
}>()

// Local reactive state - всегда с одной пустой парой в конце
const localHeaders = ref<Header[]>([...props.headers, { name: '', value: '', inactive: false }])

const updateHeader = (index: number, field: 'name' | 'value', value: string) => {
  const header = localHeaders.value[index]
  if (header) {
    header[field] = value

    // Если это последняя пара и оба поля заполнены - добавить новую пустую
    if (
      index === localHeaders.value.length - 1 &&
      header.name.trim() !== '' &&
      header.value.trim() !== ''
    ) {
      localHeaders.value.push({ name: '', value: '', inactive: false })
    }

    // Emit только непустые заголовки
    const nonEmpty = localHeaders.value.filter((h) => h.name.trim() !== '' && h.value.trim() !== '')
    emit('update:headers', nonEmpty)
  }
}
</script>

<template>
  <div class="flex flex-1 flex-col p-2 gap-2 max-w-full overflow-hidden">
    <div class="flex flex-row gap-2">
      <Input placeholder="Enter your URL here..." />
      <Button>Send</Button>
    </div>
    <div class="flex flex-row gap-2">
      <Tabs default-value="header">
        <TabsList class="grid w-full grid-cols-3">
          <TabsTrigger value="header" class="font-normal">Header</TabsTrigger>
          <TabsTrigger value="body" class="font-normal">Body</TabsTrigger>
          <TabsTrigger value="cookies" class="font-normal">Cookies</TabsTrigger>
        </TabsList>
      </Tabs>
    </div>
    <div class="flex flex-col gap-2">
      <div
        v-for="(header, index) in localHeaders"
        :key="`header-${index}`"
        class="flex flex-row gap-2"
      >
        <Input
          placeholder="Name..."
          :model-value="header.name"
          :disabled="header.inactive"
          @update:model-value="(val: string | number) => updateHeader(index, 'name', String(val))"
        />
        <Input
          placeholder="Value..."
          :model-value="header.value"
          :disabled="header.inactive"
          @update:model-value="(val: string | number) => updateHeader(index, 'value', String(val))"
        />
        <Button variant="ghost" size="icon" :disabled="header.inactive"><Trash /></Button>
      </div>
    </div>
    <!-- <TextArea class="w-full min-w-0 break-all resize-none h-full"></TextArea> -->
  </div>
</template>
