# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 36

# AI ARCHITECTURE BIBLE

# AI ASSISTANT, RAG, AGENT SYSTEM, HUMAN-IN-THE-LOOP, PROMPT SECURITY & AUTONOMOUS ANALYTICS

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: AI CONSTITUTION

---

# PURPOSE

Настоящий документ определяет принципы работы искусственного интеллекта OpenInvest.

ИИ является помощником пользователя.

ИИ никогда не становится управляющим капиталом.

OpenInvest категорически запрещает автономное принятие инвестиционных решений искусственным интеллектом.

---

# AI PHILOSOPHY

ИИ должен:

объяснять;

анализировать;

визуализировать;

структурировать;

напоминать;

обучать.

---

ИИ запрещено:

рекомендовать покупать;

рекомендовать продавать;

обещать прибыль;

гарантировать доходность;

заменять финансового консультанта.

---

# AI POSITION

```text
User

↓

AI Assistant

↓

OpenAPI

↓

Go Backend

↓

Canonical Database
```

---

ИИ никогда не получает прямой доступ к PostgreSQL.

---

# AI TYPES

## Assistant AI

Работает с пользователем.

---

## Analytics AI

Строит аналитические выводы.

---

## Tax AI

Помогает понять декларацию.

---

## Portfolio AI

Объясняет изменения капитала.

---

## Education AI

Объясняет финансовые термины.

---

# HUMAN IN THE LOOP

Любое критическое действие требует подтверждения пользователя.

---

ИИ может:

подготовить XML;

подготовить PDF;

подготовить письмо;

подготовить сценарий.

---

ИИ не может:

отправить XML;

удалить профиль;

изменить сделки;

изменить налоги;

создать портфель без пользователя.

---

# RAG ARCHITECTURE

ИИ отвечает только на основе проверенных источников.

```text
Official Documents

↓

Knowledge Index

↓

Embeddings

↓

Retriever

↓

Context Builder

↓

LLM

↓

Verified Answer
```

---

# KNOWLEDGE SOURCES

Используются только:

---

документация OpenInvest;

---

официальные документы ЦБ;

---

официальные документы ФНС;

---

официальные документы MOEX;

---

официальная документация продукта.

---

# AI MEMORY

ИИ не хранит:

пароли;

ИНН;

паспорт;

JWT;

Refresh Token;

SMTP;

секреты.

---

ИИ хранит только контекст текущего диалога.

---

# PROMPT SECURITY

Любой Prompt проходит:

```text
Input

↓

Validation

↓

Sanitization

↓

Policy Check

↓

Execution
```

---

# PROMPT INJECTION DEFENSE

Игнорируются попытки:

---

получить системные инструкции;

---

получить ключи;

---

обойти права доступа;

---

изменить бизнес-логику.

---

# TOOL PERMISSIONS

ИИ имеет доступ только к:

---

Portfolio Read

---

Analytics Read

---

Tax Preview

---

Dividend Calendar

---

Notifications Preview

---

ИИ не имеет доступа к:

Delete

Update

Execute Payments

Send Declaration

---

# AI CONTEXT WINDOW

Приоритет контекста:

```text
Current User Data

↓

Portfolio

↓

Snapshots

↓

Analytics

↓

Knowledge Base

↓

General Information
```

---

# AI EXPLANATION MODE

Любой вывод сопровождается:

---

формулой;

---

источником;

---

датой;

---

уровнем уверенности.

---

Пример:

```text
XIRR

12.48%

Источник:

Transactions

Snapshots

Дата расчета:

2027-03-12

Уверенность:

100%
```

---

# AI PORTFOLIO REVIEW

ИИ может автоматически сформировать:

---

сводку месяца;

---

динамику капитала;

---

дивиденды;

---

налоги;

---

инфляционную корректировку;

---

реальную доходность.

---

# AI GOAL PLANNER

Пользователь может указать:

---

накопить на квартиру;

---

создать пенсионный капитал;

---

получать дивиденды;

---

накопить на образование.

---

ИИ показывает только математические сценарии.

---

Без инвестиционных рекомендаций.

---

# AI EXPLANATION STYLE

Запрещается использовать сложную терминологию без пояснения.

---

Вместо:

```text
Sortino Ratio = 2.11
```

использовать:

```text
Ваш портфель показывает высокую доходность при относительно низком риске отрицательных колебаний.
```

---

# AI REPORTS

Генерируются:

---

Weekly

---

Monthly

---

Quarterly

---

Yearly

---

# AI ANALYTICS

Могут включать:

---

XIRR

---

TWR

---

Sharpe

---

Sortino

---

Max Drawdown

---

Dividend CAGR

---

Yield on Cost

---

Inflation Adjusted Return

---

Purchasing Power Index

---

# PERSONAL CAPITAL REVIEW

ИИ генерирует понятный отчет:

```text
За последний год

капитал вырос на 14%

после инфляции:

8%

после налогов:

6%

дивиденды составили:

43 200 ₽

покупательная способность выросла на:

5%
```

---

# AI SCENARIOS

Допускаются:

---

Что если докупить?

---

Что если не продавать?

---

Что если инфляция вырастет?

---

Что если дивиденды сократятся?

---

# AI ETHICS

ИИ обязан:

---

указывать ограничения;

---

не скрывать неопределенность;

---

не создавать иллюзию точности;

---

не придумывать данные.

---

# AGENT SYSTEM

Builder Agent

↓

Review Agent

↓

QA Agent

↓

Security Agent

↓

Performance Agent

↓

Documentation Agent

↓

AI Assistant

---

Каждый агент имеет собственную ответственность.

---

# AGENT COMMUNICATION

Агенты взаимодействуют только через:

---

Events

---

OpenAPI

---

Contracts

---

Documentation

---

Запрещено использовать скрытые зависимости.

---

# AI OBSERVABILITY

Каждый AI ответ логируется:

---

Model Version

---

Prompt Version

---

Knowledge Version

---

Response Time

---

Confidence

---

Без хранения персональных данных.

---

# COST OPTIMIZATION

Приоритет:

---

локальная аналитика;

---

кэш;

---

RAG;

---

только затем LLM.

---

# FUTURE READY

Архитектура должна позволять добавить:

---

локальную LLM;

---

корпоративного AI;

---

Family AI Advisor;

---

Voice Assistant;

---

Apple Intelligence;

---

Android Gemini Integration.

---

# AI REVIEW CHECKLIST

Перед публикацией новой AI-функции необходимо ответить:

---

Использует ли она официальные данные?

---

Может ли пользователь проверить вывод?

---

Есть ли Human-in-the-Loop?

---

Может ли ИИ случайно принять финансовое решение вместо пользователя?

---

Можно ли объяснить результат простым языком?

---

# FINAL AI PRINCIPLE

> **OpenInvest AI — это интеллектуальный финансовый интерпретатор, а не инвестиционный советник.**

> **ИИ существует для повышения понимания, прозрачности и удобства работы пользователя со своим капиталом.**

> **Любая функция, которая уменьшает контроль пользователя над собственными финансами, считается нарушением философии продукта и должна быть отклонена на этапе архитектурного ревью.**
