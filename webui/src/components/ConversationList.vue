<script>
import FancyList from "./FancyList.vue";
import LoadingSpinner from "./LoadingSpinner.vue";
import Search from "./Search.vue";
import { errorManager } from "../services/axios";

export default {
  components: {
    Search,
    FancyList,
    LoadingSpinner,
  },
  emits: ["open-conversation"],
  data() {
    return {
      userId: localStorage.getItem("user")
        ? JSON.parse(localStorage.getItem("user")).userId
        : JSON.parse(sessionStorage.getItem("user")).userId,
      search: "",
      convs: [],
      new_users: [],
      loading: false,
      loading_convs: false,
      paused: false,
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
  async created() {
    this.pid = setInterval(async () => {
      if (this.paused) {
        return;
      }
      await this.getConvs();
    }, 7000);
  },
  async mounted() {
    console.log("Mounted conversation list component");
    await this.getConvs();
  },
  updated() {
    console.log("Updated conversation list component");
  },
  beforeUnmount() {
    clearInterval(this.pid);
    this.convs.forEach((conv) => {
      URL.revokeObjectURL(conv.photo);
    });
  },
  methods: {
    handleOpenConversation(name, convId, groupId) {
      this.$emit("open-conversation", name, convId, groupId);
    },
    async getItemPhoto(item) {
      try {
        this.loading = true;
        let response = null;

        if (item.groupId !== 0) {
          response = await this.$axios.get(`/groups/${item.groupId}/photo`, {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          });
        } else {
          response = await this.$axios.get(`/users/${item.userId}/photo`, {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          });
        }

        return URL.createObjectURL(response.data);
      } catch (error) {
        errorManager.addError("Error while getting item photo" + error);
      } finally {
        this.loading = false;
      }
    },
    async getConvs() {
      try {
        this.loading_convs = true;

        if (!this.userId) {
          return;
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

        this.convs.sort((a, b) => {
          return b.lastMsgTimeStmp - a.lastMsgTimeStmp;
        });
        console.log("Getting Photos");
        for (const conv of this.convs) {
          conv.photo = await this.getItemPhoto(conv);
        }
        const cs = JSON.parse(sessionStorage.getItem("convs"));
        if (cs) {
          cs.forEach((c) => {
            URL.revokeObjectURL(c.photo);
          });
        }
        sessionStorage.setItem("convs", JSON.stringify(this.convs));
      } catch (e) {
        errorManager.addError("Error while getting convs:" + e.toString());

        console.error("Error convs:" + this.convs);
      } finally {
        this.loading_convs = false;
      }
    },
    handleFilter(new_search) {
      this.search = new_search;
    },
  },
};
</script>
<template>
  <div v-if="loading_convs">
    <LoadingSpinner :loading="false" />
  </div>

  <div id="search-convs">
    <div id="header">
      <p>Conversations</p>
    </div>
    <Search :placeholder="'Search Conversations'" @filter="handleFilter">
      <template #search />
      <template #search-result>
        <FancyList :items="displayItems">
          <template #item="item">
            <li
              class="search-result-item"
              @click="handleOpenConversation(item.name, item.id, item.groupId)"
            >
              <div
                class="conversationpw-container"
                role="button"
                title="Click on conversation to open"
              >
                <img v-if="item.photo" :src="item.photo" alt="ProfilePicture" />
                <div class="details">
                  <div class="conv-name noto-color-emoji-regular">
                    {{ item.name }}
                  </div>
                  <div
                    class="last-msg-content noto-color-emoji-regular"
                    :title="item.lastMsgContent"
                  >
                    {{
                      item.lastMsgContent === ""
                        ? "Send a message!"
                        : item.lastMsgContent
                    }}
                    <div class="timestamp">
                      {{ new Date(item.lastMsgTimeStmp * 1000) }}
                    </div>
                  </div>
                </div>
              </div>
            </li>
          </template>
        </FancyList>
      </template>
    </Search>
  </div>
</template>
