import axios from "axios";

import { reactive } from 'vue'

// Global error state
export const errorState = reactive({
  errors: []
})

// Error management functions
export const errorManager = {
  addError(message, id = null) {
    const errorId = id || Date.now() + Math.random()
    errorState.errors.push({ id: errorId, message })
    return errorId
  },
  
  removeError(id) {
    const index = errorState.errors.findIndex(error => error.id === id)
    if (index > -1) {
      errorState.errors.splice(index, 1)
    }
  },
  
  clearErrors() {
    errorState.errors = []
  },
  
  hasErrors() {
    return errorState.errors.length > 0
  }
}

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

// Request interceptor
instance.interceptors.request.use(
  (config) => {
    // Clear previous errors on new request (optional)
    if (config.clearErrors !== false) {
      errorManager.clearErrors()
    }
    
    // Add auth token
    /*const userId = localStorage.getItem("user") ? 
      JSON.parse(localStorage.getItem("user")).userId :
      JSON.parse(sessionStorage.getItem("user")).userId
    
    if (userId) {
      config.headers.Authorization = `Bearer ${userId}`
    }*/
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
instance.interceptors.response.use(
  (response) => response,
  (error) => {
    let errorMessage = 'An unexpected error occurred'
    
    if (error.response) {
      // Server responded with error status
      const { status, data } = error.response
      
      switch (status) {
        case 400:
          errorMessage = data.message || 'Bad request. Please check your input.'
          break
        case 401:
          errorMessage = 'Unauthorized. Please login again.'
          // Optionally redirect to login
          // window.location.href = '/login'
          break
        case 403:
          errorMessage = 'Access forbidden. You don\'t have permission to perform this action.'
          break
        case 404:
          errorMessage = 'Resource not found.'
	  //	  window.location.href = '/notfound'
          break
        case 409:
          errorMessage = data.message || 'Conflict. The resource already exists.'
          break
        case 422:
          // Validation errors
          if (data.errors) {
            // Handle multiple validation errors
            Object.values(data.errors).flat().forEach(msg => {
              errorManager.addError(msg)
            })
            return Promise.reject(error)
          } else {
            errorMessage = data.message || 'Validation failed.'
          }
          break
        case 500:
          errorMessage = data.message || 'Server error. Please try again later.'
          break
        default:
          errorMessage = data.message || `Server error (${status})`
      }
    } else if (error.request) {
      // Network error
      errorMessage = 'Network error. Please check your internet connection.'
    } else if (error.code === 'ECONNABORTED') {
      // Timeout error
      errorMessage = 'Request timeout. Please try again.'
    }
    
    // Add error to global state
    errorManager.addError(errorMessage)
    
    return Promise.reject(error)
  }
)


export default instance;
