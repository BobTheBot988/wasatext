<script>
import BaseModal from './BaseModal.vue';
import Search from './Search.vue';
export default {
    components: {
        BaseModal,
        Search
    },
    methods: {
        add_user_to_group(user_id) {
            alert("Forwarding:", user_id)
        },
    }
}
</script>
<template>
  <Teleport to="modal-container">
    <BaseModal v-if="open">
      <Search @escape_user_id="add_user_to_group" />
    </BaseModal>
  </Teleport>
</template>