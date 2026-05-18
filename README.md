# Production-Ready Queue lab

Реализация паттерна Retry Queue + Dead Letter Queue на Go и RabbitMQ.

## Архитектура

Producer отправляет сообщение в Main Queue. Consumer Main пытается обработать. Если ошибка временная (сеть) - сообщение уходит в Retry Queue, где Consumer Retry делает до 3 повторных попыток. Если ошибка постоянная (валидация) или попытки исчерпаны - сообщение попадает в Dead Letter Queue для ручного разбора.

## Запуск

```
docker-compose up --build
```
## Сервисы

Веб-интерфейс: 
````
http://localhost:8082
````
RabbitMQ Management: 
````
http://localhost:15672 (guest/guest)
````

## Проверка работы

### 1. Успешная обработка
Нажать "Отправить корректный заказ".
#### Логи: 
````
queue-consumer-main   | 18:36:03 [MAIN CONSUMER] Получено: [ORDER-001] Новый заказ (retries=0)            
queue-consumer-main   | 18:36:03 [MAIN CONSUMER] Успешно обработан: ORDER-001
````

### 2. Retry Queue
Остановить consumer-retry:
```
docker-compose stop consumer-retry
```

Нажать "Отправить заказ с сетевой ошибкой". 
#### Логи:
````
queue-consumer-retry  | Статистика ретраев:                                                               
queue-consumer-retry  |   Всего ретраев: 2                                                                
queue-consumer-retry  |   Успешно после ретрая: 0                                                         
queue-consumer-retry  |   Исчерпаны попытки (→ DLQ): 1                                                    
queue-consumer-retry  |   В процессе: 1                                                                   
queue-consumer-retry  | ------------------------------------------------------------                      
queue-consumer-retry  | [RETRY] ORDER-002 | Заказ: RETRY-TEST | Попытка: 2/3 | Ошибка: временная ошибка сети: connection timeout (демонстрация Retry) | 18:37:46
queue-consumer-retry  | [FAILED] ORDER-002 | Заказ: RETRY-TEST | Попытка: 3/3 | Ошибка: временная ошибка сети: connection timeout (демонстрация Retry) | 18:37:47    
````
#### RabbitMQ
![Retry.jpg](../../../%D0%A3%D1%87%D0%B5%D0%B1%D0%B0/6%20%D1%81%D0%B5%D0%BC%D0%B5%D1%81%D1%82%D1%80/%D0%A2%D1%80%D0%BE%D0%B4%20%28%D0%90%D0%BA%D1%83%D1%82%D0%B8%D0%BD%29/%D0%A1%D0%BA%D1%80%D0%B8%D0%BD%D1%8B%204/Retry.jpg)

consumer-main получает заказ, фиксирует ошибку сети, отправляет в queue.retry. В RabbitMQ очередь queue.retry показывает Ready: 1. Сообщение ждёт повторной обработки. 
Запустить consumer-retry: 
```
docker-compose start consumer-retry 
```
#### RabbitMQ
![Retry after.jpg](../../../%D0%A3%D1%87%D0%B5%D0%B1%D0%B0/6%20%D1%81%D0%B5%D0%BC%D0%B5%D1%81%D1%82%D1%80/%D0%A2%D1%80%D0%BE%D0%B4%20%28%D0%90%D0%BA%D1%83%D1%82%D0%B8%D0%BD%29/%D0%A1%D0%BA%D1%80%D0%B8%D0%BD%D1%8B%204/Retry%20after.jpg)

### 3. Dead Letter Queue
Нажать "Отправить заказ с ошибкой валидации". 
#### Логи:
````
queue-consumer-main   | 18:41:30 [MAIN CONSUMER] Получено: [ORDER-003] Заказ с ошибкой валидации (тип ошибки #4) (retries=0)
queue-consumer-main   | 18:41:30 [MAIN CONSUMER] ORDER-003: ошибка валидации → DEAD LETTER QUEUE (валидация не пройдена: Quantity - должно быть больше нуля, получено: -5)                                          
queue-consumer-dead   | 18:41:30 [DEAD CONSUMER] МЕРТВОЕ СООБЩЕНИЕ: ORDER-003 | Требуется ручное вмешательство!    
````
````
queue-consumer-dead   | 18:41:35 ------------ Всего мертвых сообщений ---------------: 2                  
queue-consumer-dead   | 18:41:35 1. ORDER-002 | Тело: Заказ с сетевой ошибкой (демонстрация Retry) | Создано: 18:37:46 | Попыток: 3                                                                                 
queue-consumer-dead   | 18:41:35 2. ORDER-003 | Тело: Заказ с ошибкой валидации (тип ошибки #4) | Создано: 18:41:30 | Попыток: 1  
````
#### RabbitMQ
![Dead.jpg](../../../%D0%A3%D1%87%D0%B5%D0%B1%D0%B0/6%20%D1%81%D0%B5%D0%BC%D0%B5%D1%81%D1%82%D1%80/%D0%A2%D1%80%D0%BE%D0%B4%20%28%D0%90%D0%BA%D1%83%D1%82%D0%B8%D0%BD%29/%D0%A1%D0%BA%D1%80%D0%B8%D0%BD%D1%8B%204/Dead.jpg)

### 4. Отказоустойчивость
Остановить consumer-main: 
````
docker-compose stop consumer-main
````
Нажать несколько раз любую кнопку отправки. 
#### RabbitMQ
![Main off.jpg](../../../%D0%A3%D1%87%D0%B5%D0%B1%D0%B0/6%20%D1%81%D0%B5%D0%BC%D0%B5%D1%81%D1%82%D1%80/%D0%A2%D1%80%D0%BE%D0%B4%20%28%D0%90%D0%BA%D1%83%D1%82%D0%B8%D0%BD%29/%D0%A1%D0%BA%D1%80%D0%B8%D0%BD%D1%8B%204/Main%20off.jpg)
Запустить consumer-main: 
````
docker-compose start consumer-main
````
#### RabbitMQ
![Main on.jpg](../../../%D0%A3%D1%87%D0%B5%D0%B1%D0%B0/6%20%D1%81%D0%B5%D0%BC%D0%B5%D1%81%D1%82%D1%80/%D0%A2%D1%80%D0%BE%D0%B4%20%28%D0%90%D0%BA%D1%83%D1%82%D0%B8%D0%BD%29/%D0%A1%D0%BA%D1%80%D0%B8%D0%BD%D1%8B%204/Main%20on.jpg)