<template>
  <user-profile :user-id="id" />
</template>

<script setup>
import UserProfile from "@/components/Users/UserProfile.vue";

import { useRoute, useRouter } from "vue-router";
const route = useRoute();
const router = useRouter();

const id = route.params.id;

if (!id) {
  router.replace({ name: "Users" });
}
</script>
