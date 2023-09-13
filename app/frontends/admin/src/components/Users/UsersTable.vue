<template>
  <data-table-server
    v-model:items-per-page="tableOptions.itemsPerPage"
    v-model:sort-by="tableOptions.sortBy"
    v-model:page="tableOptions.page"
    :headers="headers"
    :items-length="serverItemsLength"
    :items="users"
    :loading="loading"
    :items-per-page-options="usersItemsPerPageOptions"
    :show-select="false"
    is-first-column-fixed
    hide-title
    has-actions
    class="elevation-1"
    @click:row="goToClientsProfile"
    @update:options="loadItems"
  >
    <template v-slot:[`item.user_id`]="{ item }">
      <div>{{ item.value }}</div>
    </template>
    <template v-slot:[`item.roles`]="{ item }">
      <div>{{ item.columns.roles.join(", ") }}</div>
    </template>
    <template v-slot:[`item.dateCreated`]="{ item }">
      <div>{{ item.columns.dateCreated.substring(0, 10) }}</div>
    </template>
    <template v-slot:[`item.dateUpdated`]="{ item }">
      <div>{{ item.columns.dateUpdated.substring(0, 10) }}</div>
    </template>
    <template #[`item.actions`]="{ item }">
      <users-table-actions
        @delete="$emit('delete', item.selectable)"
        @edit="$emit('edit', item.selectable)"
        @profile="goToClientsProfile(item.selectable.id)"
        :item="item"
      />
    </template>
  </data-table-server>
</template>
<script>
import DataTableServer from "../DataTable/DataTableServer.vue";
import { UsersTableHeaders } from "../Users/Users.js";
import UsersTableActions from "../Users/UsersTableActions.vue";
import SortQuery from "../DataTable/SortQuery";

export default {
  name: "UsersTable",
  components: {
    DataTableServer,
    UsersTableActions,
  },
  data() {
    return {
      tableOptions: {
        page: 1,
        itemsPerPage: 5,
        sortBy: [],
      },
      error: {},
      users: [],
      serverItemsLength: 0,
      usersItemsPerPageOptions: [
        { title: "5", value: 5 },
        { title: "10", value: "10" },
        { title: "20", value: "20" },
      ],
    };
  },
  computed: {
    headers() {
      return UsersTableHeaders;
    },
    dataTableProps() {
      return {
        headers: this.headers,
        items: this.users,
        footerProps: this.usersItemsPerPageOptions,
        serverItemsLength: this.serverItemsLength,
        hasActions: true,
        loading: false,
        ...this.$attrs,
      };
    },
  },
  methods: {
    sortQuery(s) {
      return SortQuery(s);
    },
    goToClientsProfile(id, e) {
      let userId = id;
      if (typeof userId !== "string") {
        userId = e.item.selectable.id;
      }
      this.$router.push({
        name: "UserProfile",
        params: { id: userId },
      });
    },
    async loadItems() {
      const { page, itemsPerPage, sortBy } = this.tableOptions;

      const sort = this.sortQuery(sortBy);

      try {
        const fetchCall = await fetch(
          `${
            import.meta.env.VITE_SERVICE_API
          }/users?page=${page}&rows=${itemsPerPage}${sort ? sort : ""}`,
          {
            headers: {
              Authorization: `Bearer ${import.meta.env.VITE_SERVICE_TOKEN}`,
            },
          }
        );
        if (fetchCall.ok) {
          try {
            const fetchedData = await fetchCall.json();

            this.serverItemsLength = fetchedData.total;
            this.users = fetchedData.items;
          } catch (error) {
            console.log("Data to parse:", fetchCall);
            console.log("Users JSON Parse failed:", error);
          }
          return;
        }
      } catch (error) {
        console.log("fetchedCall failed:", error);
        this.error = error;
      }
    },
  },
};
</script>
