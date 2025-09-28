<script>
export default {
    props: ['displayItems'],
    mounted() {
        this.handleOpenConvesation()
    },
    methods: {
        async handleOpenConvesation() {
            this.$emit("open-conversation")
            alert("Opening conversation")
        }
    }
}


</script>
<template>
  <FancyList :items="displayItems">
    <template #item="item">
      <li class="search-result-item" title="Click to add user" @click="addUser(item.id)">
        <div class="conversationpw-container" role="button">
          <img v-if="item.photo" :src="item.photo" alt="ProfilePicture">
          <div class="conv-name noto-color-emoji-regular">
            {{ item.name }}
          </div>
        </div>
      </li>
    </template>
  </FancyList>
</template>
