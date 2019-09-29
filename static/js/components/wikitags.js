var markup = `
	<div>
		<div> 
			<div v-for="t in cleanTags" 
				v-show="!focus"
				@click="setFocus()"
				id="output"
				class="wikitag">{{t}}
			</div>
		</div>
		<input  v-model="tagedit"
			@blur="focus = false""
			v-show="focus"
			v-on:keyup.enter="unsetFocus"
			ref="edit"
			id="input">
		</input>
	</div>
`


export default {
	template: markup,
	props: [
		'tags'
	],
	data() {
		return {
			focus: false,
			tagedit: ''
		}
	},
	computed: {
		cleanTags: function() {
			return this.tags.split(",")
		}
	},
	methods: {
    setFocus: function () {
      console.log("Tags in focus...")
      this.tagedit = this.tags;
      this.focus = true;

      // Save off a ref for the closure 
      // and then set the focus on the next tick
      // This is needed to allow vue to make the 
      // component visible before we give it focus
      var that = this;
      Vue.nextTick(function() {
        that.$refs.edit.focus();
      });
		},
    unsetFocus: function () {
      this.focus = false;
      this.saveTags();
		},
		saveTags: function() {
			if (this.tags != this.tagedit) {
				this.$emit('savetags', this.tagedit);
			}
		}
	}
}
