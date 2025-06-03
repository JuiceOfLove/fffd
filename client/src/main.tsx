import { createContext, StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './assets/style/global.css'
import Store from './store/store';
import Router from './components/router/Router';

const store = new Store();

export const Context = createContext({
  store,
})

createRoot(document.getElementById('root')!).render(
  <Context.Provider value={{
    store
  }}>
    <StrictMode>
      <Router />
    </StrictMode>
  </Context.Provider>
)
