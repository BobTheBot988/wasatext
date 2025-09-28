<script>
import { RouterView } from "vue-router";
import NavBar from "./components/NavBar.vue";
import ErrorDisplay from "./components/ErrorDisplay.vue";

export default {
  components: {
    NavBar,
    ErrorDisplay,
  },
  data() {
    return {
      isAuthed: false,
      localStor: localStorage.getItem("token"),
      sessionStor: sessionStorage.getItem("token"),
      is_disabled: false,
      user: new Object(),
      group: null,
      users: [],
      onLine: false,
      show_online: false,
      paused: false,
      k: 0,
      n: 0,
    };
  },
  watch: {
    onLine(v) {
      if (v) {
        this.show_online = true;
        setTimeout(() => {
          this.show_online = false;
        }, 1000);
      }
    },
  },
  mounted() {
    this.isAuthenticated();
    if (this.isAuthed) {
      this.getUser();
    }
  },
  methods: {
    isAuthenticated() {
      if (this.localStor) this.isAuthed = true;
      else if (this.sessionStor) this.isAuthed = true;
      else this.isAuthed = false;
    },

    async goodLogin() {
      console.log("Login successful");
      this.isAuthed = true;
      await this.getUser();
      this.getUsers();
    },
    async getUsers() {
      try {
        console.log("Getting users");
        this.loading = true;
        const response = await this.$axios.get("/users", {
          headers: {
            Authorization: `Bearer ${this.token}`,
          },
        });
        this.users = response.data;

        for (const user of this.users) {
          user.groupId = 0;
          //user.photo = await this.getItemPhoto(user);
        }
        sessionStorage.setItem("users", JSON.stringify(this.users));
      } catch (error) {
        errorManager.addError(
          "Error while getting users for groupmodal",
          error.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    async getItemPhoto(item) {
      try {
        this.loading = true;
        let response = null;
        let url = null;
        if (item.id) {
          url = `/users/${item.id}/photo`;
        } else if (item.userId) {
          url = `/users/${item.userId}/photo`;
        }
        response = await this.$axios.get(url, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
          responseType: "blob",
        });

        return URL.createObjectURL(response.data);
      } catch (error) {
        console.log(error);
      } finally {
        this.loading = false;
      }
    },
    getUser() {
      if (localStorage.getItem("user")) {
        this.user = JSON.parse(localStorage.getItem("user"));
      } else if (sessionStorage.getItem("user")) {
        this.user = JSON.parse(sessionStorage.getItem("user"));
      } else {
        this.user = {
          userName: "Default",
          userId: -1,
        };
      }
    },
    updateUser(new_username) {
      this.user.userName = new_username;
      if (this.localStor) {
        localStorage.setItem("user", JSON.stringify(this.user));
      } else {
        sessionStorage.setItem("user", JSON.stringify(this.user));
      }
    },
    async updateUserImage() {
      this.user.photo = await this.getItemPhoto(this.user);
      if (this.localStor) {
        localStorage.setItem("user", JSON.stringify(this.user));
      } else {
        sessionStorage.setItem("user", JSON.stringify(this.user));
      }
    },
    refreshPage() {
      this.$forceUpdate();
    },

    handleLogOut() {
      if (localStorage.getItem("token") !== null) {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
      } else {
        sessionStorage.removeItem("token");
        sessionStorage.removeItem("user");
        sessionStorage.removeItem("convs");
        sessionStorage.removeItem("users");
        sessionStorage.removeItem("group");
      }
      this.isAuthed = false;

      this.$router.push("/session");
    },
    changeGroup(group) {
      this.group = group;
      console.log(this.group);
    },
  },
};
</script>

<template>
  <div style="padding: 1rem; height: 100%">
    <div id="main" class="grid" :class="{ disabled_div: is_disabled }">
      <ErrorDisplay />
      <NavBar
        id="nav"
        :key="n"
        :is-authed="isAuthed"
        :user="user"
        @turn-main="is_disabled = !is_disabled"
        @get-user="getUser"
        @log-out="handleLogOut"
        @pause="paused = true"
        @unpause="paused = false"
      />
      <RouterView
        :key="k"
        :paused="paused"
        :is-authed="isAuthed"
        :user="user"
        :group="group"
        @send-group-up="changeGroup"
        @change-username="updateUser"
        @change-user-image="updateUserImage"
        @handle-good-login="goodLogin"
        @pause="paused = true"
        @unpause="paused = false"
      />
    </div>
  </div>
</template>
