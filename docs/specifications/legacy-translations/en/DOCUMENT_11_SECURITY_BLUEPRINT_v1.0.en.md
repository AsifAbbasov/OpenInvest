# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 11

# SECURITY, TRUST, PRIVACY & COMPLIANCE BLUEPRINT

Version: 1.0

Status: Mandatory

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_11_SECURITY_BLUEPRINT_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `f647beb1bdbf4cb6b565a8464f697f1e8907cb41`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. DOCUMENT PURPOSE

This document defines the OpenInvest security architecture.

It is one of the project's primary documents.

No module may be implemented in violation of the requirements of this document.

Security is part of the architecture, not an additional feature.

---

# 2. CORE PHILOSOPHY

OpenInvest is built on the principle of

Trust by Design

and

Privacy by Design.

We are not simply selling the user a dividend calculator.

We are selling trust.

---

# 3. OUR MAIN COMPETITIVE DIFFERENTIATOR

Most investment services collect the maximum possible amount of data.

OpenInvest collects the minimum necessary amount of data.

The user fully controls:

• what is stored;

• where it is stored;

• how long it is stored;

• when it is deleted;

• to whom it is sent.

---

# 4. PRINCIPLES

## Principle 1

User owns data.

Not the company.

Not the developer.

Not AI.

Not the server.

---

## Principle 2

Minimal Collection.

If a function can work without personal data,

requesting personal data is prohibited.

---

## Principle 3

Explicit Consent.

Any storage of personal information requires separate consent.

---

## Principle 4

Delete Anytime.

Any data must be removable with one user action.

---

## Principle 5

Transparency.

The user always knows:

what is stored;

when it is stored;

why it is stored;

when it will be deleted.

---

# 5. DATA CLASSIFICATION

## Public

quotes

dividends

market

news

indices

---

## Internal

portfolio

transactions

notifications

watchlist

---

## Sensitive

email

tax profile

passport

INN

address

tax residency

---

## Critical

password hash

refresh token

session

API secret

SMTP secret

JWT secret

---

# 6. MINIMAL PROFILE

To use the service, the following are sufficient:

email

password

timezone

language

currency

THAT IS ALL.

---

# 7. OPTIONAL FIELDS

The user may choose NOT to provide:

passport

INN

SNILS

address

registration

place of birth

phone

tax residency

broker data

---

# 8. TAX PROFILE

Three modes.

---

## Mode A

No Tax Profile

OpenInvest generates an empty declaration.

The user fills everything in independently.

No data is stored.

---

## Mode B

Temporary Tax Profile

The user enters:

full name

INN

address

passport

the system generates XML/PDF

sends it by email

immediately deletes the information.

---

## Mode C

Persistent Tax Profile

The user permits data storage.

All fields can be changed.

All fields can be exported.

All fields can be deleted.

Disabled by default.

---

# 9. CONSENT

All sensitive data has separate checkboxes.

☐ store INN

☐ store passport

☐ store address

☐ reuse

No preselected checkboxes.

---

# 10. TRUST DASHBOARD

Every user has a screen:

Security & Privacy

---

It shows:

Password updated

Last login

Active devices

Active browsers

Active tokens

Last export

Last data deletion

Tax profile

Passport stored?

INN stored?

Address stored?

Email verified?

2FA enabled?

---

# 11. EXPORT

The user can export:

portfolio

dividends

transactions

settings

notifications

watchlist

tax files

audit

---

Formats:

CSV

JSON

XML

PDF

ZIP

---

# 12. DELETE

Deletion must work with one action.

Delete:

portfolio

tax profile

personal data

notifications

sessions

entire account

---

After deletion:

recovery is impossible.

---

# 13. PASSWORD

Argon2id

unique salt

minimum 12 characters

breach check

---

# 14. AUTHENTICATION

JWT Access

Refresh Token

Device ID

Session ID

Rotation

Expiration

---

# 15. TWO FACTOR

Supported:

Email

Authenticator

TOTP

Passkeys (in the future)

---

# 16. DATABASE ENCRYPTION

Sensitive fields:

AES-256

---

Refresh Tokens

Hash

---

Passwords

Argon2id

---

# 17. DATABASE SPLIT

Storing investment and identity data together is prohibited.

users

profiles

tax_profiles

transactions

snapshots

audit_logs

are separated.

---

# 18. AUDIT

Every action is recorded.

Date

IP

Device

Browser

Action

Result

RequestID

---

Deleting logs is prohibited.

Only archiving is permitted.

---

# 19. AI TRANSPARENCY

AI never independently changes:

portfolio

taxes

return

dividends.

---

Every AI output is accompanied by:

Source

Date

Confidence

Reason

---

# 20. HUMAN IN THE LOOP

Mandatory confirmation:

XML generation

PDF generation

email sending

tax export

mass deletion

portfolio import

---

# 21. OFFICIAL DATA ONLY

Only official or permitted sources are used.

MOEX

CBR

Rosstat

e-disclosure

Issuer websites

---

# 22. RATE LIMIT STRATEGY

100 000 users

must not create

100 000 requests.

---

The following scheme is used:

MOEX

↓

Parser

↓

Go Cache

↓

PostgreSQL

↓

API

↓

100 000 clients

---

# 23. CACHE POLICY

The exchange is queried by the server.

The user never contacts it directly.

---

# 24. DDOS PROTECTION

Rate Limit

Circuit Breaker

Retry

Backoff

Queue

---

# 25. EMAIL SECURITY

SMTP credentials

never reach the frontend.

All operations are performed by the server.

---

# 26. XML SECURITY

XML is created only in memory.

After generation:

sent

or

downloaded

or

deleted.

---

# 27. PDF SECURITY

PDF is not stored permanently.

Generation

↓

sending

↓

automatic deletion.

---

# 28. ZIP GENERATION

archive/zip

bytes.Buffer

RAM only

no temporary files.

---

# 29. API SECURITY

All endpoints:

JWT

Rate Limit

Validation

Audit

Authorization

---

# 30. LOGGING

PII must not be written to logs.

Prohibited:

passport

INN

address

full email

password

refresh token

---

# 31. MOBILE SECURITY

Android

EncryptedSharedPreferences

---

iOS

Keychain

---

no tokens in ordinary Storage.

---

# 32. DEVOPS SECURITY

Secrets:

.env

GitHub Secrets

Vault

---

Prohibited:

commit API keys

commit passwords

commit SMTP credentials

---

# 33. OPEN SOURCE POLICY

Before using a library, check:

License

Popularity

Maintenance

Security

Performance

Bundle Size

---

# 34. DEPENDENCY POLICY

Prohibited:

unsupported packages

dead projects

GPL without necessity

questionable npm libraries

---

# 35. PRIVACY ADVANTAGE

Main marketing thesis:

"OpenInvest does not require a passport, INN, or address to work.

You choose your own level of privacy.

Your investments belong to you."

---

# 36. USER AGREEMENT REQUIREMENTS

Codex must create:

User Agreement

Privacy Policy

Cookie Policy

Data Processing Policy

Disclaimer

AI Disclaimer

Investment Disclaimer

Tax Disclaimer

---

# 37. LEGAL POSITION

OpenInvest:

is not a broker;

is not an investment adviser;

is not a tax consultant;

does not guarantee returns;

does not execute trades;

does not hold users' money.

---

# 38. INCIDENT RESPONSE

When an incident is detected:

1. block sessions;

2. rotate secrets;

3. notify the user;

4. investigate;

5. publish RCA;

6. eliminate the cause.

---

# 39. TRUST AS PRODUCT FEATURE

OpenInvest should compete not by number of functions,

but by level of trust.

After registration, the user should understand:

what is stored;

what is not stored;

what is deleted automatically;

that they have full control over their data.

---

# 40. SUCCESS METRIC

If a user can use OpenInvest:

without entering a passport,

without entering INN,

without entering an address,

without entering broker data,

while still receiving full portfolio analytics,

a dividend calendar,

XIRR calculation,

inflation adjustment,

tax XML/PDF,

then the Trust & Privacy architecture is considered correctly implemented.
