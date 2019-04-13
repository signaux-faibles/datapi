<template>
  <div>
    <b-button v-b-modal.user variant="light" style="margin-right: 15px">
      <img src="@/assets/account-plus.svg"/>
    </b-button>
    <b-modal hide-footer id="user" title="Build an user object">
      <b-form-input id="inputEmail" type="text" v-model="email" placeholder="Email Address"/>
      <b-form-input id="inputPassword" type="password" v-model="password" placeholder="Password"/>
      <b-form-input id="scope" type="text" v-model="scope" placeholder="scope (coma separated tags)"/>
      <b-form-textarea id="object" rows=11 v-model="display"/>
    </b-modal>
  </div>
</template>

<script>
export default {
  name: 'User',
  data () {
    return {
      password: "",
      email: "",
      scope: "",
      alert: 0
    }
  },
  computed: {
    display () {
      var user = {
        "key": {
          "type": "credentials",
          "email": this.email
        },
        "value": {
          "password": this.password,
          "scope": (this.scope.trim() == "")?undefined:this.scope.split(',').map(s => s.trim())
        }
      }
      return JSON.stringify(user, null, 5)
    }
  },
}
</script>

<style scoped>
  input {
    margin: 5px;
  }
  textarea {
    font-size: 12px;
    margin: 5px;
  }
</style>