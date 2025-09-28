<script>
import CustomHeader from "./CustomHeader.vue";
import Message from "./Message.vue";
import { errorManager } from "../services/axios";
import ConversationInput from "./ConversationInput.vue";
export default {
  components: {
    Message,
    CustomHeader,
    ConversationInput,
  },
  props: ["paused", "conv_name", "userId", "groupId", "conversationId"],
  emits: [
    "close-conversation",
    "open",
    "closed",
    "messageSent",
    "messageDeleted",
    "sendGroupUp",
  ],
  data() {
    return {
      user: this.getUser(),
      my_msgs: [{}],
      other_msgs: [{}],
      msgs: [],
      conv: null,
      loading: false,
      replMessage: null,
      image_chosen: false,
      file: null,
      pid: 0,
    };
  },
  created() {
    this.pid = setInterval(async () => {
      if (this.paused) {
        return;
      }
      await this.getMessages();
    }, 7000);
  },
  updated() {
    console.warn("Updating Conversation");
  },
  async mounted() {
    console.log("Mounted Conversation");

    await this.getMessages();
    this.scrollToBottomOfConversation();
  },
  beforeUnmount() {
    clearInterval(this.pid);
  },

  methods: {
    getUser() {
      if (localStorage.getItem("user")) {
        return JSON.parse(localStorage.getItem("user"));
      } else if (sessionStorage.getItem("user")) {
        return JSON.parse(sessionStorage.getItem("user"));
      }
      return null;
    },

    async getMessages() {
      this.loading = true;
      this.errormsg = null;

      try {
        let response = await this.$axios.get(
          `/users/${this.userId}/conversations/${this.conversationId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        this.conv = response.data;
        if (!this.conv.messages) {
          return;
        }
        this.conv.messages.sort((a, b) => {
          return a.timestamp - b.timestamp;
        });

        this.msgs = [];
        this.conv.messages.forEach((element) => {
          if (element.sender.id === this.userId) {
            element.mine = true;
          }
          this.msgs.push(element);
        });
      } catch (e) {
        errorManager.addError("Error while getting message:", e.toString());
      } finally {
        this.loading = false;
      }
    },
    deleteComment(msg, commentId) {
      console.log(commentId);
      for (const [i, value] of msg.commentList.entries()) {
        if (value.id === commentId) {
          delete msg.commentList[i];
          break;
        }
      }
    },
    deleteMessage(msgId) {
      console.log("delete message from list");
      for (const [i, msg] of this.msgs.entries()) {
        if (msg.id === msgId) {
          delete this.msgs[i];
          break;
        }
      }

      this.$emit("messageDeleted", this.msgs[this.msgs.length - 1]);
      console.log(this.msgs);
      if (this.msgs.length === 0) {
        this.$emit("close-conversation");
      }
    },
    sendGroupUp(group) {
      this.$emit("sendGroupUp", group);
    },
    scrollToBottomOfConversation() {
      /*const theEnd = document.getElementsByName(
        "msg" + this.msgs.length.toString()
      );
      if (!theEnd[0]) {
        return;
      }
      console.log("Scrolling to bottom:", theEnd[0]);
      theEnd[0].scrollIntoView({ behavior: "smooth" });*/
      const the_end = document.getElementById("the_end");
      if (!the_end) {
        return;
      }
      the_end.scrollIntoView({ behavior: "smooth" });
    },
    scrollToMessage(msgName) {
      const msg = document.getElementsByName(msgName);
      if (msg !== null) {
        console.log("Msg:", msg);
        msg[0].scrollIntoView({ behavior: "smooth" });

        msg[0].classList.remove("highlight");
        setTimeout(() => {
          msg[0].classList.add("highlight");

          // Remove the highlight class after animation completes
          setTimeout(() => {
            msg[0].classList.remove("highlight");
          }, 2000);
        }, 300);
      }
      // Add highlight animation after a brief delay
    },
    replyMessage(re) {
      this.replMessage = re;
    },
    closeReplyMessage() {
      this.replMessage = null;
    },
  },
};
</script>

<template>
  <div v-if="conv" id="conversation">
    <CustomHeader
      :user-id="userId"
      :user-name="conv_name"
      :group-id="groupId"
      @send-group-up="sendGroupUp"
    />
    <div id="conv_body" ref="conv_body" class="grid">
      <ul id="message_list" ref="message_list">
        <Message
          v-for="msg in msgs"
          ref="message"
          :key="msg"
          :msg="msg"
          :user-id="userId"
          :conversation-id="conversationId"
          :date="new Date(msg.timestamp * 1000)"
          @scroll-to-msg="scrollToMessage"
          @open="$emit('open')"
          @close="$emit('closed')"
          @comment-deleted="deleteComment(msg, commentId)"
          @new-comment="$emit('messageDeleted')"
          @message-deleted="deleteMessage"
          @reply-message="replyMessage"
        />
        <div id="the_end" ref="bottomEl" style="height: 1px; width: 1px" />
      </ul>
      <!--<button class="godown-btn btn" @click="scrollToBottomOfConversation">⬇️</button>-->
    </div>
    <div id="conv_footer">
      <div
        v-if="replMessage"
        id="re_message"
        style="
          max-width: 66ch;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        "
      >
        <button class="btn" title="Cancel Reply" @click="closeReplyMessage">
          🗑️
        </button>
        <span>
          {{ replMessage.content }}
        </span>
      </div>
      <div id="input-area">
        <ConversationInput
          :conversation-id="conversationId"
          :repl-message="replMessage"
          @message-sent="
            $emit('messageSent');
            scrollToBottomOfConversation();
          "
        />
      </div>
    </div>
  </div>
</template>
