# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 21

# SECURITY, PRIVACY BY DESIGN, ZERO TRUST, COMPLIANCE & USER TRUST

Version: 1.0

Status: Approved

Priority: CRITICAL

Classification: MUST IMPLEMENT

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_21_SECURITY_PRIVACY_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `961aee8245d086cf74170eacebdb29f0a846a4c6`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the entire OpenInvest security architecture.

This is one of the most important documents in the project.

OpenInvest should stand out not only through analytics, but also through user trust.

---

# PRODUCT PHILOSOPHY

Most investment services say:

> "Give us all your data, and we will protect it."

OpenInvest says otherwise:

> **"It is better not to store your data at all when it is possible to avoid doing so."**

---

# TRUST FIRST

The product's primary value:

```
Trust

>

Features
```

---

# PRIVACY BY DESIGN

Every new feature must answer the question:

```
Can this be done

WITHOUT storing

personal data?
```

If the answer is "Yes",

storing personal data is prohibited.

---

# ZERO TRUST ARCHITECTURE

No one inside the system

is considered trusted.

Even an internal service must:

* authenticate;
* be logged;
* have restricted permissions.

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

Investment data

must never

be stored together

with personal data.

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

Storage:

Argon2id

or

bcrypt

---

Prohibited:

SHA1

SHA256

MD5

---

# AUTHENTICATION

Supported:

Email

Google (future)

Apple (future)

Telegram (future)

Passkeys (future)

---

# SESSION MODEL

Access Token

15 minutes

---

Refresh Token

30 days

---

# DEVICE BINDING

Each session is bound to:

```
Browser

OS

Device ID

Approximate Region
```

---

# LOGIN NOTIFICATION

When signing in from a new device:

Email

*

Push

---

# TWO FACTOR AUTHENTICATION

Optional.

---

Support:

TOTP

Authenticator

Passkeys

---

SMS is not a mandatory method.

---

# API SECURITY

All APIs operate only over HTTPS.

---

Minimum TLS version:

```
TLS 1.3
```

---

# API KEYS

Never

reach the Frontend.

---

All keys:

Backend Only

---

# SECRET STORAGE

```
.env

Vault

Secret Manager
```

---

Prohibited:

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

Fields:

INN

Passport

Address

are encrypted separately.

---

# PRIVATE MODE

The user can select:

```
☑ Do not store passport

☑ Do not store INN

☑ Do not store address

☑ Do not store phone
```

---

# TEMPORARY MODE

The user enters:

INN

Passport

Address

---

OpenInvest:

generates a declaration

↓

sends XML/PDF

↓

completely destroys the data

from process memory.

---

# NEVER STORE

By default, storing the following is prohibited:

passport;

INN;

address;

place of registration.

---

# USER CONSENT

Each checkbox:

separate.

---

Prohibited:

```
☑ I agree to everything
```

---

# RIGHT TO DELETE

The profile always contains a button:

```
Delete profile completely
```

---

After confirmation,

the following are deleted:

personal data;

sessions;

exports;

notifications.

---

Investment records:

are anonymized.

---

# EXPORT MY DATA

The user can download:

```
JSON

CSV

XML

PDF
```

---

# AUDIT

Every action is recorded.

---

But Audit

does not contain:

passwords;

passport;

INN.

---

# EMAIL SECURITY

All emails are signed.

---

SPF

DKIM

DMARC

---

# FILE SECURITY

PDF

XML

ZIP

are created

only in RAM.

---

They are never

stored

in the server's temporary directory.

---

# DOWNLOAD LINKS

Signed links.

---

TTL

15 minutes.

---

After that:

automatic deletion.

---

# RATE LIMIT

Authorization

5 attempts

per 15 minutes.

---

Password Reset

3 attempts

per hour.

---

Tax Export

10 requests

per day.

---

# DDOS PROTECTION

Cloudflare

or equivalent.

---

# BOT PROTECTION

Invisible CAPTCHA

Behavior Analysis

Rate Limiting

---

# SQL INJECTION

Use only:

Prepared Statements

ORM

Query Builder

---

Prohibited:

String Concatenation.

---

# XSS

All data is escaped.

---

# CSP

Content Security Policy

is mandatory.

---

# CSRF

The following are used:

SameSite Cookies

CSRF Tokens

---

# SECURITY HEADERS

Mandatory:

```
HSTS

X-Frame-Options

Referrer-Policy

Permissions-Policy

X-Content-Type-Options
```

---

# LOGGING POLICY

Never log:

passwords;

JWT;

Refresh;

INN;

passport;

address.

---

# BACKUPS

Encrypted

---

Automatic

---

Daily

---

Retention

30 days

---

# INCIDENT RESPONSE

When a compromise is detected:

1.

disable the API;

2.

revoke all Refresh Token;

3.

notify users;

4.

create an Audit Report.

---

# THIRD PARTY POLICY

Before integrating any service, verify:

official status;

license;

GDPR;

local legislation;

data retention policy.

---

# AI SECURITY

AI is prohibited from:

generating tax data without verification;

providing investment recommendations;

changing user data independently;

signing documents.

---

# HUMAN IN THE LOOP

Any tax document:

XML

PDF

Email

must be confirmed by the user.

---

# USER TRUST PAGE

A separate section is created in the application:

## Why OpenInvest is secure

It shows:

```
What data we collect

What data we DO NOT collect

How it is encrypted

How to delete a profile

How to export data

How a declaration is generated

How audit logs work
```

---

# COMPETITIVE ADVANTAGE

Most brokers require:

passport;

INN;

address;

phone;

dozens of mandatory fields.

---

OpenInvest allows use:

without a passport;

without INN;

without an address;

without storing personal data.

---

The user chooses:

```
Private Mode

or

Convenience Mode
```

---

# TRUST SCORE

For every new feature, Builder Agent must check:

1.

Does it increase user trust?

2.

Can it be implemented without new personal data?

3.

Can it be done locally?

4.

Can data storage be avoided?

5.

Can all data be deleted with one click?

6.

Would it be comfortable to explain this feature to an auditor, a lawyer, and an ordinary user?

If even one answer is negative, the feature is sent for architectural reconsideration.

---

# FINAL PRINCIPLE

> **OpenInvest does not sell investment ideas. OpenInvest sells peace of mind, transparency, and control over one's own finances.**

> **The user always owns their data, understands every number, and can leave the system without losing control over their information.**
