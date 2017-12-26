var markup = `
	<div> 
		<h1>Single-file JavaScript Component</h1> 
		{{ msg }}
	 </div> 
`
export default {
	template: markup,
	data() {
		return {
			msg: 'Oh hai from the component' 
		}
	}
}
