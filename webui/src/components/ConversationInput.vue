<script>
import { errorManager } from "../services/axios";
export default {
  props: ["userId", "conversationId", "replMessage"],
  emits: ["messageSent"],
  data() {
    return {
      message: "",
      loading: false,
      user: localStorage.getItem("user")
        ? JSON.parse(localStorage.getItem("user"))
        : JSON.parse(sessionStorage.getItem("user")),
      image_chosen: false,
      file: null,
    };
  },
  watch: {
    replMessage(new_rp, old_rp) {
      old_rp = new_rp;
      const input_area = document.getElementById("message_area");
      if (!input_area) return;
      input_area.focus();
    },
  },

  mounted() {
    const area = document.getElementById(message_area);
    if (!area) return;
    area.focus();
  },
  methods: {
    deleteChosenImage() {
      console.log("Deleting image");
      this.file = null;
      this.image_chosen = false;
    },

    getImage() {
      try {
        console.log("Getting image");
        // Get the file input element
        const fileInput = document.querySelector('input[name="fileUp"]');
        this.file = fileInput.files[0];

        // Check if a file was selected
        if (!this.file) {
          throw new Error("No file selected");
        }

        // Check file type (only allow PNG as specified in your input accept attribute)
        if (this.file.type !== "image/png") {
          throw new Error("Only PNG files are allowed");
        }
        console.log(this.file);
        this.image_chosen = true;
        return this.file;
      } catch (error) {
        errorManager.addError(error);
      }
    },
    async sendPhoto() {
      this.loading = true;
      this.errormsg = null;

      console.log("Sending image");
      try {
        // Create FormData to send the file
        const formData = new FormData();
        formData.append("photo", this.file);

        formData.append(
          "sender",
          JSON.stringify({
            id: this.user.userId,
            name: this.user.userName,
          })
        );
        formData.append("convId", this.conversationId);
        formData.append("content", this.message);
        formData.append(
          "repliedId",
          this.replMessage ? this.replMessage.id : null
        );
        formData.append(
          "repliedConvId",
          this.replMessage ? this.replMessage.conversationId : null
        );
        console.warn("Formadata:", formData);
        let response = await this.$axios.post(
          `/users/${this.user.userId}/conversations/${this.conversationId}/messages/photo`,
          formData,
          {
            headers: {
              Authorization: `Bearer ${this.user.userId}`,
            },
          }
        );
        this.some_data = response.data;
        this.message = "";
      } catch (e) {
        errorManager.addError("Error while sending picture:" + e.toString());
      } finally {
        this.loading = false;
        this.deleteChosenImage();
      }
    },
    checkCorrectnessMessage() {
      this.message = this.message.trim();
      const len = this.message.length;
      if (len === 0) {
        this.message = "";
        this.errormsg = "The message was empty";
        return false;
      } else if (len > 1000) {
        this.errormsg = "The message must be shorter then 1000 characters";
        return false;
      }
      // implement regex
      return true;
    },
    async sendMessage() {
      this.loading = true;
      this.errormsg = null;
      if (!this.checkCorrectnessMessage()) {
        return;
      }
      console.warn("Sending message!\nReplMessage:", this.replMessage);
      try {
        let response = await this.$axios.post(
          `/users/${this.user.userId}/conversations/${this.conversationId}/messages`,
          {
            sender: {
              id: this.user.userId,
              name: this.user.userName,
            },
            convId: this.conversationId,
            content: this.message,
            repliedId: this.replMessage ? this.replMessage.id : null,
            repliedConvId: this.replMessage
              ? this.replMessage.conversationId
              : null,
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        console.log("Message:", this.message);
        console.log("Message Response:", response);
        this.some_data = response.data;
        this.message = "";
      } catch (e) {
        errorManager.addError("Error While sending message" + e.toString());
      } finally {
        this.loading = false;
      }
    },
    async handleSendMessage() {
      if (this.file !== null) {
        await this.sendPhoto();
        this.message = "";
        this.$emit("messageSent");
        return;
      }
      if (this.message.length === 0) {
        errorManager.addError("Message empty");
        return;
      }

      await this.sendMessage();

      this.message = "";

      this.$emit("messageSent");
    },
  },
};
</script>
<template>
  <div id="left_msg">
    <label v-if="!image_chosen" class="custom-file-upload" title="Send image">
      🖼️
      <input
        name="fileUp"
        type="file"
        accept=".png"
        @input="getImage"
        @keyup.enter="getImage"
      >
    </label>

    <input
      v-else
      name="custom-file-upload-delete"
      type="button"
      value="🖼️X"
      title="Delete current selected image"
      @click="deleteChosenImage"
      @keyup.enter="deleteChosenImage"
    >
  </div>

  <textarea
    id="message_area"
    ref="message_area"
    v-model="message"
    minlength="1"
    maxlength="1000"
    placeholder="Write your message..."
    autocorrect="on"
    label="message"
    spellcheck="true"
    autofocus
    :disabled="loading"
    @keydown.enter.exact.prevent="handleSendMessage"
    @keydown.shift.enter="message = message + '\n'"
  />
  <button
    id="sub_button"
    class="btn"
    type="submit"
    title="Send"
    :disabled="loading"
    @click="handleSendMessage"
    @keydown.enter="handleSendMessage"
  >
    →
  </button>
</template>
