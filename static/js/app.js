import MainWikiContent from './components/mainwikicontent.js';
import WikiTags from './components/wikitags.js';
import WikiFlag from './components/wikiflag.js';


var wikiword = function() {
	return  {
		type: 'lang',
		regex: /\{\{([^\}^#]+)[#]*(.*)\}\}/g,
		//regex: /test2/g,
		replace: '<a href=\"/wiki/view/$1#$2\">$1</a>'
		// replace: 'success'
	};
};
showdown.extension('wikiword', wikiword);

var converter = new showdown.Converter(
	{
		extensions:['wikiword'],
		parseImgDimensions: true,
		simplifiedAutoLink: true,
		excludeTrailingPunctuationFromURLs: true,
		tables: true,
		tasklists: true,
		openLinksInNewWindow: true,
		emoji: true,
		strikethrough: true
	}
);

var app = new Vue({
	el: '#app',
	delimiters: ['${', '}'],
	components: {
		MainWikiContent,
		WikiTags,
		WikiFlag
	},
	data: {
		title: '',
		wikimd: { "Body":"# Temp Heading", "Tags":"", "Published":false}
	},
	methods: {
		convertMarkdown: function(md) {
			// return marked(md);
			return converter.makeHtml(md);
		},
		getwiki(wiki) {
			this.title = wiki;
			axios.get('/api?wiki=' + wiki)
				.then(response => {
					this.wikimd = response.data;
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
			this.wikimd.Tags = tags;
			this.savewiki()
		},
		saveflag(vals) {
			var flag, val;
			[flag, val] = vals;
			if (flag === 'Published') {
				this.wikimd.Published = val;
			} else if (flag === 'Encrypted') {
				this.wikimd.Encrypted = val;
			}
			this.savewiki()
		},
		savewiki() {
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

