<template>
  <LoginForm :loading="loading" :error-msg="errormsg" @form-submit="handleFormSubmit" />
</template>

<script>
import LoginForm from './LoginForm.vue'
import { errorManager } from '../services/axios';

export default {
    name: 'LoginComponent',
    components: {
        LoginForm
    },
    emits: ['login-success', 'login-error'],
    data() {
        return {
            isLogged: localStorage.getItem("token"),
            errormsg: null,
            loading: false,
            some_data: null,
        }
    },
    methods: {
        async getUserPhoto(userId) {
            try {
                const response = await this.$axios.get(`/users/${userId}/photo`, {
                    "Authorization": `Bearer ${this.userId}`,
                    "responseType": "blob"
                })
                console.log(response.data)
                return response.data
            } catch (error) {
                console.error(error);
            }
        },

        async login(formData) {
            this.loading = true
            this.errormsg = null
            try {
                let response = await this.$axios.post("/session", {
                    name: formData.username
                })

                if (response.status === 200 || response.status === 201 || response.status === 202) {
                    if (!formData.wantLegacy) {
                        sessionStorage.setItem('token', response.data.userId)
                        sessionStorage.setItem('user', response.data)
                    } else {
                        localStorage.setItem('token', response.data.userId)
                        localStorage.setItem('user', response.data)
                    }
                    this.$emit("login-success")
                }
                else {
                    this.$emit("login-error")
                    errorManager.addError("Failed to login")
                }
            } catch (e) {
                errorManager.addError("Error while logging in" + e.toString())
            } finally {
                this.loading = false
            }
        },
        handleFormSubmit(formData) {
            this.login(formData)
            this.$emit('submit', formData.username)
        }
    }
}
</script>