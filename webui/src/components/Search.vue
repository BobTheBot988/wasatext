<!-- eslint-disable vue/require-prop-types -->
<script>


export default {
    props: ["placeholder"],
    emits: ['escape_user_id', 'closeModal', 'escape_user_ids', 'filter'],
    data() {
        return {
            search: "",
        }
    },
    methods: {
        handleFilter() {
            this.$emit("filter", this.search)
        }
    }
}

</script>
<template>
  <div id="search-form">
    <input id="form-text" v-model="search" type="text" :placeholder="placeholder" @input="handleFilter">
    <slot name="search" />
  </div>

  <div id="search-result">
    <slot name="search-result" />
  </div>
</template>