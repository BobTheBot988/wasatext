<script>
import Conversation from "../components/Conversation.vue";
import ConversationList from "../components/ConversationList.vue";

export default {
  components: {
    Conversation,
    ConversationList,
  },
  props: ["paused", "user", "users"],
  emits: ["sendGroupUp", "pause", "unpaused"],
  data() {
    return {
      convId: -1,
      name: "",
      conv_selected: false,
      k: 0,
      groupId: 0,
    };
  },
  methods: {
    handle_conv_selected(name, convId, groupId) {
      if (this.convId === convId) {
        return;
      }
      this.conv_selected = false;
      setTimeout(() => {
        this.conv_selected = true;
        this.convId = convId;
      }, 0.0);
      this.groupId = groupId;
      this.name = name;
    },
    handle_conv_close() {
      this.conv_selected = false;
      this.convId = -1;
    },
    handle_message_sent() {
      this.k += 1;
    },
    handle_message_delete() {
      this.k -= 1;
    },
    sendGroupUp(group) {
      this.$emit("sendGroupUp", group);
    },
  },
};
</script>

<template>
  <div id="conversation-view" class="grid">
    <div id="left">
      <div id="conv-list">
        <ConversationList
          :key="k"
          :user-id="user.userId"
          @open-conversation="handle_conv_selected"
          @close-conversation="handle_conv_close"
        />
      </div>
    </div>

    <div id="right">
      <Conversation
        v-if="conv_selected"
        :key="k"
        :paused="paused"
        :conv_name="name"
        :user-id="user.userId"
        :group-id="groupId"
        :conversation-id="convId"
        @send-group-up="sendGroupUp"
        @message-deleted="handle_message_delete"
        @message-sent="handle_message_sent"
        @open="$emit('pause')"
        @closed="$emit('unpause')"
        @close-conversation="handle_conv_close"
      />
      <div v-else>Select a conversation</div>
    </div>
  </div>
</template>
