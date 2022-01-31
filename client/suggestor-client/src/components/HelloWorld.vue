<template>
  <div id="app">
    <!-- <p class="text-grey-dark mb-12">Using <a class="text-grey-dark" href="https://github.com/RomainSimon/vue-simple-search-dropdown" target="_blank">vue-simple-search-dropdown</a></p> -->
    <div class="row">
      <div class="multi">
        <multiselect
          v-model="selectedSearchBy"
          :multiple="false"
          :options="searchBys"
          placeholder="Search By"
        ></multiselect>
      </div>
      <div class="multi">
        <multiselect
          v-model="selectedSearchType"
          :multiple="false"
          :options="searchTypes"
          @input="assignFields"
          placeholder="Search Type"
        ></multiselect>
      </div>
      <!-- <div class="multi" v-if="fields.length>0">
        <multiselect
          v-model="value"
          :multiple="true"
          :options="fields"
          placeholder="Pick Fields"
        ></multiselect>
      </div> -->
    </div>
    <input v-model="message" @input="abcd" placeholder="search" />

    <div v-for="(opt, index) in options" :key="index">
      {{ opt.output }}
    </div>
  </div>
</template>

<script>
import axios from "axios";
import Multiselect from "vue-multiselect";

export default {
  components: {
    Multiselect,
  },
  name: "HelloWorld",
  props: {
    msg: String,
  },
  data() {
    return {
      options: [],
      searchBys: ["prefix", "infix","term"],
      searchTypes: ["address", "name"],
      value: [],
      message: "",
      info: [],
      fields: [],
      fieldValue: null,
      selectedSearchBy: "",
      selectedSearchType: "",
    };
  },
  methods: {
    assignFields(){
      if(this.selectedSearchType=='address'){
        this.fields=["city","state","country"]
      }
      if(this.selectedSearchType=='name'){
        this.fields=["fname","lname","middleName"]
      }
    },
    abcd() {
      console.log(
        "abcd: ",
        this.message,
        this.selectedSearchBy,
        this.selectedSearchType,
        this.value.toString()
      );
     // http://localhost:8001/autocomplete/search?index_name=suggestions&text=rash&searchBy=infix&searchType=address&fields=city,country
      axios
        .get(
          "http://localhost:8001/autocomplete/search?index_name=suggestions&text=" +
            this.message+"&searchBy="+this.selectedSearchBy+"&searchType="+this.selectedSearchType+"&fields="+this.value.toString()
        )
        .then((response) => (this.options = response.data || []));
    },
    getDropdownValues(keyword) {
      console.log("keyword", keyword);
      // axios
      //       .get('http://localhost:8001/autocomplete/search?index_name=suggestions&text='+keyword)
      //       .then(response => (this.options = response.data || []))
    },
  },
  mounted() {
    axios
      .get(
        "http://localhost:8001/autocomplete/search?index_name=suggestions&text=mah"
      )
      .then((response) => (this.info = response));
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
.multi {
  width: 20%;
  display: inline-flex;
}
</style>
<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
