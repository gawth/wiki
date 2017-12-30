
var markup = `
	<div>
		<label for="checkbox">{{label}} </label>
		<input type="checkbox" id="checkbox" v-model="lclflag"></input>
	</div>
	`

export default {
	template: markup,
	props: [
		'flag',
		'label'
		],
	data() {
		return {
			lclflag: this.flag
		}
	},
	watch: {
		lclflag: 'handleFlag',
		flag: function(newFlag) { this.lclflag = newFlag }
	},
	methods: {
		handleFlag: function() {
			if (this.flag != this.lclflag) {
				this.$emit('saveflag', [this.label, this.lclflag]);
			}
		}
	}

}
