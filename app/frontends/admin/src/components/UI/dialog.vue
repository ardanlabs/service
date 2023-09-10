<template>
  <v-dialog
    :scrollable="scrollable"
    :content-class="contentClass"
    :persistent="persistent"
    v-bind="$attrs"
  >
    <v-card>
      <template v-if="!noTitle">
        <v-card-title
          class="d-flex justify-space-between align-center text-body-1 text--dark px-6 py-3"
        >
          <div class="flex-grow-1">
            <slot name="title" />
          </div>
          <v-btn v-if="showCloseButton" flat icon @click="$emit('close')">
            <v-icon>$clear</v-icon>
          </v-btn>
        </v-card-title>
        <v-divider />
      </template>
      <v-card-text class="py-6">
        <slot name="body" />
      </v-card-text>
      <v-divider />
      <v-card-actions v-if="showActions">
        <slot name="actions" />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "UiDialog",
  props: {
    noTitle: {
      type: Boolean,
      default: false,
    },
    scrollable: {
      type: Boolean,
      default: false,
    },
    showActions: {
      type: Boolean,
      default: true,
    },
    showCloseButton: {
      type: Boolean,
      default: true,
    },
    contentClass: {
      type: String,
      default: "ui-dialog",
    },
    persistent: {
      type: Boolean,
      default: false,
    },
  },
};
</script>

<style lang="scss">
.ui-dialog {
  background: #fff;
  max-width: 475px !important;

  .v-theme--light.v-card {
    background: none !important;
  }
}
.ui-dialog--confirmation {
  max-width: 445px !important;
}
</style>
