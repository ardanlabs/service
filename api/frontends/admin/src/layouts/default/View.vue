<template>
  <v-main>
    <router-view v-slot="{ Component }">
      <v-slide-x-transition>
        <component :is="Component" />
      </v-slide-x-transition>
    </router-view>
  </v-main>
</template>

<script setup>
//
</script>
