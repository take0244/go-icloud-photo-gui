import React, { useState } from 'react';
import { Login } from '@/src/LoginPage';
import { Code2Fa } from '@/src/Code2FaPage';
import { Photos } from '@/src/PhotosPage';

function App() {
  const [page, setPage] = useState('login');
  return (
    <div id="App">
      {page === 'login' && (<Login setPage={setPage} />)}
      {page === 'code2fa' && (<Code2Fa setPage={setPage} />)}
      {page === 'photos' && (<Photos setPage={setPage} />)}
    </div>
  )
}

export default App
