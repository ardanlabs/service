<template>
  <ui-dialog
    v-bind="dialogProps"
    @click:outside="closeDialog"
    @close="closeDialog"
  >
    <template #title>
      <div>{{ dialogTitle }}</div>
    </template>
    <template #body>
      <div class="d-flex flex-column justify-center">
        <div class="text-body-1 text--light pt-7 pb-4">
          <v-form v-model="valid" ref="form">
            <v-row no-gutters>
              <v-col cols="12" class="px-2">
                <v-text-field
                  v-model="form.type"
                  variant="outlined"
                  label="Type"
                  :rules="[requiredRule]"
                />
              </v-col>
              <v-col cols="12" class="px-2">
                <v-text-field
                  v-model="form.address.address1"
                  variant="outlined"
                  label="Address Line 1"
                  :rules="[requiredRule]"
                />
              </v-col>
              <v-col md="6" sm="12" class="px-2">
                <v-text-field
                  v-model="form.address.address2"
                  variant="outlined"
                  label="Address Line 2"
                />
              </v-col>
              <v-col md="6" sm="12" class="px-2">
                <v-text-field
                  v-model="form.address.zipCode"
                  variant="outlined"
                  label="ZIP Code"
                  :rules="[requiredRule]"
                />
              </v-col>
              <v-col cols="12" class="px-2">
                <v-autocomplete
                  v-model="form.address.country"
                  class="ui-text-field"
                  variant="outlined"
                  :items="countries"
                  item-value="value"
                  item-text="title"
                  clearable
                  :rules="[requiredRule]"
                  label="Country"
                />
              </v-col>
              <v-col md="6" sm="12" class="px-2">
                <v-text-field
                  v-model="form.address.state"
                  variant="outlined"
                  label="State"
                  :rules="[requiredRule]"
                />
              </v-col>
              <v-col md="6" sm="12" class="px-2">
                <v-text-field
                  v-model="form.address.city"
                  variant="outlined"
                  label="City"
                  :rules="[requiredRule]"
                />
              </v-col>
            </v-row>
          </v-form>
        </div>
      </div>
    </template>
    <template #actions>
      <div class="d-flex flex-grow-1 justify-end">
        <v-btn class="align-self-right" text @click="closeDialog">
          Cancel
        </v-btn>

        <v-btn class="align-self-right ml-4" color="primary" @click="editHome">
          {{ dialogButtonText }}
        </v-btn>
      </div>
    </template>
  </ui-dialog>
</template>
<script>
import Countries from "../Users/Countries.js";
import UiDialog from "../UI/dialog.vue";

export default {
  name: "HomeEdit",
  components: { UiDialog },
  props: {
    userId: {
      type: String,
      default: "",
    },
    home: {
      type: Object,
      default: () => {},
    },
    edit: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      valid: false,
      form: {
        type: "",
        address: {
          address1: "",
          address2: "",
          zipCode: "",
          city: "",
          state: "",
          country: "",
        },
      },
    };
  },
  beforeMount() {
    if (this.edit) {
      this.form = this.home;
    }
  },
  watch: {
    home: {
      handler() {
        if (Object.keys(this.home).length) {
          this.form = Object.assign({}, this.home);
        }
      },
      deep: true,
    },
  },
  computed: {
    dialogTitle() {
      return this.edit ? "Edit Home" : "Add Home";
    },
    dialogButtonText() {
      return this.edit ? "Edit" : "Add";
    },
    dialogProps() {
      return {
        scrollable: true,
        ...this.$attrs,
      };
    },
    countries() {
      return Countries;
    },
  },
  methods: {
    requiredRule(v) {
      return !!v || "This field is required";
    },
    async editHome() {
      if ("form" in this.$refs) {
        await this.$refs.form.validate();
      }

      if (this.valid) {
        let url = `${import.meta.env.VITE_SERVICE_API}/homes`;

        if (this.edit) {
          url += `/${this.form.id}`;
          delete this.form.id;
          delete this.form.userID;
          delete this.form.dateCreated;
          delete this.form.dateUpdated;
        }

        let nh = this.form;

        if (!this.edit) {
          nh.userID = this.userId;
        }
        nh = JSON.stringify(nh);

        try {
          const fetchCall = await fetch(url, {
            method: this.edit ? "PUT" : "POST",
            headers: {
              Accept: "application/json",
              "Content-type": "application/json",
              Authorization: `Bearer ${import.meta.env.VITE_SERVICE_TOKEN}`,
            },
            body: nh,
          });
          let homePostData;

          try {
            homePostData = await fetchCall.json();
          } catch (error) {
            const errors = [
              { message: "Returned post data couldn't be parsed" },
              { message: `Error: ${error}` },
            ];
            this.$emit("error", errors);
          }

          switch (fetchCall.status) {
            case 200:
            case 201:
              this.$emit("success");
              this.form = {
                type: "",
                address: {
                  address1: "",
                  address2: "",
                  zipCode: "",
                  city: "",
                  state: "",
                  country: "",
                },
              };
              break;
            default:
              const errors = [
                { message: "Creating home went wrong" },
                { message: `Error Code: ${fetchCall.status}` },
                { message: homePostData.error },
              ];
              if (homePostData.fields) {
                for (
                  let i = 0;
                  i < Object.values(homePostData.fields).length;
                  i++
                ) {
                  errors.push({
                    message: Object.values(homePostData.fields)[i],
                  });
                }
              }
              this.$emit("error", errors);
              break;
          }
        } catch (error) {
          const errors = [
            { message: "Post call failed" },
            { message: `Error: ${error}` },
          ];
          this.$emit("error", errors);
        }
      }
    },
    closeDialog() {
      this.$emit("close");
    },
    success() {
      this.$emit("success");
      this.closeDialog();
    },
  },
};
</script>
<style lang="scss" scoped>
.text--caption {
  font-size: 20px;
  font-weight: 500;
}
</style>
