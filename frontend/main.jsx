import React, { useState, useEffect } from 'react';
import ReactDOM from 'react-dom/client';
import axios from 'axios';

// Настройка axios
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/',
});

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Основной компонент приложения
function App() {
  const [user, setUser] = useState(null);
  const [page, setPage] = useState('articles');
  const [articles, setArticles] = useState([]);
  const [formData, setFormData] = useState({ 
    login: '', password: '', 
    text: '', title: '' 
  });
  const [error, setError] = useState('');

  // Проверка аутентификации
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        setUser({ login: payload.login });
      } catch (e) {
        console.error('Invalid token');
      }
    }
    fetchArticles();
  }, []);

  const fetchArticles = async () => {
    try {
      const response = await api.get('/article');
      setArticles(response.data);
    } catch (err) {
      setError('Ошибка загрузки статей');
    }
  };

  // Обработчики форм
  const handleLogin = async (e) => {
    e.preventDefault();
    try {
      const response = await api.post('/login', {
        login: formData.login,
        password: formData.password
      });
      localStorage.setItem('token', response.data.token);
      setUser({ login: formData.login });
      setPage('articles');
    } catch (err) {
      setError('Неверные учетные данные');
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    if (formData.password.length < 8) {
      setError('Пароль должен быть не менее 8 символов');
      return;
    }
    
    try {
      await api.post('/register', {
        login: formData.login,
        password: formData.password
      });
      setError('Регистрация успешна! Теперь войдите');
      setPage('login');
    } catch (err) {
      setError(err.response?.data || 'Ошибка регистрации');
    }
  };

  const handleCreateArticle = async (e) => {
    e.preventDefault();
    try {
      await api.post('/article', { text: formData.text });
      setFormData({ ...formData, text: '' });
      fetchArticles();
      setError('');
    } catch (err) {
      setError('Ошибка создания статьи');
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setUser(null);
    setPage('articles');
  };

  // Рендер страниц
  const renderContent = () => {
    switch (page) {
      case 'login':
        return (
          <div className="card">
            <h2>Вход</h2>
            {error && <div className="error">{error}</div>}
            <form onSubmit={handleLogin}>
              <div className="form-group">
                <input 
                  type="text" 
                  placeholder="Логин" 
                  value={formData.login}
                  onChange={(e) => setFormData({...formData, login: e.target.value})}
                  required
                />
              </div>
              <div className="form-group">
                <input 
                  type="password" 
                  placeholder="Пароль" 
                  value={formData.password}
                  onChange={(e) => setFormData({...formData, password: e.target.value})}
                  required
                />
              </div>
              <button type="submit">Войти</button>
              <p style={{ marginTop: '10px' }}>
                Нет аккаунта? <a href="#" onClick={() => setPage('register')}>Зарегистрируйтесь</a>
              </p>
            </form>
          </div>
        );
      
      case 'register':
        return (
          <div className="card">
            <h2>Регистрация</h2>
            {error && <div className="error">{error}</div>}
            <form onSubmit={handleRegister}>
              <div className="form-group">
                <input 
                  type="text" 
                  placeholder="Логин (мин. 5 символов)" 
                  value={formData.login}
                  onChange={(e) => setFormData({...formData, login: e.target.value})}
                  minLength={5}
                  required
                />
              </div>
              <div className="form-group">
                <input 
                  type="password" 
                  placeholder="Пароль (мин. 8 символов)" 
                  value={formData.password}
                  onChange={(e) => setFormData({...formData, password: e.target.value})}
                  minLength={8}
                  required
                />
              </div>
              <button type="submit">Зарегистрироваться</button>
              <p style={{ marginTop: '10px' }}>
                Уже есть аккаунт? <a href="#" onClick={() => setPage('login')}>Войдите</a>
              </p>
            </form>
          </div>
        );
      
      case 'create-article':
        return (
          <div className="card">
            <h2>Создать статью</h2>
            {error && <div className="error">{error}</div>}
            <form onSubmit={handleCreateArticle}>
              <div className="form-group">
                <textarea 
                  placeholder="Текст статьи" 
                  value={formData.text}
                  onChange={(e) => setFormData({...formData, text: e.target.value})}
                  rows={6}
                  required
                />
              </div>
              <button type="submit">Опубликовать</button>
            </form>
          </div>
        );
      
      default: // Страница со статьями
        return (
          <>
            <div className="card">
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <h2>Последние статьи</h2>
                {user && (
                  <button onClick={() => setPage('create-article')}>
                    + Новая статья
                  </button>
                )}
              </div>
            </div>
            
            <div className="articles-grid">
              {articles.map(article => (
                <div key={article.id} className="article-card">
                  <h3>{article.author}</h3>
                  <p>{article.text}</p>
                </div>
              ))}
            </div>
          </>
        );
    }
  };

  return (
    <div>
      <header>
        <div className="container">
          <nav>
            <div className="logo">Блог на Go</div>
            <div className="nav-links">
              {user ? (
                <>
                  <span>Привет, {user.login}!</span>
                  <a href="#" onClick={() => setPage('articles')}>Статьи</a>
                  <a href="#" onClick={handleLogout}>Выйти</a>
                </>
              ) : (
                <>
                  <a href="#" onClick={() => setPage('articles')}>Статьи</a>
                  <a href="#" onClick={() => setPage('login')}>Вход</a>
                  <a href="#" onClick={() => setPage('register')}>Регистрация</a>
                </>
              )}
            </div>
          </nav>
        </div>
      </header>

      <div className="container">
        {renderContent()}
      </div>
    </div>
  );
}

// Рендер приложения
ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
