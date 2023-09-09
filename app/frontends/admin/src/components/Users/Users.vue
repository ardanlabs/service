<template>
  <v-container class="fill-height">
    <v-responsive class="align-start text-center fill-height">
      <v-card>
        <v-card-title class="d-flex align-center justify-space-between">
          <div class="v-card-title">Users</div>
          <v-btn @click="openUser">Add User</v-btn>
        </v-card-title>
        <v-card-text>
          <users-table
            ref="usersTable"
            @delete="openDelete($event)"
            @edit="openEdit($event)"
          />
        </v-card-text>
      </v-card>
      <users-edit
        v-model="dialogs.edit.open"
        :edit="dialogs.edit.edit"
        :user="dialogs.edit.item"
        @close="dialogs.edit.open = false"
        @error="failure"
        @success="success"
      />
      <ui-confirmation-dialog
        v-model="dialogs.confirmation.open"
        :title="dialogs.confirmation.title"
        :text="dialogs.confirmation.text"
        :button-text="dialogs.confirmation.buttonText"
        @confirm="sendDeleteUser"
        @input="dialogs.confirmation.open = false"
      />
      <ui-success-dialog
        v-model="dialogs.success.open"
        :title="dialogs.success.title"
        :subtitle="dialogs.success.subtitle"
        @close="dialogs.success.open = false"
      />
      <ui-failure-dialog
        v-model="dialogs.failure.open"
        :title="dialogs.failure.title"
        :subtitle="dialogs.failure.subtitle"
        :errors="dialogs.failure.errors"
        @close="dialogs.failure.open = false"
      />
    </v-responsive>
  </v-container>
</template>

<script>
import UsersTable from "./UsersTable.vue";
import UsersEdit from "./UsersEdit.vue";

import UiConfirmationDialog from "../UI/confirmation-dialog.vue";
import UiSuccessDialog from "../UI/success-dialog.vue";
import UiFailureDialog from "../UI/failure-dialog.vue";
import { nextTick } from "vue";

export default {
  components: {
    UsersTable,
    UsersEdit,
    UiConfirmationDialog,
    UiSuccessDialog,
    UiFailureDialog,
  },
  data() {
    return {
      dialogs: {
        confirmation: {
          open: false,
          title: "Delete User",
          text: "Do you want to delete this user?",
          buttonText: "Delete",
          item: {},
        },
        edit: {
          open: false,
          edit: false,
          item: {},
        },
        success: {
          open: false,
          title: "",
          subtitle: "",
        },
        failure: {
          open: false,
          title: "",
          subtitle: "",
          errors: [],
        },
      },
    };
  },
  methods: {
    success() {
      const action = this.dialogs.edit.edit ? "edited" : "added";
      this.dialogs.success.title = "Success";
      this.dialogs.success.subtitle = `User ${action} successfully`;
      this.dialogs.success.open = true;
      this.dialogs.edit.open = false;
      this.dialogs.edit.edit = false;
      this.dialogs.edit.item = {};
      if ("usersTable" in this.$refs) this.$refs.usersTable.loadItems();
    },
    failure(errors) {
      this.dialogs.failure.errors = errors;
      this.dialogs.failure.title = "Something went wrong";
      this.dialogs.failure.subtitle = "Creating user went wrong";
      this.dialogs.failure.open = true;
    },
    openUser() {
      this.dialogs.edit.open = true;
      this.dialogs.edit.edit = false;
    },
    async openEdit(item) {
      this.dialogs.edit.edit = true;
      await nextTick();
      this.dialogs.edit.item = item;
      await nextTick();
      this.dialogs.edit.open = true;
    },
    sendDeleteUser() {
      console.log(this.dialogs.confirmation.item.id);
    },
    openDelete(item) {
      console.log(item);
      this.dialogs.confirmation.item = item;
      this.dialogs.confirmation.open = true;
    },
  },
};
</script>
