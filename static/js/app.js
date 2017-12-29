import MainWikiContent from './components/mainwikicontent.js';
import WikiTags from './components/wikitags.js';
import Published from './components/pubflag.js';

var app = new Vue({
	el: '#app',
	delimiters: ['${', '}'],
	components: {
		MainWikiContent,
		WikiTags,
		Published
	},
	data: {
		title: '',
		wikimd: { "Body":"# Temp Heading", "Tags":"", "Published":false}
	},
	methods: {
		getwiki(wiki) {
			console.log('Getting data for : ' + wiki);
			this.title = wiki;
			axios.get('/api?wiki=' + wiki)
				.then(response => {
					this.wikimd = response.data;
					console.log("Got data : " + this.wikimd);
				})
				.catch(e => {
					console.log("ERROR: " + e);
					this.errors.push(e)
				})
		},
		setwiki(msg) {
			var title, body;
			[title, body] = msg;
			this.wikimd.Body = body;
			this.savewiki()
		},
		savetags(tags) {
			console.log("Save tags: " + tags);
			this.wikimd.Tags = tags;
			this.savewiki()
		},
		savepubflag(val) {
			console.log("Save pub flag: " + val);
			this.wikimd.Published = val;
			this.savewiki()
		},
		savewiki() {
			console.log("Saving wiki : " + this.title);
			axios.post('/api?wiki=' + this.title, this.wikimd)
				.then(response => {
					console.log("Saved :-)")
				})
				.catch(e => {
					console.log("ERROR: " + e);
					this.errors.push(e)
				})
		}
	}
});

