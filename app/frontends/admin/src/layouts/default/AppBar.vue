<template>
  <v-navigation-drawer expand-on-hover rail>
    <v-list>
      <v-list-item
        prepend-avatar="https://www.ardanlabs.com/images/ardanlabs-logo.svg"
      ></v-list-item>
    </v-list>

    <v-divider></v-divider>

    <v-list density="compact" nav>
      <v-list-item
        v-for="menu in availableMenus"
        :prepend-icon="menu.icon"
        :title="menu.title"
        :to="{ name: menu.name }"
        variant="text"
      ></v-list-item>
    </v-list>
  </v-navigation-drawer>
</template>

<script setup>
const availableMenus = [
  { title: "Users", name: "Users", icon: "fas fa-users" },
];
</script>
