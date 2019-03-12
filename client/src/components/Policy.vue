<template>
  <div>
    <b-button v-b-modal.policy variant="light" style="margin-right: 15px">
      <img src="@/assets/camera-iris.svg"/>
    </b-button>
    <b-modal hide-footer size="lg" id="policy" style="width: 80%" title="Policy object builder">
      <div style="float: left;width: 45%;">
        <b-form-input id="inputName" type="text" v-model="name" placeholder="Policy Name"/>
        <b-form-input id="inputMatch" type="text" v-model="match" placeholder="Buckets (re2 expression)"/>
        <b-form-input id="inputKey" type="text" v-model="key" placeholder="Key (map[string]string json)"/>
        <b-form-input id="inputScope" type="text" v-model="scope" placeholder="Limit policy to user scope (coma separated tags)"/>
        <b-form-input id="inputRead" type="text" v-model="read" placeholder="Add to read objects scope (coma separated tags)"/>
        <b-form-input id="inputWrite" type="text" v-model="write" placeholder="Add to written objects scope (coma separated tags)"/>
        <b-form-input id="inputPromote" type="text" v-model="promote" placeholder="Promote scope (coma separated tags)"/>
      </div>
      <div style="float: left;width: 45%;">
        <b-form-textarea rows=13 v-model="policy"/>
      </div>
    </b-modal>

  </div>
</template>

<script>
export default {
  name: 'Policy',
  data () {
    return {
      name: '',
      match: '',
      key: '',
      scope: '',
      read: '',
      write: '',
      promote: ''
    }
  },
  computed: {
    policy () {
      try {
      var key = JSON.parse(this.key)
      } catch {
        var key = ''
      }
      var builtpolicy = [{
        'key': {
          'type': 'policy',
          'name': this.name
        },
        'value': {
          'match': this.match,
          'key': key,
          'scope': (this.scope.trim() == "")?[]:this.scope.split(',').map(s => s.trim()),
          'read': (this.read.trim() == "")?[]:this.read.split(',').map(s => s.trim()),
          'write': (this.write.trim() == "")?[]:this.write.split(',').map(s => s.trim()),
          'promote': (this.promote.trim() == "")?[]:this.promote.split(',').map(s => s.trim())
        }
      }]
      return JSON.stringify(builtpolicy, null, 2)
    }
  }
}
</script>

<style scoped>
  input {
    width: 90%;
    margin: 5px;
  }
  textarea {
    font-size: 12px;
    width: 110%;
    margin: 5px;
  }
</style>