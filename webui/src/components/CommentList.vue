<script>
import { errorManager } from "../services/axios";
export default {
  props: ["commentList", "messageId", "conversationId", "userId"],
  emits: ["commentDeleted", "comment-list-opened", "comment-list-closed"],
  data() {
    return {
      comList: [],
      show: false,
    };
  },
  mounted() {
    console.log("Comment List Was Mounted");
    this.getCommentList(this.messageId, this.conversationId, this.userId);
  },
  updated() {
    console.log("Comment List Was Updated");
  },
  methods: {
    async getCommentList(messageId, conversationId, userId) {
      try {
        const response = await this.$axios.get(
          `/users/${userId}/conversations/${conversationId}/messages/${messageId}/comments`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 200) {
          this.comList = response.data;
          console.log("Comment List : ", this.comList);
        }
      } catch (error) {
        errorManager.addError(
          "Error while getting comment:" + error.toString()
        );
      }
    },
    async deleteComment(messageId, commentId, conversationId, userId) {
      try {
        const response = await this.$axios.delete(
          `/users/${userId}/conversations/${conversationId}/messages/${messageId}/comments/${commentId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          for (const [i, value] of this.comList.entries()) {
            if (value.id === commentId) {
              console.log("Deleting message");
              delete this.comList[i];
              break;
            }
          }
        }
      } catch (error) {
        errorManager.addError(
          "Error while deleting comment:" + error.toString()
        );
      }
    },
    handleOpen() {
      if (this.show) {
        this.show = false;
        this.$emit("comment-list-closed");
      } else {
        this.show = true;
        this.$emit("comment-list-opened");
      }
    },
  },
};
</script>

<template>
  <input
    type="button"
    class="btn comment_button"
    value="📃"
    title="show_reactions"
    @keyup.enter="handleOpen"
    @click="handleOpen"
  >
  <ul v-if="show && comList.length > 0" class="comment_list_element">
    <li v-for="comment in comList" :key="comment" class="comment_container">
      <div v-if="comment" title="comment" class="comment">
        {{ comment.content }} {{ comment.userName }}
      </div>
      <input
        v-if="comment && comment.userId === userId"
        type="button"
        class="btn"
        value="✖️"
        title="delete comment"
        @click="
          deleteComment(
            comment.messageId,
            comment.id,
            comment.conversationId,
            comment.userId
          )
        "
      >
    </li>
  </ul>
</template>
