<script>
import BaseModal from "./BaseModal.vue";
import FancyList from "./FancyList.vue";
import LoadingSpinner from "./LoadingSpinner.vue";
import Search from "./Search.vue";
import { errorManager } from "../services/axios";

export default {
  components: {
    BaseModal,
    Search,
    FancyList,
    LoadingSpinner,
  },
  props: ["open", "userId"],
  data() {
    return {
      step_one: true,
      group_name: "",
      search: "",
      users: [],
      new_users: new Set([this.userId]),
      loading: false,
    };
  },
  computed: {
    filteredUsers() {
      if (this.users === null) {
        return [];
      }
      return this.users.filter(
        (user) =>
          user.id !== this.userId &&
          user.name.toLowerCase().includes(this.search.toLowerCase())
      );
    },
    displayItems() {
      return this.filteredUsers;
    },
  },
  mounted() {
    this.step_one = true;
  },
  unmounted() {
    this.step_one = true;
  },
  beforeUpdate() {
    this.getUsers();
  },
  methods: {
    addUser(user_id) {
      const li_user = document.getElementById(user_id);
      if (this.new_users) {
        if (this.new_users.has(user_id)) {
          li_user.classList.remove("chosen");
          this.new_users.delete(user_id);
          return;
        }
      }

      this.new_users.add(user_id);
      if (li_user !== null) {
        li_user.classList.remove("chosen");
        li_user.classList.add("chosen");
      }
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
        console.log(error);
        errorManager.addError(error);
      } finally {
        this.loading = false;
      }
    },
    add_user_to_group(user_id) {
      alert("Adding user:", user_id);
    },
    async getUsers() {
      try {
        console.log("Getting users");
        this.loading = true;
        const response = await this.$axios.get("/users", {
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

        // Process photos sequentially instead of concurrently
        for (const user of this.users) {
          user.groupId = 0;
          user.photo = await this.getItemPhoto(user);
        }
      } catch (error) {
        errorManager.addError(
          "Error while getting users for groupmodal",
          error.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    async create_new_group() {
      this.loading = true;
      const neu_users = Array.from(this.new_users);
      try {
        let response = await this.$axios.post(
          `/groups`,
          {
            userIdList: neu_users,
            name: this.group_name,
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        this.some_data = response.data;
        console.log("Response:", response);
        if (response.status === 204) {
          // There might be a better way to do this. e.g. key
          this.$router.go("/conversations");
        }
      } catch (e) {
        errorManager.addError("Error while creating new group" + e.toString());
      } finally {
        this.loading = false;
      }
    },
    handleInputName() {
      this.group_name = this.group_name.trim();
      const len = this.group_name.length;
      if (len === 0) {
        console.warn("The name must not be empty");
        this.loading = false;
        this.group_name = "";
        errorManager.addError("The name must not be empty");
        this.$router.go();
        return;
      }
      this.step_one = false;
    },
    handleFilter(new_search) {
      this.search = new_search;
    },
  },
};
</script>
<template>
  <BaseModal :icon="'/feather-sprite-v4.29.0.svg#users'" :val="' New Group'">
    <template #title>
      <LoadingSpinner :loading="loading" />
      <div v-if="step_one">Insert Group Name</div>
      <div v-else>Choose the users</div>
    </template>
    <template v-if="step_one" #default>
      <div style="display: flex">
        <input
          id="group-name"
          v-model="group_name"
          :disabled="loading"
          type="text"
          style="width: 100%"
          placeholder="Insert Group Name"
          @keyup.enter="handleInputName"
        />
        <input
          type="button"
          :disabled="loading"
          value="➡️"
          class="btn"
          @click="handleInputName"
        />
      </div>
    </template>

    <Search
      v-if="!step_one"
      :placeholder="'Search Users'"
      @filter="handleFilter"
      @escape_user_ids="create_new_group"
    >
      <template #search>
        <div id="search-footer">
          <input
            type="button"
            title="Create Group"
            value="➡️"
            class="btn"
            :disabled="loading"
            @click="create_new_group"
          />
        </div>
      </template>
      <template #search-result>
        <FancyList :items="displayItems">
          <template #item="item">
            <li
              :id="item.id"
              class="search-result-item"
              title="Click to add user"
              @click="addUser(item.id)"
            >
              <div class="conversationpw-container" role="button">
                <img v-if="item.photo" :src="item.photo" alt="ProfilePicture" />
                <div class="conv-name noto-color-emoji-regular">
                  {{ item.name }}
                </div>
              </div>
            </li>
          </template>
        </FancyList>
      </template>
    </Search>
  </BaseModal>
</template>
