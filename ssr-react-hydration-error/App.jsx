export default function App() {
    // Bug: this evaluates to "server" on the server and to "client" in the browser
    const renderedBy = typeof window === 'undefined' ? 'server' : 'client';

    return <p>Rendered by: {renderedBy}</p>;
}
