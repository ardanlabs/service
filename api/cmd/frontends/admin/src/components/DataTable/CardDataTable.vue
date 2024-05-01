<template>
  <v-card :flat="flat" :elevation="3">
    <v-card-title :hide-title="hideTitle">
      <span>{{ title }}</span>
    </v-card-title>
    <v-card-text>
      <data-table-server v-bind="$attrs">
        <template v-for="(_, slot) of $slots" #[slot]="scope">
          <slot :name="slot" v-bind="scope" />
        </template>
      </data-table-server>
    </v-card-text>
  </v-card>
</template>

<script>
import DataTableServer from "../DataTable/DataTableServer.vue";

export default {
  name: "UiDataTable",
  inheritAttrs: false,
  components: { DataTableServer },
  props: {
    title: {
      type: String,
      default: "",
    },
    tableHeight: {
      type: String,
      default: "500px",
    },
    flat: {
      type: Boolean,
      default: false,
    },
    itemClass: {
      type: Function,
      default: () => {},
    },
    hideTitle: {
      type: Boolean,
      default: false,
    },
    showBorder: {
      type: Boolean,
      default: true,
    },
    isFirstColumnFixed: {
      type: Boolean,
      default: false,
    },
    hasActions: {
      type: Boolean,
      default: false,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    headersLength: {
      type: Number,
      default: 6,
    },
  },
  computed: {
    classes() {
      return {
        "ui-data-table": true,
        "ui-data-table--fixed": this.isFirstColumnFixed,
        "ui-data-table--actions": this.hasActions,
      };
    },
  },
};
</script>

<style lang="scss">
@import "@/styles/variables.scss";
.ui-data-table--actions thead.v-data-table-header > tr > th:last-child > span {
  opacity: 0;
}
.ui-data-table.v-data-table {
  table {
    cursor: pointer;
    thead.v-data-table__thead th {
      font-size: $data-table-header-font-size;
      font-weight: $data-table-header-font-weight;
      text-transform: uppercase;
      color: $theme-light-text-dark !important;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
      border-bottom: none !important;
    }
    td {
      border-bottom: none !important;
      font-size: $data-table-body-font-size;
      font-weight: $data-table-body-font-weight;
      color: $theme-light-text-dark !important;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }

    tbody > tr:nth-child(odd) td {
      background: $theme-light-table-row-odd-color;
      border-bottom: thin solid rgba(0, 0, 0, 0.12);
      span:not([class*="status"]) {
        color: #0c466b !important;
      }
    }
    tbody > tr:nth-child(even) td {
      background: $theme-light-table-row-even-color;
      border-bottom: thin solid rgba(0, 0, 0, 0.12);
      span:not([class*="status"]) {
        color: #0c466b !important;
      }
    }
  }
}
.ui-data-table--actions {
  tbody > tr > td:last-child,
  thead.v-data-table__thead > tr > th:last-child {
    display: none;
    transition: opacity 0.5s ease-in-out;
    border: none !important;
    right: 0;
    position: sticky;
  }
  tbody > tr > td:last-child {
    opacity: 0;
    display: none;
    background: red;
  }

  tbody > tr:hover {
    td:last-child,
    th:last-child {
      transition: opacity 0.5s ease-in-out;
      opacity: 1;
      display: block;
    }
  }
}
.ui-data-table--fixed {
  table > tbody > tr:nth-child(odd) td:nth-child(1) {
    background: $theme-light-table-row-odd-color;
  }
  table > tbody > tr:hover > td:nth-child(1) {
    background: $theme-light-table-row-hover-color;
  }
}
.ui-data-table--fixed {
  table > tbody > tr:hover td {
    background: $theme-light-table-row-hover-color;
  }
  div table > tbody > tr > td:nth-child(1),
  div table > thead.v-data-table__thead > tr th:nth-child(1) {
    position: sticky;
    left: 0;
    background: $theme-light-table-row-even-color;
    z-index: 1;
  }
}
table thead .v-progress-linear {
  z-index: 2;
}
</style>
