<template>
  <div>
    <b-navbar toggleable="lg" type="dark" variant="info">
    <b-navbar-brand href="#">datAPI</b-navbar-brand>

    <b-navbar-toggle target="nav_collapse" />

     <!-- Right aligned nav items -->
      <b-navbar-nav class="ml-auto">
        <b-form-input 
          type="text" 
          v-model="email"
        />
        <b-form-input
          type="password"
          v-model="password"
        />
        <b-nav-item-dropdown right>
          <!-- Using button-content slot -->
          <template slot="button-content"><em>User</em></template>
          <b-dropdown-item href="#">Profile</b-dropdown-item>
          <b-dropdown-item href="#">Signout</b-dropdown-item>
        </b-nav-item-dropdown>
      </b-navbar-nav>
  </b-navbar>
    <div class="header">
 
      <b-button variant="success">Button</b-button>
    <button v-on:click="login()">
      login
    </button>
    <br/>
      <span class="small">Authorization: Bearer {{ token?token.slice(0,15) + "....":"anonymous" }}</span>
    <br/>
      {{ readToken }}
    </div>
    <div>
      bucket: <input type="text" v-model="bucket"/>
    </div>  
    <div class="query inner">
      <textarea v-model="queryBuffer"/><br/>
      <button class="send" v-on:click="sendQuery()" :disabled="bucket==''">execute query</button>
      </div>
    <div class="data inner">
      <textarea readonly v-model="readResponseBuffer"/>
    </div>
    <div class="putBuffer inner">
      <textarea v-model="putBuffer"/><br/>
      <button class="send" v-on:click="putQuery()" :disabled="bucket==''">send data</button>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

var client = axios.create(
  {
    headers: {
      'Content-Type': 'application/json'
    },
  }
)

export default {
  name: 'Main',
  data () {
    return {
      test: 'test',
      email: null,
      password: null,
      token: null,
      queryBuffer: "",
      responseBuffer: "",
      putBuffer: "",
      bucket: "",
    }
  },
  computed: {
    content: {
      get () {
        return JSON.stringify(this.queryReturn)
      },
      set () {

      }
    },
    readToken () {
      if (this.token) {
        var base64Url = this.token.split('.')[1];
        var base64 = base64Url.replace('-', '+').replace('_', '/');
        return JSON.parse(window.atob(base64));
      } else {
        return ""
      }
    },
    readResponseBuffer () {
      return JSON.stringify(this.responseBuffer, null,2)
    }
  },
  methods: {
    login () {
      var self = this
      let payload = {
        email: this.email,
        password: this.password
      }
      client.post("http://localhost:3000/login", payload).then(r => {
        self.token = r.data.token
        client.defaults.headers.common['Authorization'] = `Bearer ` + this.token
      }).catch(function() { 
        self.token = null 
        client.defaults.headers.common['Authorization'] = null
      })
    },
    sendQuery () {
      client.post("http://localhost:3000/data/get/" + this.bucket, this.queryBuffer).then(response => {
        this.responseBuffer = response.data
      }).catch(error => {
        this.responseBuffer = error.response
      })
    },
    putQuery () {
      let query = {}
      try {
        query = JSON.parse(this.putBuffer)
      }
      catch(error) {
        this.responseBuffer = "" + error
        return
      }
      client.post("http://localhost:3000/data/put/" + this.bucket, query).then(r => {
        this.responseBuffer = r.data
      }).catch(e => {
        this.responseBuffer = e.response
      })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.small {
  font-size: 15px;
}

div.inner {
  padding: 10px;
}
button.send{
  width: 80%;
}
textarea {
  height: 100px;
  width: 80%;
}
</style>
