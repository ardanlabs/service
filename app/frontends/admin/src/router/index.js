// Composables
import { createRouter, createWebHistory } from "vue-router";

const routes = [
  {
    path: "/",
    component: () => import("@/layouts/default/Default.vue"),
    meta: { transition: "slide-right" },
    children: [
      {
        path: "",
        name: "Users",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(/* webpackChunkName: "users" */ "@/views/Users.vue"),
      },
    ],
  },
  {
    path: "/user-profile/:id",
    component: () => import("@/layouts/default/Default.vue"),
    meta: { transition: "slide-right" },
    children: [
      {
        path: "",
        name: "UserProfile",
        // route level code-splitting
        // this generates a separate chunk (about.[hash].js) for this route
        // which is lazy-loaded when the route is visited.
        component: () =>
          import(
            /* webpackChunkName: "userProfile" */ "@/views/UserProfile.vue"
          ),
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router;
