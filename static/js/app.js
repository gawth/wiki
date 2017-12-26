import MainWikiContent from './components/mainwikicontent.js';

var app = new Vue({
	el: '#app',
	delimiters: ['${', '}'],
	components: {
		MainWikiContent
	},
	data: {
		  mainmsg: 'Hello Vue!'
		}
});

