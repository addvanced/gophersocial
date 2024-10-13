import './App.css'

export const API_URL = import.meta.env.VITE_API_URL as string || 'http://localhost:8080/v1'

function App() {

  return (
    <>
      <div>App home screen</div>
    </>
  )
}

export default App
