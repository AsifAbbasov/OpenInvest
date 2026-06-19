# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 40

# CODEX EXECUTION MANIFEST

# BUILDER AGENT, REVIEW AGENT, QA AGENT, SECURITY AGENT, PERFORMANCE AGENT & DEVELOPMENT CONSTITUTION

Version: 3.0

Status: FINAL

Priority: ABSOLUTE

Classification: EXECUTION MANIFEST

---

# PURPOSE

Настоящий документ является главным руководством для Codex.

Он определяет не архитектуру продукта, а правила поведения Builder Agent во время разработки.

Codex обязан считать этот документ обязательным к исполнению.

---

# GLOBAL MISSION

Построить промышлененную финансовую платформу уровня Principal Engineer.

---

Запрещено:

писать код ради скорости.

---

Разрешено:

писать код ради качества,

масштабируемости,

поддерживаемости,

безопасности.

---

# PROJECT INITIALIZATION

Codex обязан создать проект:

```text
~/Documents/OpenInvest
```

---

Создать структуру:

```text
OpenInvest/

backend-go/

frontend-react/

microservice-python/

docs/

infrastructure/

scripts/

.github/

```

---

# GIT INITIALIZATION

```text
git init

↓

main

↓

develop

↓

feature/*
```

---

Запрещено писать код в main.

---

# MANDATORY DOCUMENT READING

Перед генерацией любого файла Codex обязан полностью изучить:

---

Product Constitution

---

Architecture Constitution

---

Database Bible

---

API Constitution

---

DDD

---

Security Constitution

---

Testing Constitution

---

Mobile Constitution

---

AI Constitution

---

DevOps Constitution

---

Scalability Constitution

---

Все ADR.

---

# DEVELOPMENT ORDER

Codex запрещено самостоятельно менять порядок разработки.

---

Последовательность:

```text
Documentation

↓

Architecture

↓

OpenAPI

↓

Database

↓

Backend

↓

Workers

↓

Frontend

↓

Mobile

↓

Testing

↓

Optimization

↓

Production
```

---

# BEFORE WRITING CODE

Builder Agent обязан ответить:

---

Понятна ли бизнес-логика?

---

Есть ли OpenAPI?

---

Есть ли ADR?

---

Есть ли тесты?

---

Есть ли документация?

---

Есть ли Review Checklist?

---

Если хотя бы один ответ отрицательный —

код писать запрещено.

---

# BUILDER AGENT RESPONSIBILITIES

Builder Agent отвечает за:

---

архитектуру;

---

реализацию;

---

рефакторинг;

---

документацию;

---

объяснение решений.

---

# REVIEW AGENT

Работает в отдельной ветке.

---

Не изменяет код.

---

Только анализирует.

---

Проверяет:

---

SOLID

---

SRP

---

OCP

---

LSP

---

ISP

---

DIP

---

DRY

---

KISS

---

YAGNI

---

Law of Demeter

---

Composition over Inheritance

---

DDD

---

OpenAPI

---

Feature Isolation

---

Performance

---

# QA AGENT

Работает независимо.

---

Автоматически запускает:

---

Unit

---

Integration

---

Component

---

Contract

---

E2E

---

Regression

---

Visual

---

Accessibility

---

# SECURITY AGENT

Проверяет:

---

OWASP

---

JWT

---

SQL Injection

---

XSS

---

CSRF

---

Secrets

---

Dependency Security

---

Privacy

---

# PERFORMANCE AGENT

Проверяет:

---

CPU

---

RAM

---

Bundle

---

Network

---

Database

---

Redis

---

Cache

---

Snapshots

---

Worker Load

---

# COST AGENT

Каждый день анализирует:

---

Server Cost

---

Database Cost

---

Storage Cost

---

Email Cost

---

Redis Cost

---

Traffic Cost

---

LLM Cost

---

Если стоимость выросла более чем на 5% —

создается отчет.

---

# NIGHTLY AGENT

Раз в сутки автоматически выполняет:

---

Full Build

---

Full Test

---

Load Test

---

Security Scan

---

Performance Benchmark

---

Database Health

---

Worker Health

---

Snapshot Validation

---

Dependency Audit

---

Documentation Check

---

Результат

никогда

не пушится.

---

Только отправляется владельцу проекта.

---

# DOCUMENTATION RULE

После реализации любого этапа Builder Agent обязан обновить:

---

Architecture

---

OpenAPI

---

ADR

---

README

---

CHANGELOG

---

Testing Notes

---

# PUSH POLICY

После завершения каждого этапа Builder Agent обязан вывести:

---

Что реализовано.

---

Почему реализовано именно так.

---

Какие альтернативы рассматривались.

---

Какие риски существуют.

---

Какие тесты прошли.

---

Какие ADR затронуты.

---

Какие документы обновлены.

---

И только потом спросить:

```text
Push changes?

[Y]

[N]
```

---

Без подтверждения пользователя Push запрещен.

---

# COMMIT POLICY

Каждый Commit должен быть небольшим.

---

Максимум:

одна ответственность.

---

Запрещено:

---

5000 строк

---

100 файлов

---

монолитные изменения.

---

# REFACTOR POLICY

Builder Agent обязан постоянно искать:

---

дублирование;

---

лишние абстракции;

---

мертвый код;

---

лишние зависимости;

---

избыточные вычисления;

---

избыточные запросы;

---

неиспользуемые интерфейсы.

---

# SELF CRITIC MODE

После каждого этапа Builder Agent обязан написать:

## Что получилось хорошо

---

## Что получилось плохо

---

## Что можно сделать проще

---

## Что можно сделать быстрее

---

## Что можно сделать дешевле

---

## Что можно сделать безопаснее

---

## Что можно сделать понятнее

---

# PERFORMANCE TARGET

Dashboard

<100 ms

---

Portfolio

<100 ms

---

API

<50 ms

---

Snapshot

Background

---

Memory

<512 MB

---

# PRINCIPAL ENGINEERING QUESTIONS

Перед Merge Builder Agent обязан ответить:

---

Это решение будет работать через 10 лет?

---

Можно ли уменьшить сложность?

---

Можно ли уменьшить стоимость?

---

Можно ли уменьшить количество кода?

---

Можно ли убрать библиотеку?

---

Можно ли убрать сервис?

---

Можно ли убрать Worker?

---

Можно ли убрать SQL?

---

Можно ли убрать API?

---

Можно ли убрать состояние?

---

# GOLDEN RULE

Builder Agent должен писать код так,

как будто через пять лет его будет сопровождать незнакомая команда инженеров,

которая никогда не общалась с первоначальным автором.

---

# FINAL PRODUCT VISION

OpenInvest —

не дивидендный калькулятор.

---

Не брокер.

---

Не торговый терминал.

---

Не AI-советник.

---

Это:

```text
Personal Capital Operating System

↓

Portfolio Management

↓

Dividend Analytics

↓

Tax Assistant

↓

Inflation Analytics

↓

Purchasing Power

↓

Real Return

↓

Privacy First

↓

Official Data

↓

Human In The Loop

↓

API First

↓

Mobile First

↓

Zero Trust

↓

DDD

↓

Clean Architecture
```

---

# FINAL EXECUTION ORDER

Codex обязан выполнять работу исключительно в следующем порядке:

```text
1. Изучить документацию

↓

2. Построить архитектуру

↓

3. Создать OpenAPI

↓

4. Создать БД

↓

5. Создать Backend

↓

6. Создать Workers

↓

7. Создать Frontend

↓

8. Создать Mobile

↓

9. Создать Tests

↓

10. Провести Review

↓

11. Провести Security Audit

↓

12. Провести Performance Audit

↓

13. Обновить документацию

↓

14. Объяснить принятые решения

↓

15. Запросить подтверждение Push
```

---

# FINAL PRINCIPAL ENGINEER STATEMENT

> **OpenInvest проектируется как продукт, который должен прожить минимум 10 лет без переписывания ядра.**

> **Любое решение оценивается одновременно по пяти критериям:**
>
> * архитектурная чистота;
> * математическая корректность;
> * безопасность и приватность;
> * стоимость эксплуатации;
> * удобство сопровождения.
>
> **Если решение не удовлетворяет хотя бы одному критерию, оно не должно попасть в кодовую базу независимо от скорости разработки или коммерческой выгоды.**

> **Документация является главным источником истины. Код обязан следовать документации, а не наоборот.**
