<template>
  <div>
    <form @submit.prevent="submitForm">
      <div class="form-group">
        <input
          v-model="formData.username"
          type="text"
          placeholder="Username"
          :disabled="loading"
        >
      </div>

      <div class="form-group">
        <label>
          <input v-model="formData.wantLegacy" type="checkbox"> Remember me
        </label>
      </div>

      <div v-if="errorMsg" class="error-message">
        {{ errorMsg }}
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? "Logging in..." : "Login" }}
      </button>
    </form>
  </div>
</template>

<script>
export default {
  name: "LoginForm",

  emits: {
    "form-submit": (username) => {
      return username.length > 0;
    },
  },
  data() {
    return {
      formData: {
        username: "",
        wantLegacy: false,
      },
    };
  },
  methods: {
    submitForm() {
      console.log("Submitting form");
      if (username.length > 66) {
        errorManager.addError("The username can't be this long");
      }
      this.$emit("form-submit", this.formData);
    },
  },
};
</script>
