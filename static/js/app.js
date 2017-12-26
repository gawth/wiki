import MainWikiContent from './components/mainwikicontent.js';

var app = new Vue({
	el: '#app',
	delimiters: ['${', '}'],
	components: {
		MainWikiContent
	},
	data: {
		wikimd: '# This is a title'
	},
	methods: {
		getwiki(wiki) {
			console.log('Getting data for : ' + wiki);
			axios.get('/api?wiki=' + wiki)
				.then(response => {
					this.wikimd = response.data.Body;
				})
				.catch(e => {
					this.errors.push(e)
				})
		}
	}
});

