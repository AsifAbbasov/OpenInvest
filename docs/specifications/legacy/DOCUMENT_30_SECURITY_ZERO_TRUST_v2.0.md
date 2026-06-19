# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 30

# SECURITY, ZERO TRUST, PRIVACY BY DESIGN, DATA PROTECTION, CRYPTOGRAPHY & USER TRUST CONSTITUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: SECURITY FOUNDATION

---

# PURPOSE

Настоящий документ определяет фундаментальные принципы безопасности OpenInvest.

Безопасность рассматривается не как отдельная функция, а как неотъемлемая часть архитектуры.

Любое решение, которое улучшает функциональность, но ухудшает безопасность или приватность пользователя, автоматически считается неправильным.

---

# 1. MAIN PRINCIPLE

OpenInvest не должен требовать доверия.

OpenInvest должен быть спроектирован так, чтобы пользователь мог пользоваться продуктом даже не доверяя ему полностью.

---

# 2. ZERO TRUST ARCHITECTURE

Никто не считается доверенным.

Не пользователь.

Не сервер.

Не внутренний сервис.

Не AI.

Не администратор.

---

Каждый запрос проходит повторную проверку.

---

# 3. PRIVACY BY DESIGN

Любая новая функция проходит пять вопросов:

---

Можно ли вообще не собирать эти данные?

---

Можно ли их обезличить?

---

Можно ли удалить их сразу после использования?

---

Можно ли предоставить пользователю выбор?

---

Можно ли сделать функцию без хранения данных?

---

Если ответ "Да", данные не сохраняются.

---

# 4. MINIMAL DATA COLLECTION

OpenInvest собирает минимально возможный объем информации.

---

Обязательные данные:

Email

Password Hash

Settings

Portfolio

Transactions

---

Необязательные:

Имя

ИНН

Паспорт

Адрес

Телефон

Дата рождения

---

По умолчанию они отсутствуют.

---

# 5. PRIVATE MODE

Пользователь может работать полностью анонимно.

---

Доступно:

Портфель

Дивиденды

XIRR

Инфляция

Real Return

Календарь

Графики

---

Недоступно:

Автоматическое заполнение декларации личными данными.

---

# 6. TAX MODE

Существует три режима.

---

## Anonymous

Пользователь получает пустую декларацию.

Заполняет вручную.

---

## Assisted

Пользователь вводит данные один раз.

После генерации документы автоматически очищаются.

---

## Saved

Пользователь сознательно разрешает хранение данных.

---

# 7. DELETE POLICY

Любые персональные данные удаляются:

---

по запросу пользователя;

---

при закрытии аккаунта;

---

после истечения срока хранения.

---

Удаление необратимо.

---

# 8. EXPORT POLICY

Пользователь всегда может скачать:

---

весь профиль;

---

портфель;

---

историю;

---

налоговые документы;

---

настройки.

---

Форматы:

JSON

CSV

PDF

ZIP

---

# 9. DATA SEPARATION

Инвестиционные данные

никогда

не смешиваются

с персональными.

---

```id="opr8ak"
Identity Database

↓

UUID

↓

Investment Database
```

---

Даже при компрометации одной базы невозможно восстановить вторую.

---

# 10. ENCRYPTION

---

TLS 1.3

---

AES-256

---

Argon2id

---

JWT Rotation

---

Refresh Rotation

---

Encrypted Backups

---

# 11. PASSWORD POLICY

Пароли никогда не хранятся.

---

Используется:

Argon2id

Salt

Memory Cost

Time Cost

Parallelism

---

# 12. SECRET MANAGEMENT

Запрещено хранить:

API Keys

SMTP

JWT

DB Passwords

в Git.

---

Используются:

Environment Variables

или

Secrets Manager.

---

# 13. SESSION SECURITY

Access Token

15 минут.

---

Refresh Token

30 дней.

---

Rotation обязательна.

---

# 14. DEVICE MANAGEMENT

Пользователь видит:

---

активные устройства;

---

последний вход;

---

браузер;

---

операционную систему.

---

Может завершить любую сессию.

---

# 15. AUDIT LOG

Все критические действия фиксируются.

---

Вход

---

Удаление данных

---

Создание XML

---

Изменение Email

---

Смена пароля

---

Экспорт профиля

---

# 16. AUDIT IMMUTABILITY

Audit Log запрещено изменять.

---

Допускается только:

Append Only.

---

# 17. HUMAN IN THE LOOP

ИИ не имеет права:

---

самостоятельно отправлять декларацию;

---

изменять портфель;

---

создавать сделки;

---

изменять персональные данные.

---

Финальное действие всегда подтверждает пользователь.

---

# 18. AI SANDBOX

ИИ работает только с копией данных.

---

ИИ никогда не получает:

пароль;

Refresh Token;

JWT;

SMTP;

ключи;

секреты.

---

# 19. API SECURITY

Все API проходят:

Authentication

↓

Authorization

↓

Validation

↓

Rate Limit

↓

Business Rules

↓

Logging

---

# 20. RATE LIMITS

Anonymous

30/min

---

User

100/min

---

Premium

300/min

---

Admin

Separate Channel

---

# 21. DDOS STRATEGY

Используются:

---

Rate Limiter

---

Cache

---

CDN

---

Compression

---

Queue

---

Circuit Breaker

---

# 22. SQL SECURITY

Только параметризованные запросы.

---

Запрещены:

String Concatenation

Dynamic SQL

Raw User Input

---

# 23. XSS / CSRF

Используются:

---

CSP

---

HTTPOnly

---

SameSite

---

Secure Cookies

---

CSRF Tokens

---

# 24. FILE SECURITY

XML

PDF

ZIP

создаются

только в RAM.

---

После отправки уничтожаются.

---

# 25. EMAIL SECURITY

Email содержит:

---

минимум информации;

---

никаких сумм;

---

никаких налоговых данных в теле письма.

---

Только защищенное вложение.

---

# 26. BACKUP SECURITY

Backup:

AES-256

↓

Separate Storage

↓

Integrity Check

↓

Restore Test

---

# 27. DEPENDENCY SECURITY

Каждая зависимость проходит:

---

License Review

---

Security Review

---

Maintenance Review

---

Popularity Review

---

# 28. OPEN SOURCE POLICY

Используются только:

---

MIT

---

Apache 2.0

---

BSD

---

или совместимые лицензии.

---

# 29. TRUST & PRIVACY SCORE

Каждая новая функция получает оценку:

---

Privacy

---

Security

---

Complexity

---

Performance

---

Cost

---

Если Privacy Score ниже 95/100 —

функция не принимается.

---

# 30. USER PROMISE

OpenInvest никогда не продает данные пользователей.

---

OpenInvest никогда не требует обязательный паспорт.

---

OpenInvest никогда не требует обязательный ИНН.

---

OpenInvest никогда не хранит документы без явного согласия пользователя.

---

OpenInvest позволяет удалить все данные одним действием.

---

# FINAL CONSTITUTION

> **Доверие пользователей является самым дорогим активом OpenInvest.**

> **Любая функция, способная увеличить прибыль проекта ценой снижения приватности или безопасности пользователей, считается архитектурным дефектом и должна быть отклонена независимо от коммерческой выгоды.**

> **OpenInvest строится по принципу Zero Trust + Privacy by Design + Human in the Loop и рассматривает защиту капитала, данных и доверия пользователя как основную ценность продукта.**
