# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 30

# SECURITY, ZERO TRUST, PRIVACY BY DESIGN, DATA PROTECTION, CRYPTOGRAPHY & USER TRUST CONSTITUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: SECURITY FOUNDATION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_30_SECURITY_ZERO_TRUST_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `de138367aadaa8a99b65758b0093392700c7b562`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the fundamental security principles of OpenInvest.

Security is considered not as a separate function, but as an integral part of the architecture.

Any decision that improves functionality but worsens user security or privacy is automatically considered incorrect.

---

# 1. MAIN PRINCIPLE

OpenInvest must not require trust.

OpenInvest must be designed so that the user can use the product even without fully trusting it.

---

# 2. ZERO TRUST ARCHITECTURE

No one is considered trusted.

Not the user.

Not the server.

Not an internal service.

Not AI.

Not an administrator.

---

Every request undergoes repeated verification.

---

# 3. PRIVACY BY DESIGN

Every new feature is evaluated with five questions:

---

Can these data be not collected at all?

---

Can they be anonymized?

---

Can they be deleted immediately after use?

---

Can the user be given a choice?

---

Can the feature be implemented without storing data?

---

If the answer is "Yes", the data are not stored.

---

# 4. MINIMAL DATA COLLECTION

OpenInvest collects the minimum possible amount of information.

---

Required data:

Email

Password Hash

Settings

Portfolio

Transactions

---

Optional:

Name

TIN

Passport

Address

Phone

Date of birth

---

By default, they are absent.

---

# 5. PRIVATE MODE

The user can work completely anonymously.

---

Available:

Portfolio

Dividends

XIRR

Inflation

Real Return

Calendar

Charts

---

Unavailable:

Automatic filling of the declaration with personal data.

---

# 6. TAX MODE

There are three modes.

---

## Anonymous

The user receives a blank declaration.

Fills it in manually.

---

## Assisted

The user enters data once.

After generation, the documents are automatically cleared.

---

## Saved

The user consciously permits data storage.

---

# 7. DELETE POLICY

Any personal data are deleted:

---

at the user's request;

---

when the account is closed;

---

after the retention period expires.

---

Deletion is irreversible.

---

# 8. EXPORT POLICY

The user can always download:

---

the entire profile;

---

the portfolio;

---

history;

---

tax documents;

---

settings.

---

Formats:

JSON

CSV

PDF

ZIP

---

# 9. DATA SEPARATION

Investment data

are never

mixed

with personal data.

---

```id="opr8ak"
Identity Database

↓

UUID

↓

Investment Database
```

---

Even if one database is compromised, it is impossible to reconstruct the other.

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

Passwords are never stored.

---

The following are used:

Argon2id

Salt

Memory Cost

Time Cost

Parallelism

---

# 12. SECRET MANAGEMENT

It is prohibited to store:

API Keys

SMTP

JWT

DB Passwords

in Git.

---

The following are used:

Environment Variables

or

Secrets Manager.

---

# 13. SESSION SECURITY

Access Token

15 minutes.

---

Refresh Token

30 days.

---

Rotation is mandatory.

---

# 14. DEVICE MANAGEMENT

The user sees:

---

active devices;

---

last login;

---

browser;

---

operating system.

---

Can terminate any session.

---

# 15. AUDIT LOG

All critical actions are recorded.

---

Login

---

Data deletion

---

XML creation

---

Email change

---

Password change

---

Profile export

---

# 16. AUDIT IMMUTABILITY

Audit Log must not be modified.

---

Only the following is allowed:

Append Only.

---

# 17. HUMAN IN THE LOOP

AI has no right to:

---

submit a declaration independently;

---

modify the portfolio;

---

create transactions;

---

modify personal data.

---

The final action is always confirmed by the user.

---

# 18. AI SANDBOX

AI works only with a copy of the data.

---

AI never receives:

password;

Refresh Token;

JWT;

SMTP;

keys;

secrets.

---

# 19. API SECURITY

All APIs go through:

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

The following are used:

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

Only parameterized queries.

---

Prohibited:

String Concatenation

Dynamic SQL

Raw User Input

---

# 23. XSS / CSRF

The following are used:

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

are created

only in RAM.

---

After sending, they are destroyed.

---

# 25. EMAIL SECURITY

Email contains:

---

a minimum of information;

---

no amounts;

---

no tax data in the body of the email.

---

Only a protected attachment.

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

Every dependency undergoes:

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

Only the following are used:

---

MIT

---

Apache 2.0

---

BSD

---

or compatible licenses.

---

# 29. TRUST & PRIVACY SCORE

Every new feature receives a score for:

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

If Privacy Score is below 95/100 —

the feature is not accepted.

---

# 30. USER PROMISE

OpenInvest never sells user data.

---

OpenInvest never requires a mandatory passport.

---

OpenInvest never requires a mandatory TIN.

---

OpenInvest never stores documents without the user's explicit consent.

---

OpenInvest allows all data to be deleted with one action.

---

# FINAL CONSTITUTION

> **User trust is OpenInvest's most valuable asset.**

> **Any feature capable of increasing project profit at the cost of reducing user privacy or security is considered an architectural defect and must be rejected regardless of commercial benefit.**

> **OpenInvest is built on the principle of Zero Trust + Privacy by Design + Human in the Loop and treats protection of the user's capital, data, and trust as the product's core value.**
