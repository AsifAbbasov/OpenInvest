# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 20


Version: 1.0

Status: Approved

Priority: Critical

---

# PURPOSE

Этот документ определяет обязательную систему контроля качества проекта OpenInvest.

Цель:

**не допустить попадания плохого кода в production.**

Каждый коммит обязан пройти автоматические и ручные проверки.

---

# DEVELOPMENT PHILOSOPHY


никогда

не является последней инстанцией.

После каждого изменения код должен пройти:

```

↓


↓

QA Agent

↓

Integration Tests

↓

Performance Tests

↓

Security Tests

↓


↓

Git Push
```

---

# PROJECT BRANCH STRATEGY

Использовать только Git Flow.

```
main

↓

develop

↓

feature/*

↓

review/*

↓

hotfix/*
```

---

# REVIEW CHECKLIST


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

Occam Razor

---

Separation of Concerns

---

Composition over Inheritance

---

Clean Architecture

---

Feature Isolation

---

Type Safety

---

Memory Usage

---

API Consistency

---

Security

---

Accessibility

---

Performance

---

# REVIEW REPORT

После проверки Agent обязан выдать:

```
Architecture Score

Code Quality Score

Performance Score

Security Score

Maintainability Score

Overall Score
```

---

# AUTOMATIC REJECTION

PR автоматически отклоняется если:

Architecture Score

<90

или

Security Score

<95

или

Tests Coverage

<90%

---

# QA AGENT

QA Agent —

это независимый тестировщик.

Он ничего не знает о реализации.

Он тестирует продукт глазами пользователя.

---

# QA RESPONSIBILITIES

Smoke Tests

Regression Tests

Integration Tests

UI Tests

UX Tests

Accessibility Tests

Performance Tests

API Tests

Load Tests

Security Tests

---

# SMOKE TESTS

После каждого merge:

```
Application Starts

↓

Authorization Works

↓

Portfolio Opens

↓

Catalog Opens

↓

Tax Module Opens

↓

Charts Render
```

---

# REGRESSION TESTS

Проверяются:

старые функции;

новые функции;

интеграция между ними.

---

# UNIT TESTS

Target Coverage

95%

Minimum

90%

---

# INTEGRATION TESTS

Обязательно тестировать:

React

↓

API

↓

Go

↓

PostgreSQL

↓

Redis

↓

Python Worker

---

# E2E TESTS

Playwright

---

Сценарии:

Регистрация

Авторизация

Создание портфеля

Добавление сделки

Редактирование

Удаление

Экспорт XML

Экспорт PDF

Email

Logout

---

# VISUAL TESTS

Автоматически проверяются:

Desktop

Tablet

Mobile

---

# SUPPORTED RESOLUTIONS

```
390

768

1024

1280

1440

1920
```

---

# SUPPORTED BROWSERS

Chrome

Safari

Firefox

Edge

---

# MOBILE TESTS

iPhone

Android

---

Landscape

Portrait

---

# PERFORMANCE TESTS

Lighthouse

Target

95+

---

Metrics

FCP

LCP

CLS

TTI

INP

---

# MEMORY TESTS

Frontend

не должен:

создавать утечки памяти;

создавать бесконечные render loops.

---

# API TESTS

Проверяются:

Latency

Errors

Timeouts

Retries

Rate Limits

Compression

Caching

---

# LOAD TESTS

k6

или

Locust

---

Минимальный сценарий:

10000 одновременных пользователей.

---

# SECURITY TESTS

JWT

XSS

CSRF

SQL Injection

Rate Limit

Broken Auth

Session Hijacking

---

# PRIVACY TESTS

Проверяется:

не сохраняются ли

паспортные данные,

если пользователь выбрал

Private Mode.

---

# TAX TESTS

Проверяется:

XML

PDF

ZIP

Email

---

# FINANCIAL TESTS

Самые критичные.

---

Проверяется:

Weighted Average Cost

XIRR

NKD

Coupon

Inflation

Real Value

Dividend Yield

---

# GOLDEN DATASETS

Создать специальные тестовые портфели.

---

Portfolio A

Только акции

---

Portfolio B

Акции + облигации

---

Portfolio C

Докупки

---

Portfolio D

Продажи

---

Portfolio E

Валютные дивиденды

---

Portfolio F

Налоговая декларация

---

#  TESTS

 Assistant обязан:

не давать инвестиционных рекомендаций;

не обещать прибыль;

не искажать расчеты.

---

# NIGHTLY AGENT

Каждые сутки запускается

Nightly QA Agent.

---

Он:

клонирует develop;

запускает все тесты;

проверяет производительность;

строит отчет.

---

# NIGHTLY REPORT

```
Architecture

Tests

Coverage

Performance

Security

Accessibility

Regression

Memory

API

Overall Status
```

---

# WEEKLY REPORT

Формируется автоматически.

---

Содержит:

Количество ошибок

Количество исправлений

Среднее время ответа

Использование RAM

Использование CPU

Cache Hit Ratio

Slow Queries

---

# CI/CD

Pipeline

```
Commit

↓

Lint

↓

Build

↓

Unit Tests

↓

Integration Tests

↓

E2E Tests

↓


↓

QA Agent

↓

Security Scan

↓


↓

Merge

↓

Deploy
```

---

# DEPLOYMENT RULES


никогда

не деплоит самостоятельно.

---

Production Deploy

возможен только после:


*

QA Agent

*


---

# FINAL QUALITY RULE

Любая новая функция должна пройти проверку по одному вопросу:

> **"Если завтра проектом будут пользоваться 100 000 инвесторов, останется ли эта архитектура такой же быстрой, безопасной, дешевой в эксплуатации и понятной для сопровождения?"**

Если ответ отрицательный — функция отправляется на переработку.
