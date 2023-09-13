<template>
  <div class="ma-6 fill-height justify-start">
    <v-responsive class="align-start text-center fill-height">
      <v-card>
        <v-card-title class="d-flex align-center justify-space-between">
          <div>Users</div>
          <v-btn @click="openNewUser">Add User</v-btn>
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
        @success="successEdit"
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
  </div>
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
    successDelete(userName) {
      this.dialogs.success.title = "Delete Success";
      this.dialogs.success.subtitle = `User ${userName} deleted successfully`;
      this.dialogs.success.open = true;
      if ("usersTable" in this.$refs) this.$refs.usersTable.loadItems();
    },
    successEdit() {
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
    openNewUser() {
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
    async openDelete(item) {
      this.dialogs.confirmation.item = item;
      await nextTick();
      this.dialogs.confirmation.open = true;
    },
    async sendDeleteUser() {
      const { id, name } = this.dialogs.confirmation.item;
      try {
        const fetchCall = await fetch(
          `${import.meta.env.VITE_SERVICE_API}/users/${id}`,
          {
            method: "DELETE",
            headers: {
              Accept: "application/json",
              "Content-type": "application/json",
              Authorization: `Bearer ${import.meta.env.VITE_SERVICE_TOKEN}`,
            },
          }
        );

        switch (fetchCall.status) {
          case 204:
            this.successDelete(name);
            this.dialogs.confirmation.item = {};
            break;
          default:
            let userDeleteData;
            try {
              userDeleteData = await fetchCall.json();
            } catch (error) {
              const errors = [
                { message: "Returned delete data couldn't be parsed" },
                { message: `Error: ${error}` },
              ];
              this.failure(errors);
            }
            const errors = [
              { message: "Deleting user went wrong" },
              { message: `Error Code: ${fetchCall.status}` },
              { message: userDeleteData.error },
            ];
            this.failure(errors);
            break;
        }
      } catch (error) {
        const errors = [
          { message: "Post call failed" },
          { message: `Error: ${error}` },
        ];
        this.failure(errors);
      }
    },
  },
};
</script>
