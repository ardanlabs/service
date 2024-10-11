<template>
  <v-skeleton-loader
    v-if="loading"
    :height="tableHeight"
    :type="`table-tbody@${headersLength}`"
  />
  <v-data-table-server v-else v-bind="$attrs" :class="classes">
    <template v-for="(_, slot) of $slots" #[slot]="scope">
      <slot :name="slot" v-bind="scope" />
    </template>
  </v-data-table-server>
</template>
<script>
export default {
  props: {
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
    tableHeight: {
      type: String,
      default: "500px",
    },
    isFirstColumnFixed: {
      type: Boolean,
      default: false,
    },
    itemClass: {
      type: Function,
      default: () => {},
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
.ui-data-table.v-data-table table tbody > tr:nth-child(odd) td:last-child > * {
  background-color: #fff;
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
    background: $theme-light-table-row-even-color;
    z-index: 1;
  }
}
table thead .v-progress-linear {
  z-index: 2;
}
</style>
