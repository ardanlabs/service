<template>
  <v-card
    show-title
    style="height: 100%"
    title-padding="pb-0 pt-4 pl-7 pr-4"
    class="pb-4"
  >
    <v-card-title>
      <v-row no-gutters>
        <v-col cols="6" class="justify-start align-center d-flex">
          Homes
        </v-col>
        <v-col cols="6" class="justify-end align-center d-flex">
          <v-btn outlined @click="openNewHome">Add Home</v-btn>
        </v-col>
      </v-row>
    </v-card-title>
    <v-card-text>
      <v-row no-gutters class="d-flex justify-space-around">
        <v-col cols="12" class="justify-start">
          <user-homes-table
            ref="homesTable"
            :user-id="userId"
            @delete="openDelete($event)"
            @edit="openEdit($event)"
          />
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
  <home-edit
    v-model="dialogs.edit.open"
    :user-id="userId"
    :edit="dialogs.edit.edit"
    :home="dialogs.edit.item"
    @close="dialogs.edit.open = false"
    @error="failure"
    @success="successEdit"
  />
  <ui-confirmation-dialog
    v-model="dialogs.confirmation.open"
    :title="dialogs.confirmation.title"
    :text="dialogs.confirmation.text"
    :button-text="dialogs.confirmation.buttonText"
    @confirm="sendDeleteHome"
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
</template>
<script>
import UserHomesTable from "../Users/UserHomesTable";
import HomeEdit from "../Users/HomeEdit";

import UiConfirmationDialog from "../UI/confirmation-dialog.vue";
import UiSuccessDialog from "../UI/success-dialog.vue";
import UiFailureDialog from "../UI/failure-dialog.vue";
import { nextTick } from "vue";

export default {
  name: "UserHomes",
  components: {
    UserHomesTable,
    HomeEdit,
    UiConfirmationDialog,
    UiSuccessDialog,
    UiFailureDialog,
  },
  props: {
    userId: {
      type: String,
      default: "",
      required: true,
    },
  },
  data() {
    return {
      dialogs: {
        edit: {
          open: false,
          edit: false,
          item: {},
        },
        confirmation: {
          open: false,
          title: "Delete Home",
          text: "Do you want to delete this home?",
          buttonText: "Delete",
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
    openNewHome() {
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
    async sendDeleteHome() {
      const { id } = this.dialogs.confirmation.item;
      try {
        const fetchCall = await fetch(
          `${import.meta.env.VITE_SERVICE_API}/homes/${id}`,
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
            this.successDelete();
            this.dialogs.confirmation.item = {};
            break;
          default:
            let homeDeleteData;
            try {
              homeDeleteData = await fetchCall.json();
            } catch (error) {
              const errors = [
                { message: "Returned delete data couldn't be parsed" },
                { message: `Error: ${error}` },
              ];
              this.failure(errors);
            }
            const errors = [
              { message: "Deleting home went wrong" },
              { message: `Error Code: ${fetchCall.status}` },
              { message: homeDeleteData.error },
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
    async openDelete(item) {
      this.dialogs.confirmation.item = item;
      await nextTick();
      this.dialogs.confirmation.open = true;
    },
    successDelete() {
      this.dialogs.success.title = "Delete Success";
      this.dialogs.success.subtitle = `Home deleted successfully`;
      this.dialogs.success.open = true;
      if ("homesTable" in this.$refs) this.$refs.homesTable.loadItems();
    },
    successEdit() {
      const action = this.dialogs.edit.edit ? "edited" : "added";
      this.dialogs.success.title = "Success";
      this.dialogs.success.subtitle = `Home ${action} successfully`;
      this.dialogs.success.open = true;
      this.dialogs.edit.open = false;
      this.dialogs.edit.edit = false;
      this.dialogs.edit.item = {};
      if ("homesTable" in this.$refs) this.$refs.homesTable.loadItems();
    },
    failure(errors) {
      this.dialogs.failure.errors = errors;
      this.dialogs.failure.title = "Something went wrong";
      this.dialogs.failure.subtitle = "Creating user went wrong";
      this.dialogs.failure.open = true;
    },
  },
};
</script>
