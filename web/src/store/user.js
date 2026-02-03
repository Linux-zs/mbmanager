import { defineStore } from 'pinia'
import { authAPI } from '../api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: null
  }),

  getters: {
    isLoggedIn: (state) => !!state.token
  },

  actions: {
    async login(username, password) {
      const res = await authAPI.login({ username, password })
      this.token = res.token
      localStorage.setItem('token', res.token)
      return res
    },

    async logout() {
      try {
        await authAPI.logout()
      } catch (error) {
        console.error('Logout error:', error)
      } finally {
        this.token = ''
        this.userInfo = null
        localStorage.removeItem('token')
      }
    }
  }
})
