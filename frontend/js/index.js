import React from 'react';
import ReactDOM from 'react-dom/client'; // Используйте версию client, если используете React >= 18.x
import './index.css';                   // Импортируйте стили, если они существуют
import App from './main';                // Здесь мы импортируем наш главный компонент

const rootElement = document.getElementById('root');   // Найдем элемент <div id="root"></div>, куда будем монтировать наше приложение

if (!rootElement) throw new Error('Root element not found!');

const root = ReactDOM.createRoot(rootElement);          // Создаем объект Root с помощью createRoot
root.render(<App />);                                   // Рендерим компонент App внутрь корня