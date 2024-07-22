# webcrawler
Сервис, который индексирует web-страницы в Elasticsearch, а также возвращает ссылки на страницы по введенному запросу

№№ Архитектура
![image](https://github.com/user-attachments/assets/654018cc-4c79-416c-85c7-a2b7f2f2868a)
Для полнотекстового поиска используется Elasticsearch, в котором хранится ключевые слова содержания web-страниц. Для хранения полной информации о странице используется Postgres.

## Установка и конфигурация
- Склонировать репозиторий:
  ```
  git@github.com:mx4alex/webcrawler.git
  ```
- Настроить конфигурацию в файле `config.yaml`
- Запустить *docker compose*
  ```
  docker compose up --build
  ```

## Использование

### Сервис поддерживает следующий эндпоинт:
- `GET /webcrawler/search/{input}` выполняет поиск web-страниц по введенному запросу

* Формат запроса:
```
curl -X GET http://localhost:8080/webcrawler/search/Технология
```
* Формат ответа:
```json
[
    {
        "url": "https://indicator.ru/engineering-science/startap-iz-sfu-razrabotal-tekhnologiyu-obogasheniya-grafita-sposobnuyu-izmenit-rynok-materialov-05-07-2024.htm",
        "text": "Гильманшина. Основной инновацией проекта является технология термического обогащения графита, что позволит"
    },
    {
        "url": "https://indicator.ru/medicine/rasshifrovannyi-genom-shtamma-listerii-pomozhet-sdelat-vakciny-bezopasnee-05-07-2024.htm",
        "text": "до начала XXI века оставалась технология получения вакцин путем аттенуации вакцинных"
    },
    {
        "url": "https://indicator.ru/mathematics/cfa-dadut-novyi-tolchok-v-razvitii-klassicheskikh-finansovykh-instrumentov.htm",
        "text": "консенсуса и их нельзя изменить. Технология распределенного реестра — и блокчейна"
    },
    {
        "url": "https://indicator.ru/medicine/pmef-2024-ne-vmesto-a-vmeste-potencial-primeneniya-ii-v-rossiiskom-zdravookhranenii.htm",
        "text": "инструмента; риск потери времени, когда технология уже готова, но не применяется"
    },
    {
        "url": "https://indicator.ru/medicine/novyi-podkhod-uprostit-monitoring-krovotoka-pri-operaciyakh-na-golovnom-mozge-19-03-2024.htm",
        "text": "кровотечениях или появлении бликов, новая технология демонстрирует значительный потенциал для улучшения"
    }
]
```

![image](https://github.com/user-attachments/assets/2f1da4df-8aeb-46a5-8fed-efc6926aac94)
