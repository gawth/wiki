
var markup = `
	<div>
		<label for="checkbox">Published </label>
		<input type="checkbox" id="checkbox" v-model="published"></input>
	</div>
	`

export default {
	template: markup,
	props: [
		'flag'
		],
	data() {
		return {
			published: this.flag
		}
	},
	watch: {
		published: 'handleFlag',
		flag: function(newFlag) { this.published = newFlag }
	},
	methods: {
		handleFlag: function() {
			console.log("Published: " + this.published);
			if (this.flag != this.published) {
				console.log("Calling savepubflag");
				this.$emit('savepubflag', this.published);
			}
		}
	}

}
