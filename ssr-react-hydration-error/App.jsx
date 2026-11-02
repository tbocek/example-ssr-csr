export default function App() {
    const renderedBy = typeof window === 'undefined' ? 'server' : 'client';
    return <p>Rendered by: {renderedBy}</p>;
}
