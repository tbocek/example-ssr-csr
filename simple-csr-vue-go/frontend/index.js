const App = {
    data() {
        return { text: '', result: '' }
    },
    methods: {
        async convert() {
            const res = await fetch('/api/toupper', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text: this.text })
            })
            const data = await res.json()
            this.result = data.result
        }
    },
    template: `
        <h1>ToUpper</h1>
        <input v-model="text" @input="convert" placeholder="type here..." />
        <p>{{ result }}</p>
    `
}
Vue.createApp(App).mount('#app')
