<script>
import { errorManager } from "../services/axios";
import LoadingSpinner from "./LoadingSpinner.vue";

export default {
  components: {
    LoadingSpinner,
  },
  props: ["user"],
  emits: ["changeUsername", "changeUserImage"],
  data() {
    return {
      new_username: "",
      user_image: "",
      loading: false,
    };
  },
  mounted() {
    if (this.user_image !== "") {
      URL.revokeObjectURL(this.user_image);
    }
    this.getUserImage();
  },
  beforeUnmount() {
    URL.revokeObjectURL(this.user_image);
  },
  methods: {
    checkUsername() {
      if (this.new_username.length > 66) {
        errorManager.addError("The username is too long at most 66 chars");
        return true;
      }

      if (this.new_username.includes(" ")) {
        errorManager.addError(
          "The username is too long at most , spaces are not allowed chars"
        );
        return true;
      }
      if (this.new_username.trim("").length == 0) {
        errorManager.addError("The username cannot be empty");
      }
      return false;
    },

    async getUserImage() {
      try {
        this.loading = true;
        const response = await this.$axios.get(
          `/users/${this.user.userId}/photo`,
          {
            headers: {
              Authorization: `Bearer ${this.user.userId}`,
            },
            responseType: "blob",
          }
        );
        this.user_image = URL.createObjectURL(response.data);
      } catch (error) {
        console.error(error);
        errorManager.addError(error);
      } finally {
        this.loading = false;
      }
    },

    async changeUsername() {
      let response = null;
      try {
        this.loading = true;

        if (this.checkUsername()) {
          this.loading = false;
          return;
        }

        response = await this.$axios.post(
          `/users/${this.user.userId}/username`,
          {
            id: this.user.userId,
            name: this.new_username,
            photo: null,
          },
          {
            headers: {
              Authorization: `Bearer ${this.user.userId}`,
            },
          }
        );
        if (response.status === 204) {
          console.log("The username has been changed");
          this.$emit("changeUsername", this.new_username);
          this.new_username = "";
        } else if (response.status === 500) {
          console.warn(response.response);
        }
      } catch (error) {
        errorManager.addError(
          "Error while changing username:" + error.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    async changeUserImage() {
      try {
        this.loading = true;
        // Get the file input element
        const fileInput = document.querySelector('input[name="fileUp"]');
        const file = fileInput.files[0];
        console.log(file);

        // Check if a file was selected
        if (!file) {
          throw new Error("No file selected");
        }

        // Check file type (only allow PNG as specified in your input accept attribute)
        if (file.type !== "image/png") {
          throw new Error("Only PNG files are allowed");
        }

        // Create FormData to send the file
        const formData = new FormData();
        formData.append("photo", file);

        const photoId = 0;

        // Send the request to the server
        const response = await this.$axios.post(
          `/photos/${photoId}/users/${this.user.userId}/photo`,
          formData,
          {
            headers: {
              Authorization: `Bearer ${this.user.userId}`,
              "Content-Type": "multipart/form-data",
            },
          }
        );

        // Handle successful response
        if (response.status === 200) {
          console.log("The profile picture has been changed");

          // Update the profile picture display
          await this.getUserImage();
          // Clear the file input
          fileInput.value = "";
        }
      } catch (error) {
        console.error("Error changing profile picture:", error);
        errorManager.addError("Error changing profile picture:", error);
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>

<template>
  <div id="container-user-info">
    <LoadingSpinner :loading="loading" />
    <div id="user-info-header">
      <img id="user-profile-picture" :src="user_image" alt="profile-picture" />

      <label
        id="change_pp"
        class="custom-file-upload btn btn-high-importance"
        title="Send image"
      >
        Change Profile Picture
        <input
          class="btn"
          type="file"
          name="fileUp"
          accept=".png"
          @input="changeUserImage"
        />
      </label>
    </div>
    <div id="container-user-name">
      <input
        v-model="new_username"
        type="text"
        :placeholder="user.userName"
        @keyup.enter="changeUsername"
      />
      <input
        class="modify-info btn"
        title="Change username"
        type="button"
        value="🖊️"
        @click="changeUsername"
        @keyup.enter="changeUsername"
      />
    </div>
  </div>
</template>
