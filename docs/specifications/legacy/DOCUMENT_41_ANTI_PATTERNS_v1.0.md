# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 41

# ANTI-PATTERNS, FORBIDDEN DECISIONS, ARCHITECTURAL SMELLS & PROJECT RED BOOK

Version: 1.0

Status: FINAL

Priority: ABSOLUTE

Classification: ENGINEERING CONSTITUTION

---

# PURPOSE

Этот документ описывает не то,

**что нужно делать**,

а

**что категорически запрещено делать.**

---

Документ обязателен для:

Builder Agent

Review Agent

QA Agent

Security Agent

Performance Agent

Documentation Agent

Human Developer

---

# FUNDAMENTAL PRINCIPLE

Любое решение должно быть максимально:

---

простым;

---

понятным;

---

предсказуемым;

---

масштабируемым;

---

обратимым.

---

# 1. ARCHITECTURAL ANTI-PATTERNS

## Запрещено

God Object

---

God Service

---

God Component

---

God Hook

---

God Store

---

Monolithic Context

---

Massive Controller

---

Massive Service

---

Massive Reducer

---

Massive SQL Query

---

# LIMITS

React Component

≤250 строк

---

Hook

≤150 строк

---

Go Service

≤300 строк

---

SQL Migration

одна ответственность

---

# 2. DATABASE ANTI-PATTERNS

Запрещено

---

SELECT *

---

N+1 Query

---

Nested Loop без необходимости

---

Дублирование данных

---

Отсутствие индексов

---

Динамический SQL

---

DROP COLUMN без миграции

---

Хранение JSON вместо нормальной модели

---

Логика приложения внутри SQL

---

# REVIEW QUESTIONS

Можно ли уменьшить JOIN?

---

Можно ли использовать Snapshot?

---

Можно ли использовать Materialized View?

---

Можно ли использовать Cache?

---

# 3. BACKEND ANTI-PATTERNS

Запрещено

---

бизнес-логика в Controller

---

SQL внутри Handler

---

Redis внутри Domain

---

HTTP внутри Domain

---

Global Variables

---

Singleton без необходимости

---

Reflection ради красоты

---

# 4. FRONTEND ANTI-PATTERNS

Запрещено

---

Props Drilling

---

State Explosion

---

Global Store для локального состояния

---

Nested Modals

---

Nested Scroll

---

20 useEffect подряд

---

UI с бизнес-логикой

---

API вызовы внутри компонентов

---

# 5. REDUX ANTI-PATTERNS

Запрещено

---

Store Everything

---

Derived State

---

Mutable State

---

Business Logic inside Slice

---

1000 строк в Slice

---

# 6. REACT ANTI-PATTERNS

Запрещено

---

Anonymous Functions Everywhere

---

Inline Objects Everywhere

---

Inline Styles Everywhere

---

Huge JSX

---

Conditional Hell

---

Magic Numbers

---

# 7. API ANTI-PATTERNS

Запрещено

---

REST + GraphQL одновременно

---

Breaking Changes

---

Versionless API

---

Huge Payload

---

GET изменяющий данные

---

POST без Idempotency

---

# 8. CACHE ANTI-PATTERNS

Запрещено

---

Infinite TTL

---

Cache Everything

---

Cache Personal Data

---

Cache XML

---

Cache JWT

---

# 9. AI ANTI-PATTERNS

ИИ запрещено

---

советовать покупать акции

---

советовать продавать акции

---

обещать доходность

---

самостоятельно менять сделки

---

отправлять декларации

---

генерировать финансовые данные

---

# 10. PRIVACY ANTI-PATTERNS

Запрещено

---

обязательный ИНН

---

обязательный паспорт

---

обязательный телефон

---

обязательный адрес

---

обязательное хранение деклараций

---

# 11. SECURITY ANTI-PATTERNS

Запрещено

---

JWT в LocalStorage

---

пароли в логах

---

секреты в Git

---

API Keys в коде

---

отключение TLS

---

# 12. TESTING ANTI-PATTERNS

Запрещено

---

100% Mock проект

---

Unit без реальных сценариев

---

E2E только Happy Path

---

игнорирование Edge Cases

---

# 13. PERFORMANCE ANTI-PATTERNS

Запрещено

---

пересчет портфеля на каждый запрос

---

пересчет XIRR на клиенте

---

полная загрузка истории

---

50 API запросов на Dashboard

---

# 14. COST ANTI-PATTERNS

Запрещено

---

LLM для простых вычислений

---

Redis для постоянного хранения

---

Python Worker для CRUD

---

постоянный polling

---

# 15. UX ANTI-PATTERNS

Запрещено

---

10 экранов регистрации

---

5 модальных окон подряд

---

таблицы на мобильном

---

скрытые кнопки

---

неочевидные действия

---

# 16. DESIGN ANTI-PATTERNS

Запрещено

---

10 цветов успеха

---

20 размеров текста

---

5 библиотек иконок

---

3 дизайн-системы одновременно

---

# 17. MOBILE ANTI-PATTERNS

Запрещено

---

WebView вместо Native

---

каждые 30 секунд обновлять API

---

большие Bundle

---

постоянные Background Jobs

---

# 18. DEVOPS ANTI-PATTERNS

Запрещено

---

Push напрямую в main

---

Deploy без тестов

---

Deploy без Review

---

Deploy без Backup

---

# 19. DOCUMENTATION ANTI-PATTERNS

Запрещено

---

код без документации

---

ADR без решения

---

OpenAPI не соответствует Backend

---

README устарел

---

# 20. PRODUCT ANTI-PATTERNS

Запрещено превращать OpenInvest в:

---

брокера

---

социальную сеть

---

новостной портал

---

чат

---

криптобиржу

---

терминал трейдера

---

# 21. MVP ANTI-PATTERNS

Запрещено

---

добавлять функции ради количества

---

делать Premium за счет урезания Free

---

писать AI "для галочки"

---

копировать интерфейс конкурентов

---

# 22. CODE REVIEW STOP LIST

Review Agent обязан немедленно отклонить Merge если обнаружено:

---

any без причины

---

TODO в Production

---

магические строки

---

магические числа

---

дублирование логики

---

циклические зависимости

---

мертвый код

---

# 23. PRINCIPAL ENGINEERING QUESTIONS

Перед Merge Builder Agent обязан спросить себя:

---

Можно удалить этот код?

---

Можно написать в два раза меньше?

---

Можно отказаться от библиотеки?

---

Можно отказаться от Worker?

---

Можно отказаться от Redis?

---

Можно отказаться от SQL?

---

Можно отказаться от API?

---

Можно отказаться от состояния?

---

Если ответ "Да",

код необходимо упростить.

---

# 24. OCCAM MODE

Всегда выбирать:

---

простое решение

---

вместо красивого.

---

понятное

---

вместо умного.

---

стабильное

---

вместо модного.

---

# 25. FINAL RED BOOK PRINCIPLE

> **Самая опасная ошибка проекта — не плохой код.**

> **Самая опасная ошибка — постепенное усложнение системы маленькими "временными" решениями.**

> **OpenInvest должен сопротивляться архитектурной энтропии.**

> **Каждый Builder Agent, Review Agent и инженер обязан оставить систему проще, понятнее, быстрее и дешевле, чем она была до его изменений.**

> **Если невозможно доказать необходимость новой сущности, нового сервиса, новой библиотеки или нового уровня абстракции — они не должны появляться в проекте.**
