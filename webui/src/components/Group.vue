<script>
import FancyList from "./FancyList.vue";
import GroupModal from "./GroupModal.vue";
import LoadingSpinner from "./LoadingSpinner.vue";
import Search from "./Search.vue";
import { errorManager } from "../services/axios";

export default {
  components: {
    GroupModal,
    Search,
    FancyList,
    LoadingSpinner,
  },
  data() {
    return {
      new_group_name: "",
      local_group: null,
      new_description: "",
      non_users: [],
      users: [],
      ultimate_list: [],
      search: "",
      test: [],
      del: 0,
      p: 0,
      userId: localStorage.getItem("user")
        ? JSON.parse(localStorage.getItem("user")).userId
        : JSON.parse(sessionStorage.getItem("user")).userId,
      loading: false,
    };
  },
  computed: {
    filteredUsers() {
      if (this.users === null) {
        return [];
      }
      return this.users.filter((user) =>
        user.name.toLowerCase().includes(this.search.toLowerCase())
      );
    },

    displayItems() {
      return this.filteredUsers.slice().sort((a, b) => {
        if (a.name < b.name) {
          return -1;
        }
        if (a.name > b.name) {
          return 1;
        }
        return 0;
      });
    },
  },
  async beforeMount() {
    // component is rendered as part of the initial request
    // pre-fetch data on server as it is faster than on the client
    this.local_group =
      sessionStorage.getItem("group") !== null
        ? JSON.parse(sessionStorage.getItem("group"))
        : null;
    console.log("Getting group", this.local_group);
    this.local_group.picture = await this.getItemPhoto();

    console.log("Getting users");
    await this.getUsers();
    console.log("The users are", this.users);
  },

  unmount() {
    if (this.local_group !== null) {
      URL.revokeObjectURL(local_group.picture);
      sessionStorage.removeItem("group");
    }
  },

  methods: {
    async getUsers() {
      try {
        this.loading = true;
        const response = await this.$axios.get(
          `/groups/${this.local_group.id}/users`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        console.log("Response:", response.data);
        if (response.data.length === 0) {
          this.users = [];
          return;
        }
        if (this.users !== null && this.users.length > 0) {
          this.users.forEach((user) => {
            URL.revokeObjectURL(user.photo);
          });
        }
        this.users = response.data;
        console.log("Users:", this.users);

        for (const user of this.users) {
          user.groupId = 0;
          user.photo = await this.getUserPhoto(user);
        }

        console.log("Updated Users:", this.users);
        this.test = this.users;
      } catch (error) {
        errorManager.addError(
          "Error while getting users for group" + error.toString()
        );
      } finally {
        this.loading = false;
      }
    },
    async getItemPhoto() {
      try {
        this.loading = true;
        let response = null;
        response = await this.$axios.get(
          `/groups/${this.local_group.id}/photo`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          }
        );
        this.p += 1;
        return URL.createObjectURL(response.data);
      } catch (error) {
        console.log(error);
      } finally {
        this.loading = false;
      }
    },

    async getUserPhoto(user) {
      try {
        this.loading = true;
        const response = await this.$axios.get(`/users/${user.id}/photo`, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
          responseType: "blob",
        });
        return URL.createObjectURL(response.data);
      } catch (error) {
        console.error(error);
        return null; // or a default image URL
      } finally {
        this.loading = false;
      }
    },
    checkName() {
      return this.new_group_name.includes(" ");
    },
    async changeGroupDesc() {
      try {
        this.loading = true;
        if (this.checkName()) {
          throw new Error("group name has spaces");
        }
        console.log("new desc " + this.new_description + "\n");

        const response = await this.$axios.post(
          `/groups/${this.local_group.id}/desc`,
          {
            id: this.local_group.id,
            name: null,
            picture: null,
            desc: this.new_description,
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          this.local_group.desc = this.new_description;
          console.log("The group name has been changed:");
          this.new_description = "";
        }
      } catch (error) {
        console.error(error);
      } finally {
        this.loading = false;
      }
    },
    async changeGroupName() {
      try {
        this.loading = true;
        if (this.checkName()) {
          throw new Error("group name has spaces");
        }
        const response = await this.$axios.post(
          `/groups/${this.local_group.id}/name`,
          {
            id: this.local_group.id,
            name: this.new_group_name,
            picture: null,
          },
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          this.local_group.name = this.new_group_name;
          console.log("The group name has been changed:");
          this.new_group_name = "";
        }
      } catch (error) {
        console.error(error);
      } finally {
        this.loading = false;
      }
    },
    async changeGroupImage() {
      try {
        this.loading = true;
        // Get the file input element
        const fileInput = document.querySelector('input[name="fileUp"]');
        const file = fileInput.files[0];
        console.log(file);

        // Check if a file was selected
        if (!file) {
          throw new Error("No file selected");
        }

        // Check file type (only allow PNG as specified in your input accept attribute)
        if (file.type !== "image/png") {
          throw new Error("Only PNG files are allowed");
        }

        // Create FormData to send the file
        const formData = new FormData();
        formData.append("photo", file);

        // Based on backend API: /photos/{photoId}/users/{userId}/photo
        // We need to provide a photoId and userId in the URL
        // For a new user photo upload, we can use a default photoId (or you might need to adjust this)
        const photoId = 0; // Replace with actual photoId if needed

        // Send the request to the server
        const response = await this.$axios.post(
          `/photos/${photoId}/groups/${this.local_group.id}/photo`,
          formData,
          {
            headers: {
              "Content-Type": "multipart/form-data",
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );

        // Handle successful response
        if (response.status === 200) {
          // Clear the file input
          fileInput.value = "";

          // Update the profile picture display
          this.local_group.picture = await this.getItemPhoto();
          console.log("The profile picture has been changed");
        }
      } catch (error) {
        let e = "Error changing profile picture:" + error.toString();
        errorManager.addError(e);
      } finally {
        this.loading = false;
      }
    },
    handleFilter(new_search) {
      this.search = new_search;
    },
    async handleDelete(userId) {
      try {
        this.loading = true;
        const response = await this.$axios.delete(
          `/groups/${this.local_group.id}/users/${userId}`,
          {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
          }
        );
        if (response.status === 204) {
          console.log("user_deleted");
          this.del += 1;
          if (this.userId === userId && this.local_group !== null) {
            sessionStorage.removeItem("group");
            this.$router.go(-1);
            return;
          }
        }
        await this.getUsers();
      } catch (error) {
        console.error(error);
        errorManager.addError(error);
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
<template>
  <div>
    <LoadingSpinner :loading="loading" />
    <div id="group-info-header">
      <img
        v-if="local_group.picture"
        id="user-profile-picture"
        :key="p"
        :src="local_group.picture"
        alt="profile-picture"
      />

      <label class="custom-file-upload btn-high-importance" title="Send image">
        Change Group Picture
        <input
          :disabled="loading"
          type="file"
          name="fileUp"
          accept=".png"
          @input="changeGroupImage"
        />
      </label>
    </div>
    <div id="container-group-name">
      {{ local_group.name }}
      <input
        v-model="new_group_name"
        :disabled="loading"
        type="text"
        placeholder="Change Group Name..."
        @keyup.enter="changeGroupName"
      />
      <input
        :disabled="loading"
        class="modify-info btn"
        title="Change group name"
        type="button"
        value="🖊️"
        @click="changeGroupName"
        @keyup.enter="changeGroupName"
      />
    </div>
    <div id="container-group-description">
      {{ local_group.desc }}
      <input
        v-model="new_description"
        :disabled="loading"
        type="text"
        placeholder="Change Group Description..."
        @keyup.enter="changeGroupDesc"
      />
      <input
        :disabled="loading"
        class="modify-info btn"
        title="Change group description"
        type="button"
        value="🖊️"
        @click="changeGroupDesc"
        @keyup.enter="changeGroupDesc"
      />
    </div>
    <div class="footer">
      <GroupModal
        :key="del"
        :user-id="userId"
        add-user="true"
        :group="local_group"
        :val="'➕'"
        :title="'Add Users'"
        @user-added="getUsers()"
      />
      <Search
        :ref="search"
        name="search"
        :placeholder="'Search For Users in the group'"
        @filter="handleFilter"
      >
        <template #search />
        <template #search-result>
          <FancyList :items="displayItems">
            <template #item="item">
              <li class="search-result-item">
                <div class="conversationpw-container" role="button">
                  <img
                    v-if="item.photo"
                    :src="item.photo"
                    alt="ProfilePicture"
                  />
                  <div class="conv-name noto-color-emoji-regular">
                    {{ item.name }}
                  </div>
                  <input
                    :disabled="loading"
                    class="btn"
                    type="button"
                    title="Remove user from group"
                    value="🗑️"
                    @click="handleDelete(item.id)"
                  />
                </div>
              </li>
            </template>
          </FancyList>
        </template>
      </Search>
    </div>
  </div>
</template>
