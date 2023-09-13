<template>
  <v-dialog
    v-bind="$attrs"
    content-class="ui-success-dialog white"
    width="410"
    v-on="$listeners"
    @click:outside="close"
  >
    <v-card>
      <v-card-title class="d-flex justify-end">
        <v-btn flat icon @click="close">
          <v-icon>$clear</v-icon>
        </v-btn>
      </v-card-title>
      <v-card-text class="pa-0">
        <div class="d-flex flex-column align-center justify-center">
          <v-img src="../../../public/error.svg" class="mb-2" width="150px" />
          <h1 class="text-h6 text-center text--dark">
            {{ title }}
          </h1>
          <h2
            v-if="subtitle"
            class="text-center text-body-1 text--light px-8 pt-3 pb-10"
          >
            {{ subtitle }}
            <slot name="subtitle" />
          </h2>
          <h2
            v-if="errors.length"
            class="text-center text-body-2 text--light px-8 pt-3 pb-10"
            width="410"
          >
            <ul>
              <li v-for="error in errors" :key="error.message">
                {{ error.message }}
              </li>
            </ul>
          </h2>
        </div>
        <v-divider />
        <v-row no-gutters>
          <v-col class="pa-2 d-flex justify-end">
            <v-btn color="primary" @click="close"> Got it </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "UiFailureDialog",
  props: {
    title: {
      type: String,
      default: "",
      required: true,
    },
    subtitle: {
      type: String,
      default: "",
    },
    errors: {
      type: Array,
      default: () => [],
    },
  },
  methods: {
    close() {
      this.$emit("close", false);
      this.$emit("input", false);
    },
  },
};
</script>
