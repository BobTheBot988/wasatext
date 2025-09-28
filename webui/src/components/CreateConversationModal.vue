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
  emits: ["close"],

  data() {
    return {
      step_one: true,
      userId: localStorage.getItem("user")
        ? JSON.parse(localStorage.getItem("user")).userId
        : JSON.parse(sessionStorage.getItem("user")).userId,
      group_name: "",
      search: "",
      users: [],
      new_users: new Set([this.userId]),
      convs_list: [],
      loading: false,
      loadingUsers: false,
    };
  },

  computed: {
    filteredUsers() {
      if (this.users === null) {
        return [];
      }
      return this.users.filter(
        (user) =>
          this.isNameSame(user.name) &&
          user.name.toLowerCase().includes(this.search.toLowerCase())
      );
    },
    displayItems() {
      return this.filteredUsers;
    },
  },
  async beforeMount() {
    console.log("BeforeMounted CreateConversationModal");
    await this.getUsers();
    console.log(this.users);
  },

  methods: {
    isNameSame(name) {
      const convs_list = new Set(JSON.parse(sessionStorage.getItem("convs")));
      if (convs_list === null || convs_list.length === 0) {
        return false;
      }
      for (const conv of convs_list) {
        if (conv.name === name) {
          return false;
        }
      }
      return true;
    },
    async getItemPhoto(item) {
      try {
        this.loading = true;
        let response = null;
        response = await this.$axios.get(`/users/${item.id}/photo`, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
          responseType: "blob",
        });

        return URL.createObjectURL(response.data);
      } catch (error) {
        errorManager.addError(
          "Error while getting user photo" + error.toString()
        );
        return null;
      } finally {
        this.loading = false;
      }
    },
    async getUsers() {
      try {
        this.loadingUsers = true;
        const response = await this.$axios.get(`/users/${this.userId}`, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
        });
        if (this.users && this.users.length > 0) {
          this.users.forEach((user) => {
            URL.revokeObjectURL(user.photo);
          });
        }
        this.users = response.data;

        // Set loading for photo fetching phase

        this.users = this.users.filter((user) => user.id !== this.userId);
        console.log("Users:", this.users);
        for (const user of this.users) {
          user.photo = await this.getItemPhoto(user);
        }
      } catch (error) {
        console.error(error);
        errorManager.addError(error);
      } finally {
        this.loadingUsers = false;
      }
    },
    async create_new_conv(new_user_id) {
      this.loading = true;

      if (this.userId === new_user_id) {
        this.loading = false;
        throw new Error("Cannot create a conversation with yourself");
      }

      try {
        console.log("requesting new conv");

        let response = await this.$axios.post(
          `/conversations/create`,
          {
            userIdList: [this.userId, new_user_id],
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );

        console.log("Data from server: ", response.data);
        this.some_data = response.data;
        console.log(response);

        switch (response.status) {
          case 204:
            console.log("Conversation created successfully");
            this.$router.go("");
            break;
          case 201:
            console.log("Conversation created successfully");
            this.$router.go("");
            break;
          case 400:
            throw new Error("The conversation already exists");
          case 500:
            throw new Error("Internal Server error");
        }
      } catch (e) {
        errorManager.addError(
          "Error while creating new conversation:" + e.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    handleNewConversation(new_user_id) {
      this.create_new_conv(new_user_id);
    },
    handleFilter(value) {
      this.search = value;
    },
  },
};
</script>
<template>
  <BaseModal
    :icon="'/feather-sprite-v4.29.0.svg#user-plus'"
    :val="' New Conversation'"
    @close="$emit('close')"
  >
    <template #title>
      <span>Choose the user</span>
    </template>

    <div class="modal-content-wrapper">
      <!-- Conversation creation loading -->
      <div v-if="loading" class="loading-overlay">
        <LoadingSpinner :loading="loading" />
        <span>Creating conversation...</span>
      </div>

      <Search :placeholder="'Search Users ...'" @filter="handleFilter">
        <template #search-result>
          <!-- Show loading spinner while fetching users -->
          <div v-if="loadingUsers" class="text-center py-4">
            <LoadingSpinner :loading="true" />
            <div>Loading users...</div>
          </div>

          <!-- Show users list when loaded -->
          <FancyList :items="displayItems">
            <template #item="item">
              <li
                class="search-result-item"
                :class="{ disabled: loading }"
                @click="!loading && handleNewConversation(item.id)"
              >
                <div
                  class="conversationpw-container"
                  role="button"
                  :title="
                    loading
                      ? 'Creating conversation...'
                      : 'Click on user to start a conversation'
                  "
                >
                  <img
                    v-if="item.photo"
                    :src="item.photo"
                    alt="ProfilePicture"
                  />
                  <div class="conv-name noto-color-emoji-regular">
                    {{ item.name }}
                  </div>
                </div>
              </li>
            </template>
          </FancyList>
        </template>
      </Search>
    </div>
  </BaseModal>
</template>
