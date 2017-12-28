var markup = `
		<div class="pure-u-22-24"> 
			<h1>{{title}}</h1> 
			<div class="pure-form">
				<textarea v-model="wikiedit" 
					@blur="unsetFocus()" 
					@input="update" 
					v-show="focus" 
					v-on:keydown.enter="handleEnter"
					v-on:keydown.83="handleCtrlS"
					id="input" 
					rows=20
					class="pure-input-1">
				</textarea>
			</div>
			<div v-html="compiledMarkdown" 
				v-show="!focus" 
				@click="setFocus()" 
				class="wikiBody"
				id="output">
			</div>
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
			this.saveWiki();
		    this.focus = false;
		    document.getElementById('output').focus();
		},
		handleEnter: function(e) {
			if (e.ctrlKey) {
				this.unsetFocus();
			}
		},
		handleCtrlS: function(e) {
			if (e.ctrlKey) {
				this.saveWiki();
			}
		},
		saveWiki: function() {
			if (this.wikiedit != this.wikimd) {
				this.$emit('setwiki', [this.title, this.wikiedit]);
			}
		}
	}
}
