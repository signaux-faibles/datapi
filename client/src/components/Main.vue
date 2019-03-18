<template>
  <div>
    <b-navbar toggleable="lg" type="dark" variant="primary">
    <b-navbar-brand class="title" href="#">datapi-admin</b-navbar-brand>
    <b-navbar-toggle target="nav_collapse" />
      <b-navbar-nav class="ml-auto" >
        <div v-if="token==null">
          <b-button v-b-modal.modal1 variant="light">Sign in</b-button>
          <b-modal hide-footer id="modal1"  title="Login">
            <b-form-input id="inputEmail" type="text" v-model="email" placeholder="Email Address"/>
            <b-form-input id="inputPassword" type="password" v-model="password" placeholder="Password"/>
            <b-button  class="btn btn-lg btn-primary btn-block"  v-on:click="login()">sign in</b-button>
          </b-modal>
        </div>
        <User v-if="token!=null"/>
        <Policy v-if="token!=null"/>
        <div class="ml-auto" v-if="token!=null">
           <b-button variant="light" v-on:click="logout()">Logout</b-button>
        </div>
      </b-navbar-nav>
    </b-navbar>
    <p/>
    <div>
      bucket <input type="text" v-model="bucket"/>
    </div>  
    <div class="container">
      <div class="row">
        <div class="col-6">

          <div class="query inner">
            <b-form-textarea rows="5" v-model="queryBuffer"/><br/>
            <b-button v-on:click="sendQuery()" :disabled="bucket==''">execute query</b-button>
          </div>
          <div class="putBuffer inner">
            <b-form-textarea rows="15" v-model="putBuffer"/><br/>
            <b-button v-on:click="putQuery()" :disabled="bucket==''">send objects</b-button>
          </div>
        </div>
        <div class="col-6">
          Server output
          <div class="data inner">
            <b-form-textarea width="100%" rows="25" readonly v-model="readResponseBuffer"/>
          </div>
        </div>
      </div>
    </div>

    <br/>
    <span class="small">{{ token?"signed in":"please sign in" }}</span>
    <br/>
    {{ readToken.email }}
    <br/>
    {{ readToken.scope }}
    


  </div>
</template>

<script>
import axios from 'axios'
import Policy from '@/components/Policy'
import User from '@/components/User'

var client = axios.create(
  {
    headers: {
      'Content-Type': 'application/json'
    },
  }
)

export default {
  name: 'Main',
  components: {Policy, User},
  data () {
    return {
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
        return {}
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
    logout () {
      this.token = null 
      client.defaults.headers.common['Authorization'] = null
      this.responseBuffer = ""
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
.title {
  font-family: "Good Times";
}
div.inner {
  padding: 10px;
}
button.send{
  width: 80%;
}
textarea  
{  
   font-family:"monospace";  
   font-size: 10px;   
}
</style>
