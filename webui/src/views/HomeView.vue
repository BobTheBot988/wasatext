<script>
import { errorManager } from '../services/axios';

export default {
  data: function () {
    return {
      errormsg: null,
      loading: false,
      some_data: null,
    };
  },
  mounted() {
    this.refresh();
    this.$forceUpdate();
  },
  methods: {
    async refresh() {
      this.loading = true;
      this.errormsg = null;
      try {
        let response = await this.$axios.get("/");
        this.some_data = response.data;
      } catch (e) {
        errorManager.addError(e.toString())
      } finally {
        this.loading = false;
      }
    },
  },
  async getUsers() {
    try {
      this.loading = true
      const request = await this.$axios.get("/users")
    } catch (e) {
      errorManager.addError(e.toString())
    } finally {
      this.loading = false
    }
  },
  newItem() {
    try {
      const users = this.getUsers()
    } catch (e) {

    }
  },
};
</script>

<template>
  <div>
    <div
      class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"
    >
      <h1 class="h2">Home page</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">
            Refresh
          </button>
        </div>
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-primary" @click="newItem">
            New
          </button>
        </div>
      </div>
    </div>

    <ErrorMsg v-if="errormsg" :msg="errormsg" />
    <div v-else>Select a conversation</div>
  </div>
</template>

<style></style>
