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
<img width="711" height="33" alt="Retry" src="https://github.com/user-attachments/assets/e9b2aeb8-ad5b-4be9-88d2-7e90d77906f6" />

consumer-main получает заказ, фиксирует ошибку сети, отправляет в queue.retry. В RabbitMQ очередь queue.retry показывает Ready: 1. Сообщение ждёт повторной обработки. 
Запустить consumer-retry: 
```
docker-compose start consumer-retry 
```
#### RabbitMQ
<img width="1000" height="40" alt="Retry after" src="https://github.com/user-attachments/assets/66cc55c2-f254-4c0c-85d7-8158ad575388" />


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
<img width="712" height="43" alt="Dead" src="https://github.com/user-attachments/assets/d5354320-c734-4e3a-85f0-9de20696e39e" />


### 4. Отказоустойчивость
Остановить consumer-main: 
````
docker-compose stop consumer-main
````
Нажать несколько раз любую кнопку отправки. 
#### RabbitMQ
<img width="712" height="35" alt="Main off" src="https://github.com/user-attachments/assets/9142f36a-37cb-43b5-9951-97fa4a9a477a" />

Запустить consumer-main: 
````
docker-compose start consumer-main
````
#### RabbitMQ
<img width="720" height="112" alt="Main on" src="https://github.com/user-attachments/assets/aec572ae-fec7-407d-a272-dd51d6caa0df" />
