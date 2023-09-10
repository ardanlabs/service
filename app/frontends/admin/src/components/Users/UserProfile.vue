<template>
  <div class="ma-6 fill-height justify-start">
    <v-responsive class="align-start text-center fill-height">
      <v-card>
        <v-card-title class="d-flex align-center justify-space-between">
          <div>
            <v-btn
              icon="fas fa-arrow-left"
              flat
              @click="$router.push({ name: 'Users' })"
            />
            User Profile
          </div>
        </v-card-title>
        <v-card-text class="d-flex pa-0">
          <v-col cols="4">
            <user-details :user="user" />
          </v-col>
          <v-col cols="8">
            <user-homes :user-id="user.id" />
          </v-col>
        </v-card-text>
      </v-card>
    </v-responsive>
  </div>
</template>
<script>
import UserDetails from "../Users/UserDetails.vue";
import UserHomes from "../Users/UserHomes.vue";

export default {
  components: { UserDetails, UserHomes },
  props: {
    userId: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      error: null,
      user: {},
    };
  },
  mounted() {
    this.fetchUser(this.userId);
  },
  watch: {
    userId(newVal) {
      this.fetchUser(newVal);
    },
  },
  methods: {
    async fetchUser(userId) {
      try {
        const fetchCall = await fetch(
          `${import.meta.env.VITE_SERVICE_API}/users/${userId}`,
          {
            headers: {
              Authorization: `Bearer ${import.meta.env.VITE_SERVICE_TOKEN}`,
            },
          }
        );
        if (fetchCall.ok) {
          try {
            const fetchedData = await fetchCall.json();

            this.user = fetchedData;
          } catch (error) {
            console.log("Data to parse:", fetchCall);
            console.log("Users JSON Parse failed:", error);
          }
          return;
        }
      } catch (error) {
        console.log("User fetch failed:", error);
        this.error = error;
      }
    },
  },
};
</script>
