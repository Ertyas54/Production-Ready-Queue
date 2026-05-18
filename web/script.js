const API_URL = window.location.origin;

async function sendOrder(type) {
    const resultElement = document.getElementById('result');
    resultElement.textContent = 'Отправка...';

    try {
        const response = await fetch(`${API_URL}/api/send/${type}`, {
            method: 'POST'
        });
        const data = await response.json();
        resultElement.textContent = data.message;
    } catch (error) {
        resultElement.textContent = 'Ошибка соединения с сервером';
    }
}