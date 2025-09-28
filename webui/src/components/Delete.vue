<script>
import BaseModal from "./BaseModal.vue";

export default {
  components: {
    BaseModal,
  },
  props: ["title"],
  emits: ["handleDelete", "open", "close"],
  methods: {
    handleDelete() {
      this.$emit("handleDelete");
      this.$refs.modal.closeModal();
    },
  },
};
</script>
<template>
  <BaseModal
    ref="modal"
    :val="'🗑️'"
    :title="title"
    @close="$emit('close')"
    @opened="$emit('open')"
  >
    <template #title>
      <div>
        {{ title }}
      </div>
    </template>
    <template #default>
      <input
        title="Close Modal"
        type="button"
        value="No"
        @click="$refs.modal.closeModal()"
      >
      <input
        :title="title"
        type="button"
        value="Delete"
        @click="handleDelete()"
      >
    </template>
  </BaseModal>
</template>
