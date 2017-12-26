var markup = `
	<div> 
		<h1>{{title}}</h1> 
		<div v-html="compiledMarkdown"></div>
	 </div> 
`
export default {
	template: markup,
	props: [
		'title',
		'wikimd'
	],
	data() {
		return {input: '# Main Title'}
	},
	computed: {
		compiledMarkdown: function() {
			return marked(this.wikimd)
		}
	},
	mounted: function() {
		console.log('Mounted called on ' + this.title);
		this.$emit('getwiki', this.title);
	}
}
