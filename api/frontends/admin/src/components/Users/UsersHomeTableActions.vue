<template>
  <v-btn-toggle
    class="rounded-lg homes-table-actions elevation-2"
    borderless
    background-color="rgba(255, 255, 255)"
  >
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-btn
          class="text--dark ma-1"
          size="x-small"
          v-on="on"
          @click.stop.prevent="action('edit')"
        >
          <v-icon icon="fas fa-pencil"> </v-icon>
        </v-btn>
      </template>
      <span>Edit Home</span>
    </v-tooltip>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-btn
          class="ma-1"
          size="x-small"
          v-on="on"
          @click.stop.prevent="action('delete')"
        >
          <v-icon icon="fas fa-trash"></v-icon>
        </v-btn>
      </template>
      <span> Delete Home </span>
    </v-tooltip>
  </v-btn-toggle>
</template>

<script>
export default {
  name: "UserHomesDataTableActions",
  props: {
    item: {
      type: Object,
      default: () => ({}),
    },
  },
  methods: {
    action(type) {
      this.$emit(`${type}`);
    },
  },
};
</script>
