<script>
import HeaderDefault from "./HeaderDefault.vue";
export default {
  components: {
    HeaderDefault,
  },
  props: ["userId", "groupId"],
  emits: ["sendGroupUp"],
  data() {
    return {
      group: null,
    };
  },
  mounted() {
    this.getGroupInfo();
  },
  methods: {
    async getGroupInfo() {
      try {
        const response = await this.$axios.get(`/groups/${this.groupId}`, {
          headers: {
            Authorization: `Bearer ${this.userId}`,
          },
        });
        this.group = response.data;
        this.group.id = this.groupId;
      } catch (error) {
        console.error(error);
      }
    },
    sendGroupUp(group) {
      sessionStorage.setItem("group", JSON.stringify(group));
      console.log("Saved:", group);
    },
  },
};
</script>

<template>
  <RouterLink :to="'/group'" @click="sendGroupUp(group)">
    <HeaderDefault v-if="group" class="btn">
      <div class="username" title="Group Name">
        {{ group.name }}
      </div>
      <span>CLICK ME TO MODIFY GROUP</span>
      <div class="group_description" :title="'Description: ' + group.desc">
        {{ group.desc }}
      </div>
    </HeaderDefault>
  </RouterLink>
</template>
