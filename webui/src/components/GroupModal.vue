<script>
import BaseModal from "./BaseModal.vue";
import Search from "./Search.vue";
import FancyList from "./FancyList.vue";
import { errorManager } from "../services/axios";

export default {
  components: {
    BaseModal,
    Search,
    FancyList,
  },
  props: ["val", "title", "desc", "userId"],
  emits: ["userAdded"],
  data() {
    return {
      newName: "",
      users: [],
      non_users: [],
      ultimate_list: [],
      search: "",
      group:
        sessionStorage.getItem("group") !== null
          ? JSON.parse(sessionStorage.getItem("group"))
          : null,
      loading: false,
    };
  },
  computed: {
    filteredNonUsers() {
      if (this.ultimate_list === null) {
        return [];
      }
      return this.ultimate_list.filter((user) =>
        user.name.toLowerCase().includes(this.search.toLowerCase())
      );
    },
    displayItems() {
      return this.filteredNonUsers;
    },
  },
  created() {},

  async mounted() {
    console.log("Mounted Group modal component");
    await this.getUsers();
    await this.getNonUsers();
  },
  unmount() {
    if (this.group !== null) {
      sessionStorage.removeItem("group");
    }
  },
  updated() {
    console.log("Updated Group modal component");
  },

  methods: {
    async getItemPhoto(item) {
      try {
        this.loading = true;
        let response = null;

        console.log(item);
        if (item.groupId !== 0) {
          response = await this.$axios.get(`/groups/${item.id}/photo`, {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          });
        } else {
          response = await this.$axios.get(`/users/${item.id}/photo`, {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          });
        }

        return URL.createObjectURL(response.data);
      } catch (error) {
        console.log(error);
        errorManager.addError(error);
      } finally {
        this.loading = false;
      }
    },

    async getNonUsers() {
      try {
        this.loading = true;
        const response = await this.$axios.get("/users", {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
        });
        if (this.ultimate_list !== null && this.ultimate_list.length > 0) {
          this.ultimate_list.forEach((element) => {
            URL.revokeObjectURL(element.photo);
          });
        }
        this.non_users = response.data;

        if (this.users !== null) {
          this.ultimate_list = this.non_users.filter(
            (obj1) => !this.users.some((obj2) => obj1.id === obj2.id)
          );
        }

        for (const element of this.ultimate_list) {
          element.groupId = 0;
          element.photo = await this.getItemPhoto(element);
        }
      } catch (error) {
        errorManager.addError(error);
      } finally {
        this.loading = false;
      }
    },
    async getUsers() {
      try {
        this.loading = true;
        console.log("User ID:", this.userId);
        const response = await this.$axios.get(
          `/groups/${this.group.id}/users`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        this.users = response.data;
        for (const user of this.users) {
          user.groupId = 0;
          user.photo = await this.getItemPhoto(user);
        }
      } catch (error) {
        console.error(error);
      } finally {
        this.loading = false;
      }
    },
    async changeName() {
      try {
        this.loading = true;
        const response = this.$axios.post(
          `/group/${this.group.id}/name`,
          {
            name: newName,
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );

        if (response.status === 204) {
          console.log("Group Name Changed");
        } else {
          throw new Error(response);
        }
      } catch (error) {
        console.error(error);
      } finally {
        this.loading = false;
      }
    },
    async addUser(userId) {
      try {
        this.loading = true;
        const response = await this.$axios.post(
          `/groups/${this.group.id}/users/${userId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          await this.getUsers();
          await this.getNonUsers();
          this.handleUserAdded();
        }
      } catch (error) {
        errorManager.addError(
          "Error while adding user to group" + error.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    handleUserAdded() {
      this.$refs.modal.closeModal();
      this.$emit("userAdded");
    },
    handleFilter(val) {
      this.search = val;
    },
  },
};
</script>
<template>
  <BaseModal ref="modal" :val="val" :title="title">
    <LoadingSpinner :loading="loading" />
    <template #title>
      <div class="username" title="Group Name">{{ group.name }}</div>
    </template>
    <Search
      :group-id="group.id"
      :placeholder="'Search for users...'"
      @filter="handleFilter"
      @user-added="handleUserAdded"
    >
      <template #search />
      <template #search-result>
        <FancyList :items="displayItems">
          <template #item="item">
            <li class="search-result-item">
              <div class="conversationpw-container" role="button">
                <img v-if="item.photo" :src="item.photo" alt="ProfilePicture" />
                <div class="conv-name noto-color-emoji-regular">
                  <span>
                    {{ item.name }}
                  </span>
                </div>
                <input
                  :disabled="loading"
                  class="btn"
                  type="button"
                  title="Add user to group"
                  value="➕"
                  @click="addUser(item.id)"
                />
              </div>
            </li>
          </template>
        </FancyList>
      </template>
    </Search>
  </BaseModal>
</template>
