<script setup lang="ts">
import type { Component } from 'vue'

import { ChevronsUpDown, Plus } from 'lucide-vue-next'
import { ref } from 'vue'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  teams: {
    name: string
    logo: Component
    plan: string
  }[]
}>()

const activeTeam = ref(props.teams[0] || { name: '', logo: null, plan: '' })
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" class="w-full h-12 !px-2 justify-start gap-2">
        <div
          class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground"
        >
          <component :is="activeTeam.logo" class="size-4" />
        </div>
        <div class="grid flex-1 text-left text-sm leading-tight">
          <span class="truncate font-medium">
            {{ activeTeam.name }}
          </span>
          <span class="truncate text-xs text-muted-foreground">{{ activeTeam.plan }}</span>
        </div>
        <ChevronsUpDown class="ml-auto size-4" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent
      class="w-[--reka-dropdown-menu-trigger-width] min-w-56 rounded-lg"
      align="start"
      :side="'right'"
      :side-offset="4"
    >
      <DropdownMenuLabel class="text-xs text-muted-foreground"> Teams </DropdownMenuLabel>
      <DropdownMenuItem
        v-for="(team, index) in teams"
        :key="team.name"
        class="gap-2 p-2"
        @click="activeTeam = team"
      >
        <div class="flex size-6 items-center justify-center rounded-sm border">
          <component :is="team.logo" class="size-3.5 shrink-0" />
        </div>
        {{ team.name }}
        <DropdownMenuShortcut>⌘{{ index + 1 }}</DropdownMenuShortcut>
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem class="gap-2 p-2">
        <div class="flex size-6 items-center justify-center rounded-md border bg-transparent">
          <Plus class="size-4" />
        </div>
        <div class="font-medium text-muted-foreground">Add team</div>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
