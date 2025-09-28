<script>
import ForwardingModal from "./ForwardingModal.vue";
import data from "emoji-mart-vue-fast/data/google.json";
import "emoji-mart-vue-fast/css/emoji-mart.css";
import { Picker, EmojiIndex } from "emoji-mart-vue-fast/src";
import Delete from "./Delete.vue";
import CommentList from "./CommentList.vue";
import { errorManager } from "../services/axios";
export default {
  components: {
    ForwardingModal,
    CommentList,
    Picker,
    Delete,
  },
  props: [
    "commentList",
    "msg",
    "comment",
    "comment_id",
    "msg_mine",
    "userId",
    "conversationId",
    "username",
    "msg_content",
    "date",
  ],
  emits: [
    "replyMessage",
    "NewComment",
    "messageDeleted",
    "scrollToMsg",
    "open",
    "close",
  ],
  data() {
    return {
      time: this.time_complete(),
      msg_been_read: false,
      emojiIndex: new EmojiIndex(data),
      choose_emoji: false,
      messageId: this.msg.id,
      image: "",
      msgName: "msg" + this.msg.id,
      userName: JSON.parse(
        sessionStorage.getItem("user")
          ? sessionStorage.getItem("user")
          : localStorage.getItem("user")
      ).userName,
      replmsg: null,
      k: 0,
    };
  },
  mounted() {
    console.log("Rendered message component:", this.msg.id);
    this.get_status();

    if (this.msg.pictureId > 0 && this.image === "") {
      this.getImage();
    }
    if (this.msg.repliedId) {
      this.getMessage(this.msg.repliedId);
    }
    if (!this.msg_mine && !this.msg_been_read) {
      this.readMessage();
    }

    console.log(this.msg);
  },
  beforeUpdate() {
    console.log("Before Updated message component:", this.msg.id);
  },
  beforeUnmount() {
    console.log("Before Unmount message component:", this.msg.id);
    URL.revokeObjectURL(this.image);
  },
  methods: {
    async get_status() {
      try {
        if (this.userId && this.conversationId && this.messageId) {
          const result = await this.$axios.get(
            `/users/${this.userId}/conversations/${this.conversationId}/messages/${this.messageId}/status`,
            {
              headers: {
                Authorization: `Bearer ${this.userId}`,
                "Content-Type": "application/json",
              },
            }
          );
          this.msg_been_read = result.data.hasBeenRead;
        }
      } catch (error) {
        errorManager.addError(
          "Error while getting message status:" + error.toString()
        );
      }
    },
    async handleNewComment(emoji) {
      try {
        const result = await this.$axios.post(
          `/users/${this.userId}/conversations/${this.conversationId}/messages/comments/${this.messageId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            id: this.comment_id,
            content: emoji,
            userName: this.userName,
          }
        );
        if (result.status === 204) {
          this.k += 1;
        }
      } catch (error) {
        errorManager.addError(
          "Error while getting message comment" + error.toString()
        );
        return "";
      }
    },
    get_hour() {
      let hour = this.date.getHours().toString();
      if (hour.length === 0 || hour.length === 1) {
        hour * "0" + hour;
      }

      return hour;
    },
    get_minutes() {
      let min = this.date.getMinutes().toString();
      if (min.length === 0 || min.length === 1) {
        min = "0" + min;
      }
      return min;
    },
    get_seconds() {
      let sec = this.date.getSeconds().toString();
      if (sec.length === 0 || sec.length === 1) {
        sec = "0" + sec;
      }
      return sec;
    },

    time_complete() {
      return (
        this.date.toDateString() +
        " " +
        this.get_hour() +
        ":" +
        this.get_minutes() +
        ":" +
        this.get_seconds()
      );
    },

    handleMessageForwarding() {
      this.open_modal = true;
    },

    showEmoji(emoji) {
      this.handleNewComment(emoji.native);
      this.$emit("NewComment");
      this.choose_emoji = false;
    },

    async readMessage() {
      try {
        this.loading = true;
        let response = await this.$axios.post(
          `/users/${this.userId}/conversations/${this.conversationId}/messages/read/${this.messageId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          console.log("Message Has Been Read");
        }
      } catch (error) {
        errorManager.addError(
          "error while Reading message:" + error.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    async handleDeleteMessage() {
      try {
        const result = await this.$axios.delete(
          `/users/${this.userId}/conversations/${this.conversationId}/messages/${this.messageId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (result.status === 204) {
          console.log("message deleted going up");
          this.$emit("messageDeleted", this.messageId);
        }
      } catch (error) {
        console.error(error);
        return "";
      }
    },
    async getMessage(msgId) {
      try {
        const result = await this.$axios.get(
          `/conversations/${this.conversationId}/messages/${msgId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        this.replmsg = result.data;
      } catch (error) {
        errorManager.addError("Error while getting replied to message");
      }
    },
    async getImage() {
      try {
        console.log("getting image");
        const result = await this.$axios.get(
          `/photos/${this.msg.pictureId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          },
          {}
        );
        URL.revokeObjectURL(this.image);
        this.image = URL.createObjectURL(result.data);
      } catch (error) {
        console.error(error);
        errorManager.addError(error);
      }
    },
    handleScrollToMsg(messageName) {
      this.$emit("scrollToMsg", messageName);
    },
    handleReplyMessage() {
      this.$emit("replyMessage", this.msg);
    },
  },
};
</script>
<template>
  <li
    :name="msgName"
    class="message msg_container .noto-color-emoji-regular"
    :class="{ other_msgs: !msg.mine, my_msgs: msg.mine }"
  >
    <div class="message-header">
      <div class="normal-stuff">
        <div style="width: 100%" class="username">{{ msg.sender.name }}</div>
        <button
          title="Reply To Message"
          class="reply-button msg btn"
          @click="handleReplyMessage"
        >
          REPLY
        </button>
      </div>
      <div v-if="replmsg" class="reply-message">
        <button
          class="scroll-to-msg btn"
          @click="handleScrollToMsg('msg' + msg.repliedId.toString())"
        >
          Replied to:
          {{ replmsg.content }}
        </button>
      </div>
    </div>
    <div class="content-container">
      <img
        v-if="msg.pictureId > 0"
        class="message-image"
        :src="image"
        alt="image"
        title="Image"
      />
      <div class="text">
        {{ msg.content !== "📷 Photo" ? msg.content : "" }}
      </div>
    </div>
    <div class="timestamp">{{ time }}</div>
    <div class="msg_footer">
      <div
        v-if="msg.mine"
        class="sent"
        :class="{ hidden: msg_been_read }"
        title="sent"
      >
        ✔️
      </div>
      <div
        v-if="msg.mine"
        class="received"
        :class="{ hidden: !msg_been_read }"
        title="received"
      >
        ✔️✔️
      </div>

      <ForwardingModal
        :message-id="msg.id"
        :conversation-id="conversationId"
        :user-id="userId"
        @open="$emit('open')"
        @close="$emit('close')"
      />
      <input
        title="Choose your emoji"
        type="button"
        class="btn"
        value="😊"
        :data="choose_emoji"
        @click="choose_emoji = !choose_emoji"
      />
      <CommentList
        v-if="msg.commentList"
        :key="k"
        :message-id="msg.id"
        :conversation-id="conversationId"
        :user-id="msg.sender.id"
        @comment-deleted="$emit('commentDeleted')"
        @comment-list-opened="$emit('open')"
        @comment-list-closed="$emit('close')"
      />
      <Delete
        v-if="msg.mine"
        :title="'Delete Message'"
        @handle-delete="handleDeleteMessage"
        @open="$emit('open')"
        @close="$emit('close')"
      />
    </div>

    <Picker
      v-if="choose_emoji"
      title="Pick your emoji…"
      :emoji-tooltip="true"
      :show-preview="false"
      :data="emojiIndex"
      set="apple"
      @select="showEmoji"
    />
  </li>
</template>
