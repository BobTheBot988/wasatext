<script>
export default {
  data() {
    return {
      userId: localStorage.getItem("user")
        ? localStorage.getItem("user").userId
        : sessionStorage.getItem("user").userId,
      activeObjectURLs: new Set(),
    };
  },
  methods: {
    async getItemPhoto(item) {
      try {
        let response = null;
        if (item.groupId !== 0) {
          response = await this.$axios.get(`/groups/${item.groupId}/photo`, {
            headers: {
              Authorization: `Bearer ${this.userId}`,
            },
            responseType: "blob",
          });
        } else {
          if (item.userId) {
            response = await this.$axios.get(`/users/${item.userId}/photo`, {
              headers: {
                Authorization: `Bearer ${this.userId}`,
              },
              responseType: "blob",
            });
          } else if (item.id !== this.userId) {
            response = await this.$axios.get(`/users/${item.id}/photo`, {
              headers: {
                Authorization: `Bearer ${this.userId}`,
              },
              responseType: "blob",
            });
          }
        }

        return URL.createObjectURL(response.data);
      } catch (error) {
        console.error(error);
      }
    },
    async getUsers() {
      try {
        const response = await this.$axios.get(`/users`, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
        });
        let users = response.data;
        for (const user of users) {
          user.groupId = 0;
          user.photo = await this.getUserPhoto(user);
        }
        return users;
      } catch (error) {
        console.error(error);
        return [];
      }
    },

    createAndTrackObjectURL(blob) {
      const url = URL.createObjectURL(blob);
      activeObjectURLs.add(url);
      return url;
    },

    revokeAndUntrack(url) {
      URL.revokeObjectURL(url);
      activeObjectURLs.delete(url);
    },
    async getGroupUsers(groupId) {
      try {
        const response = await this.$axios.get(`/groups/${groupId}/users`, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
        });

        let users = response.data;
        for (const user of users) {
          user.groupId = 0;
          user.photo = await this.getUserPhoto(user);
        }

        return users;
      } catch (error) {
        console.error(error);
        return [];
      }
    },
  },
};
</script>
