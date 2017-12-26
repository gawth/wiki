var markup = `
	<div> 
		<h1>{{title}}</h1> 
		 <textarea v-model="wikiedit" @blur="unsetFocus()" @input="update" v-show="focus" id="input" class="form-control"></textarea>
		<div v-html="compiledMarkdown" v-show="!focus" @click="setFocus()" id="output"></div>
	 </div> 
`
export default {
	template: markup,
	props: [
		'title',
		'wikimd'
	],
	data() {
		return {
			focus: false,
			wikiedit: ''
		}
	},
	computed: {
		compiledMarkdown: function() {
			return marked(this.wikimd)
		}
	},
	mounted: function() {
		console.log('Mounted called on ' + this.title);
		this.$emit('getwiki', this.title);
	},
	methods: {
		update: _.debounce(function (e) {
		    this.wikiedit = e.target.value
		}, 300),
	    setFocus: function () {
			this.wikiedit = this.wikimd;
		    this.focus = true;
		    document.getElementById('input').focus();
		},
	    unsetFocus: function () {
		    this.focus = false;
			this.$emit('setwiki', [this.title, this.wikiedit]);
		    document.getElementById('output').focus();
		}
	}
}
