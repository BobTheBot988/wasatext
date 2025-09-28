import { createRouter, createWebHashHistory } from "vue-router";
import LoginView from "../views/LoginView.vue";
import ConversationView from "../views/ConversationView.vue";
import UserView from "../views/UserView.vue";
import GroupView from "../views/GroupView.vue";
import NotFoundView from "../views/NotFoundView.vue";
import { errorManager } from "../services/axios";

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/conversations",
      name: "myConvs",
      component: ConversationView,
      meta: {
        requiresAuth: true, // Add meta field to indicate protected route
      },
    },
    {
      path: "/user",
      name: "myUser",
      component: UserView,
      meta: {
        requiresAuth: true, // Add meta field to indicate protected route
      },
    },
    {
      path: "/group",
      name: "currentGroup",
      component: GroupView,
      meta: {
        requiresAuth: true, // Add meta field to indicate protected route
      },
    },
    {
      path: "/session",
      alias: "/login",
      component: LoginView,
    },
    {
      path: "/:pathmatch(.*)*",
      name: "NotFound",
      component: NotFoundView,
    },
  ],
});
router.beforeEach((to, from, next) => {
  const localStor = localStorage.getItem("token");
  const sessionStor = sessionStorage.getItem("token");
  // errorManager.clearErrors()
  // Handle 404 route specifically
  if (to.name === "NotFound") {
    // If user is authenticated, show 404 page; otherwise redirect to login
    if (localStor || sessionStor) {
      next();
    } else {
      next("/session");
    }
    return;
  }

  switch (to.path) {
    case "/session":
      if (localStor || sessionStor) {
        next("/conversations");
      } else {
        next();
      }
      break;
    case "/":
      next("/conversations");
      break;

    default:
      if (to.meta.requiresAuth) {
        if (localStor || sessionStor) {
          // User is authenticated, proceed to the route
          next();
        } else {
          // User is not authenticated, redirect to login
          errorManager.addError("Please log in to access this page");
          next("/session");
        }
      } else {
        // Non-protected route, allow access
        next();
      }
      break;
  }
});

export default router;
