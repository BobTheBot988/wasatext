<script>
import { errorManager } from "../services/axios";
export default {
  props: ["isAuthed"],
  emits: ["handle-good-login", "un-successful-login"],
  data: function () {
    return {
      username: "",
      wantLegacy: false,
      errormsg: null,
      loading: false,
      some_data: null,
    };
  },
  methods: {
    async login() {
      this.loading = true;
      this.errormsg = null;
      const user = new Object();
      this.$axios.interceptors.request.use((request) => {
        console.log("Request:", request);
        return request;
      });

      this.$axios.interceptors.response.use(
        (response) => {
          console.log("Response:", response);
          return response;
        },
        (error) => {
          console.log("Error:", error);
          return Promise.reject(error);
        }
      );
      this.$axios.defaults.timeout = 10000;
      try {
        let res = await this.$axios.post(
          "/session",
          {
            name: this.username,
          },
          {
            headers: {
              "Content-Type": "application/json",
            },
            timeout: 30000,
          }
        );

        user.userName = res.data.name;
        user.userId = res.data.id;
        user.photo = res.data.photo;

        if (!this.wantLegacy) {
          localStorage.removeItem("token");
          localStorage.removeItem("user");

          sessionStorage.setItem("token", user.userId);
          sessionStorage.setItem("user", JSON.stringify(user));
        } else {
          sessionStorage.removeItem("token");
          sessionStorage.removeItem("user");

          localStorage.setItem("token", user.userId);
          localStorage.setItem("user", JSON.stringify(user));
        }
        this.loading = false;
        this.$emit("handle-good-login");
      } catch (error) {
        if (error.code === "ECONNABORTED") {
          console.log("First attempt failed, retrying...");
          // Small delay before retry
          await new Promise((r) => setTimeout(r, 100));
          return this.$axios.post("/session", { name: this.username });
        }
        this.$emit("un-successful-login");
        errorManager.addError(error.toString());
      }
    },
    async submit() {
      await this.login();
      if (this.loading === false) {
        this.$router.push("/conversations");
      }
    },
  },
};
</script>

<template>
  <div v-if="!isAuthed" id="container-login" class="grid">
    <div id="login-flex">
      <div id="login-header">
        <h1>Login Form</h1>
      </div>
      <div id="login-form">
        <input
          id="input-user-name"
          v-model="username"
          type="text"
          placeholder="Username"
          @keydown.enter="submit"
        >
      </div>
      <div id="login-footer">
        <div id="input-remember-me">
          Remember Me
          <input id="rememberMe" v-model="wantLegacy" type="checkbox">
        </div>
        <button id="button-submit" class="btn-high-importance" @click="submit">
          Submit
        </button>
      </div>
    </div>
  </div>

  <div v-else>
    <h1>You are Logged !!!</h1>
  </div>
</template>
