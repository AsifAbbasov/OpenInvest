# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 21

# SECURITY, PRIVACY BY DESIGN, ZERO TRUST, COMPLIANCE & USER TRUST

Version: 1.0

Status: Approved

Priority: CRITICAL

Classification: MUST IMPLEMENT

---

# PURPOSE

Данный документ определяет всю архитектуру безопасности OpenInvest.

Это один из самых важных документов проекта.

OpenInvest должен выделяться не только аналитикой, но и доверием пользователей.

---

# PRODUCT PHILOSOPHY

Большинство инвестиционных сервисов говорят:

> "Дайте нам все свои данные, и мы их защитим."

OpenInvest говорит иначе:

> **"Лучше вообще не хранить ваши данные, если без этого можно обойтись."**

---

# TRUST FIRST

Главная ценность продукта:

```
Trust

>

Features
```

---

# PRIVACY BY DESIGN

Любой новый функционал должен отвечать на вопрос:

```
Можно ли сделать это

НЕ сохраняя

персональные данные?
```

Если ответ "Да",

персональные данные запрещено хранить.

---

# ZERO TRUST ARCHITECTURE

Никто внутри системы

не считается доверенным.

Даже внутренний сервис обязан:

* авторизоваться;
* логироваться;
* иметь ограниченные права.

---

# PRINCIPLES

Minimal Data Collection

Least Privilege

Need To Know

Zero Trust

Encryption Everywhere

Audit Everything

Explicit Consent

User Ownership

Delete By Default

---

# USER DATA LEVELS

## LEVEL 1

Public

```
Ticker

Portfolio Name

Theme

Language
```

---

## LEVEL 2

Private

```
Email

Timezone

Notification Settings
```

---

## LEVEL 3

Sensitive

```
Passport

INN

Address

Phone
```

---

## LEVEL 4

Critical

```
Password

JWT

Refresh Token

Encryption Keys
```

---

# DATA SEPARATION

Инвестиционные данные

никогда

не должны храниться вместе

с персональными данными.

---

## Example

```
Schema:

invest.*

user_private.*

audit.*
```

---

# PASSWORDS

Хранение:

Argon2id

или

bcrypt

---

Запрещено:

SHA1

SHA256

MD5

---

# AUTHENTICATION

Поддерживаются:

Email

Google (future)

Apple (future)

Telegram (future)

Passkeys (future)

---

# SESSION MODEL

Access Token

15 минут

---

Refresh Token

30 дней

---

# DEVICE BINDING

Каждая сессия привязывается к:

```
Browser

OS

Device ID

Approximate Region
```

---

# LOGIN NOTIFICATION

При входе с нового устройства:

Email

*

Push

---

# TWO FACTOR AUTHENTICATION

Опционально.

---

Поддержка:

TOTP

Authenticator

Passkeys

---

SMS не является обязательным способом.

---

# API SECURITY

Все API работают только через HTTPS.

---

Минимальная версия TLS:

```
TLS 1.3
```

---

# API KEYS

Никогда

не попадают во Frontend.

---

Все ключи:

Backend Only

---

# SECRET STORAGE

```
.env

Vault

Secret Manager
```

---

Запрещено:

```
git

frontend

javascript bundle
```

---

# DATABASE ENCRYPTION

Disk Encryption

*

Field Encryption

---

AES-256

---

# PERSONAL DATA

Поля:

ИНН

Паспорт

Адрес

шифруются отдельно.

---

# PRIVATE MODE

Пользователь может выбрать:

```
☑ Не сохранять паспорт

☑ Не сохранять ИНН

☑ Не сохранять адрес

☑ Не сохранять телефон
```

---

# TEMPORARY MODE

Пользователь заполняет:

ИНН

Паспорт

Адрес

---

OpenInvest:

генерирует декларацию

↓

отправляет XML/PDF

↓

полностью уничтожает данные

из памяти процесса.

---

# NEVER STORE

По умолчанию запрещено хранить:

паспорт;

ИНН;

адрес;

место регистрации.

---

# USER CONSENT

Каждый чекбокс:

отдельный.

---

Запрещено:

```
☑ Я согласен на всё
```

---

# RIGHT TO DELETE

В профиле всегда есть кнопка:

```
Удалить профиль полностью
```

---

После подтверждения:

удаляются:

персональные данные;

сессии;

экспорты;

уведомления.

---

Инвестиционные записи:

анонимизируются.

---

# EXPORT MY DATA

Пользователь может скачать:

```
JSON

CSV

XML

PDF
```

---

# AUDIT

Каждое действие записывается.

---

Но Audit

не содержит:

пароли;

паспорт;

ИНН.

---

# EMAIL SECURITY

Все письма подписываются.

---

SPF

DKIM

DMARC

---

# FILE SECURITY

PDF

XML

ZIP

создаются

только в RAM.

---

Никогда

не сохраняются

во временную папку сервера.

---

# DOWNLOAD LINKS

Подписанные ссылки.

---

TTL

15 минут.

---

После:

автоматическое удаление.

---

# RATE LIMIT

Authorization

5 попыток

за 15 минут.

---

Password Reset

3 попытки

в час.

---

Tax Export

10 запросов

в день.

---

# DDOS PROTECTION

Cloudflare

или аналог.

---

# BOT PROTECTION

Invisible CAPTCHA

Behavior Analysis

Rate Limiting

---

# SQL INJECTION

Использовать только:

Prepared Statements

ORM

Query Builder

---

Запрещено:

String Concatenation.

---

# XSS

Все данные экранируются.

---

# CSP

Content Security Policy

обязателен.

---

# CSRF

Используются:

SameSite Cookies

CSRF Tokens

---

# SECURITY HEADERS

Обязательны:

```
HSTS

X-Frame-Options

Referrer-Policy

Permissions-Policy

X-Content-Type-Options
```

---

# LOGGING POLICY

Никогда не логировать:

пароли;

JWT;

Refresh;

ИНН;

паспорт;

адрес.

---

# BACKUPS

Encrypted

---

Automatic

---

Daily

---

Retention

30 дней

---

# INCIDENT RESPONSE

При обнаружении компрометации:

1.

отключение API;

2.

отзыв всех Refresh Token;

3.

уведомление пользователей;

4.

создание Audit Report.

---

# THIRD PARTY POLICY

Перед интеграцией любого сервиса проверить:

официальность;

лицензию;

GDPR;

локальное законодательство;

политику хранения данных.

---

# AI SECURITY

ИИ запрещено:

генерировать налоговые данные без проверки;

давать инвестиционные рекомендации;

изменять пользовательские данные самостоятельно;

подписывать документы.

---

# HUMAN IN THE LOOP

Любой налоговый документ:

XML

PDF

Email

должен быть подтвержден пользователем.

---

# USER TRUST PAGE

В приложении создается отдельный раздел:

## Почему OpenInvest безопасен

Показывается:

```
Какие данные мы собираем

Какие НЕ собираем

Как они шифруются

Как удалить профиль

Как экспортировать данные

Как формируется декларация

Как работают аудит-логи
```

---

# COMPETITIVE ADVANTAGE

Большинство брокеров требуют:

паспорт;

ИНН;

адрес;

телефон;

десятки обязательных полей.

---

OpenInvest позволяет работать:

без паспорта;

без ИНН;

без адреса;

без хранения персональных данных.

---

Пользователь сам выбирает:

```
Private Mode

или

Convenience Mode
```

---

# TRUST SCORE

Каждый новый функционал Builder Agent обязан проверить:

1.

Увеличивает ли он доверие пользователя?

2.

Можно ли реализовать его без новых персональных данных?

3.

Можно ли сделать его локально?

4.

Можно ли отказаться от хранения данных?

5.

Можно ли удалить все данные одним кликом?

6.

Будет ли комфортно объяснить эту функцию аудитору, юристу и обычному пользователю?

Если хотя бы один ответ отрицательный — функционал отправляется на архитектурный пересмотр.

---

# FINAL PRINCIPLE

> **OpenInvest продает не инвестиционные идеи. OpenInvest продает спокойствие, прозрачность и контроль над собственными финансами.**

> **Пользователь всегда владеет своими данными, понимает каждую цифру и может уйти из системы без потери контроля над своей информацией.**
