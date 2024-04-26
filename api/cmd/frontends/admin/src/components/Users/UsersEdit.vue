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
      <div class="d-flex flex-column justify-center px-3">
        <div class="text-body-1 text--light pt-7 pb-4">
          <v-form v-model="valid" ref="form">
            <v-text-field
              v-model="form.name"
              variant="outlined"
              label="Name"
              :rules="[requiredRule]"
            />
            <v-text-field
              v-model="form.email"
              variant="outlined"
              label="Email"
              :rules="[emailRule]"
            />
            <v-select
              v-model="form.roles"
              label="Role"
              multiple
              variant="outlined"
              :items="userRoles"
            />
            <v-text-field
              v-model="form.department"
              variant="outlined"
              label="Department"
            />
            <v-text-field
              v-if="!this.edit"
              v-model="form.password"
              :append-inner-icon="visible ? 'fas fa-eye' : 'fas fa-eye-slash'"
              :type="visible ? 'text' : 'password'"
              placeholder="Enter your password"
              variant="outlined"
              :rules="[requiredRule]"
              @click:append-inner="visible = !visible"
            />
            <v-text-field
              v-if="!this.edit"
              v-model="form.passwordConfirm"
              :append-inner-icon="
                visibleConfirm ? 'fas fa-eye' : 'fas fa-eye-slash'
              "
              :type="visibleConfirm ? 'text' : 'password'"
              variant="outlined"
              label="Confirm Password"
              :rules="[passwordRule]"
              @click:append-inner="visibleConfirm = !visibleConfirm"
            />
          </v-form>
        </div>
      </div>
    </template>
    <template #actions>
      <div class="d-flex flex-grow-1 justify-end">
        <v-btn class="align-self-right" text @click="closeDialog">
          Cancel
        </v-btn>

        <v-btn class="align-self-right ml-4" color="primary" @click="editUser">
          {{ dialogButtonText }}
        </v-btn>
      </div>
    </template>
  </ui-dialog>
</template>
<script>
import UiDialog from "../UI/dialog.vue";

export default {
  name: "UsersEdit",
  components: { UiDialog },
  props: {
    user: {
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
      visible: false,
      visibleConfirm: false,
      valid: false,
      form: {
        name: "",
        email: "",
        roles: [],
        department: "",
        password: "",
        passwordConfirm: "",
      },
    };
  },
  beforeMount() {
    if (this.edit && Object.keys(this.user).length) {
      this.form = Object.assign({}, this.user);
    }
  },
  watch: {
    user: {
      handler() {
        if (Object.keys(this.user).length) {
          this.form = Object.assign({}, this.user);
        }
      },
      deep: true,
    },
  },
  computed: {
    dialogTitle() {
      return this.edit ? "Edit User" : "Add User";
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
    userRoles() {
      return [
        { title: "Admin", value: "ADMIN" },
        { title: "User", value: "USER" },
      ];
    },
  },
  methods: {
    passwordRule(v) {
      if (!v) {
        return false;
      }
      return (
        this.form.password === this.form.passwordConfirm ||
        "Passwords don't match"
      );
    },
    requiredRule(v) {
      return !!v || "This field is required";
    },
    emailRule(v) {
      if (!v) {
        return false;
      }
      if (v.length <= 6 || v.length >= 128) {
        return false;
      }
      const emailRegExp =
        /^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/;
      return emailRegExp.test(v) || "Invalid Email Format";
    },
    async editUser() {
      if ("form" in this.$refs) {
        await this.$refs.form.validate();
      }

      if (this.valid) {
        let url = `${import.meta.env.VITE_SERVICE_API}/users`;

        if (this.edit) {
          url += `/${this.form.id}`;
          delete this.form.id;
          delete this.form.dateCreated;
          delete this.form.dateUpdated;
          delete this.form.enabled;
        }

        try {
          const fetchCall = await fetch(url, {
            method: this.edit ? "PUT" : "POST",
            headers: {
              Accept: "application/json",
              "Content-type": "application/json",
              Authorization: `Bearer ${import.meta.env.VITE_SERVICE_TOKEN}`,
            },
            body: JSON.stringify(this.form),
          });
          let userPostData;

          try {
            userPostData = await fetchCall.json();
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
                name: "",
                email: "",
                roles: [],
                department: "",
                password: "",
                passwordConfirm: "",
              };
              break;
            default:
              const errors = [
                { message: "Creating user went wrong" },
                { message: `Error Code: ${fetchCall.status}` },
                { message: userPostData.error },
              ];
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
