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
    @update:options="loadItems"
  >
    <template v-slot:[`item.address.country`]="{ item }">
      <div>{{ getCountry(item.columns["address.country"]) }}</div>
    </template>
    <template #[`item.actions`]="{ item }">
      <users-table-actions
        @delete="$emit('delete', item.selectable)"
        @edit="$emit('edit', item.selectable)"
        :item="item"
      />
    </template>
  </data-table-server>
</template>
<script>
import DataTableServer from "../DataTable/DataTableServer.vue";
import { UserHomesTableHeaders } from "../Users/Users.js";
import UsersTableActions from "../Users/UsersTableActions.vue";
import SortQuery from "../DataTable/SortQuery";
import Countries from "../Users/Countries.js";

export default {
  name: "UserHomesTable",
  components: {
    DataTableServer,
    UsersTableActions,
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
      tableOptions: {
        page: 1,
        itemsPerPage: 3,
        sortBy: [],
      },
      error: {},
      users: [],
      serverItemsLength: 0,
      usersItemsPerPageOptions: [
        { title: "1", value: 1 },
        { title: "2", value: 2 },
        { title: "3", value: 3 },
      ],
    };
  },
  computed: {
    countries() {
      return Countries;
    },
    headers() {
      return UserHomesTableHeaders;
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
    getCountry(code) {
      return this.countries.filter((e) => e.value === code)[0].title;
    },
    sortQuery(s) {
      return SortQuery(s);
    },
    async loadItems() {
      const { page, itemsPerPage, sortBy } = this.tableOptions;

      const sort = this.sortQuery(sortBy);

      try {
        const fetchCall = await fetch(
          `${import.meta.env.VITE_SERVICE_API}/homes?user_id=${
            this.userId
          }&page=${page}&rows=${itemsPerPage}${sort ? sort : ""}`,
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
