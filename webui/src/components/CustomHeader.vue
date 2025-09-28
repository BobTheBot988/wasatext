<script>

import HeaderGroup from './HeaderGroup.vue';
import HeaderUser from './HeaderUser.vue';

export default {
    components: {
        HeaderGroup,
        HeaderUser,
    },
    props: ['groupId', 'userId', 'addUser'],
    emits: ['sendGroupUp'],
    methods: {
        sendGroupUp(group) {
            this.$emit("sendGroupUp", group)
        }
    }
}
</script>
<template>
  <HeaderGroup v-if="groupId" :user-id="userId" :group-id="groupId" @send-group-up="sendGroupUp" />
  <HeaderUser v-else />
</template>