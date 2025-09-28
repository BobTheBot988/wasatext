<script>
import BaseModal from "./BaseModal.vue";
import FancyList from "./FancyList.vue";
import Search from "./Search.vue";
import LoadingSpinner from "./LoadingSpinner.vue";
import { errorManager } from "../services/axios";

export default {
  components: {
    BaseModal,
    Search,
    FancyList,
    LoadingSpinner,
  },
  props: ["userId", "conversationId", "messageId"],
  emits: ["open", "close"],
  data() {
    return {
      open: false,
      search: "",
      convs: [],
      pid: 0,
    };
  },
  computed: {
    filteredConvs() {
      if (this.convs === null) {
        return [];
      }
      if (this.search === "") {
        return this.convs;
      }
      return this.convs.filter((conv) =>
        conv.name.toLowerCase().includes(this.search.toLowerCase())
      );
    },
    displayItems() {
      return this.filteredConvs;
    },
  },
  created() {
    this.pid = setInterval(async () => {
      if (this.paused) {
        return;
      }
      await this.getConvs();
    }, 7000);
  },
  async mounted() {
    await this.getConvs();
  },
  beforeUnmount() {
    clearInterval(this.pid);
  },
  methods: {
    async forwardMessage(convId) {
      try {
        this.loading = true;
        const response = await this.$axios.post(
          `/users/${this.userId}/conversations/${this.conversationId}/messages/forward/${this.messageId}`,
          {
            forwardTo: convId,
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          this.$router.go();
          this.$refs.modal.closeModal();
        } else {
          throw new Error("error", response.status);
        }
      } catch (error) {
        console.log(error);
      } finally {
        this.loading = false;
      }
    },
    handleForwarding(name, convId) {
      this.forwardMessage(convId);
    },
    isNameSame(name) {
      if (this.convs === null) {
        return false;
      }
      for (const conv of this.convs) {
        if (conv.name === name) {
          return false;
        }
      }
      return true;
    },
    async getConvs() {
      try {
        this.loading = true;

        if (!this.userId) {
          throw new Error("Forwarding error", "no user id");
        }

        const request = await this.$axios.get(
          `/users/${this.userId}/conversations`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );

        if (this.convs !== null && this.convs.length > 0) {
          this.convs.forEach((conv) => {
            URL.revokeObjectURL(conv.photo);
          });
        }

        this.convs = request.data;
        if (this.convs === null) {
          return [];
        }
        const users = JSON.parse(sessionStorage.getItem("users"));
        this.convs.sort((a, b) => {
          return b.lastMsgTimeStmp - a.lastMsgTimeStmp;
        });
        if (users) {
          users.forEach((user) => {
            if (user.id === this.userId || !this.isNameSame(user.name)) {
              return;
            }
            user.id = -user.id;
            this.convs.push(user);
          });
        }

        for (const conv of this.convs) {
          conv.photo = await this.getItemPhoto(conv);
        }
        console.log(this.convs);
      } catch (e) {
        errorManager.addError(
          "Error while getting conversations for forwarding" + e.toString()
        );
      } finally {
        this.loading = false;
      }
    },

    async getItemPhoto(item) {
      try {
        this.loading = true;
        let url = "";

        if (item.groupId !== 0) {
          url = `/groups/${item.groupId}/photo`;
        } else {
          if (item.userId) {
            url = `/users/${item.userId}/photo`;
          } else if (item.id) {
            url = `/users/${Math.abs(item.id)}/photo`;
          }
        }
        const response = await this.$axios.get(url, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
          responseType: "blob",
        });

        return URL.createObjectURL(response.data);
      } catch (error) {
        console.error(error);
        errorManager.addError(error);
      } finally {
        this.loading = false;
      }
    },
    handleFilter(new_search) {
      this.search = new_search;
    },
  },
};
</script>
<template>
  <BaseModal
    ref="modal"
    :val="'➡️'"
    @opened="$emit('open')"
    @close="$emit('close')"
  >
    <LoadingSpinner :loading="loading" />
    <template #title> Forward to conversation</template>
    <Search :user-id="userId" placeholder="Choose Conversation">
      <template #search />
      <template #search-result>
        <FancyList :items="displayItems">
          <template #item="item">
            <li class="search-result-item" @click="forwardMessage(item.id)">
              <div
                class="conversationpw-container"
                role="button"
                title="Click on conversation to open"
              >
                <img v-if="item.photo" :src="item.photo" alt="ProfilePicture" />
                <div class="conv-name noto-color-emoji-regular">
                  {{ item.name }}
                </div>
                <div class="last-msg-content noto-color-emoji-regular">
                  {{
                    item.lastMsgContent === ""
                      ? "Send a message!"
                      : item.lastMsgContent
                  }}
                </div>
              </div>
            </li>
          </template>
        </FancyList>
      </template>
    </Search>
  </BaseModal>
</template>
