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
          Details
        </v-col>
      </v-row>
    </v-card-title>
    <v-card-text>
      <v-row no-gutters class="d-flex">
        <v-col cols="12" class="d-flex">
          <v-col lg="4" md="6" class="text-start font-weight-medium">
            Name
          </v-col>
          <v-col lg="8" md="6" class="text-dark text-right">
            {{ user.name }}
          </v-col>
        </v-col>
        <v-col cols="12" class="d-flex">
          <v-col lg="4" md="6" class="text-start font-weight-medium">
            Email
          </v-col>
          <v-col lg="8" md="6" class="text-dark text-right">
            {{ user.email }}
          </v-col>
        </v-col>
        <v-col cols="12" class="d-flex">
          <v-col lg="4" md="6" class="text-start font-weight-medium">
            Department
          </v-col>
          <v-col lg="8" md="6" class="text-dark text-right">
            {{ user.department }}
          </v-col>
        </v-col>
        <v-col cols="12" class="d-flex" v-if="user.roles">
          <v-col lg="4" md="6" class="text-start font-weight-medium">
            Roles
          </v-col>
          <v-col lg="8" md="6" class="text-dark text-right">
            {{ user.roles.join(", ") }}
          </v-col>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup>
defineProps(["user"]);
</script>
