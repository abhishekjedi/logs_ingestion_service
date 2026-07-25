#!/usr/bin/env python3
"""Publish realistic, varied OTLP logs to a service for demoing the dashboard.

Each error is emitted as the tail of a short user *journey*: a handful of
breadcrumb logs (page views, clicks, api calls) that share the error's
session.id and precede it in time, plus rich request/client attributes. This is
what powers the stacktrace / context / breadcrumbs panels in the UI.

Usage: python3 scripts/publish_test_logs.py <ingest_url> <api_key>
"""
import json
import random
import sys
import time
import urllib.request

URL, KEY = sys.argv[1], sys.argv[2]

# (exception_type, message, stacktrace, count, severity)
TEMPLATES = [
    ("TypeError", "Cannot read property 'id' of undefined",
     "at Object.handler (src/routes/checkout.js:88:14)\nat Layer.handle (src/middleware/auth.js:42:9)", 340, 17),
    ("DatabaseError", "connection timeout after 5000ms",
     "at Pool.query (src/db/pool.js:120:11)\nat OrderService.create (src/services/order.js:56:20)", 185, 17),
    ("ValidationError", "invalid email format",
     "at validate (src/validators/user.js:33:7)\nat SignupController.post (src/controllers/signup.js:18:5)", 96, 17),
    ("TimeoutError", "upstream request timed out",
     "at fetchWithTimeout (src/http/client.js:77:13)\nat PaymentGateway.charge (src/payments/gateway.js:44:9)", 61, 17),
    ("KeyError", "missing config key 'stripe_api_key'",
     "at loadConfig (src/config/loader.js:29:15)", 24, 21),
    ("PaymentDeclined", "card declined: insufficient_funds",
     "at PaymentGateway.charge (src/payments/gateway.js:51:11)", 13, 17),
]
INFO_LOGS = [("checkout completed", 9, 130), ("user signed in", 9, 110), ("cache warmed", 5, 60)]

ENVS = ["production", "production", "production", "staging"]
RELEASES = ["v2.3.1", "v2.3.1", "v2.3.0"]
SEV_TEXT = {5: "DEBUG", 9: "INFO", 13: "WARN", 17: "ERROR", 21: "FATAL"}

BROWSERS = ["Chrome 126", "Safari 17", "Firefox 127", "Edge 126"]
OSES = ["macOS 14.5", "Windows 11", "iOS 17.5", "Android 14"]
ROUTES = ["/", "/products", "/products/42", "/cart", "/checkout", "/checkout/pay", "/account"]

# The trail of steps that leads up to an error, as (severity, body) breadcrumbs.
JOURNEY = [
    (9, "page viewed: /products"),
    (5, "clicked: add-to-cart"),
    (9, "page viewed: /cart"),
    (9, "page viewed: /checkout"),
    (13, "form validation warning: expiry near"),
    (5, "clicked: place-order"),
]

now_ns = int(time.time() * 1e9)
SPREAD = int(12 * 3600 * 1e9)  # last 12 hours


def record(sev, body, attrs, ts):
    return {
        "timeUnixNano": str(ts),
        "severityNumber": sev,
        "severityText": SEV_TEXT.get(sev, "INFO"),
        "body": {"stringValue": body},
        "attributes": [{"key": k, "value": {"stringValue": v}} for k, v in attrs.items()],
    }


def post(records):
    body = json.dumps({"resourceLogs": [{"scopeLogs": [{"logRecords": records}]}]}).encode()
    req = urllib.request.Request(URL, data=body, method="POST",
                                 headers={"Content-Type": "application/json", "X-API-Key": KEY})
    with urllib.request.urlopen(req) as r:
        return r.status


def session_attrs():
    """Stable per-session identity + client context shared by a whole journey."""
    return {
        "user.id": f"user-{random.randint(1, 80)}",
        "session.id": f"sess-{random.randint(1, 4000)}",
        "deployment.environment": random.choice(ENVS),
        "service.version": random.choice(RELEASES),
        "browser.name": random.choice(BROWSERS),
        "os.name": random.choice(OSES),
        "client.address": f"203.0.113.{random.randint(1, 254)}",
    }


batch, sent = [], 0


def flush():
    global batch, sent
    if batch:
        post(batch)
        sent += len(batch)
        batch = []


def add(rec):
    batch.append(rec)
    if len(batch) >= 25:
        flush()


for etype, msg, stack, count, sev in TEMPLATES:
    for _ in range(count):
        sess = session_attrs()
        err_ts = now_ns - random.randint(0, SPREAD)

        # Breadcrumb trail: a few journey steps BEFORE the error, same session.
        trail = JOURNEY[-random.randint(2, len(JOURNEY)):]
        step_gap = int(random.randint(2, 20) * 1e9)  # 2-20s between steps
        for i, (bsev, body) in enumerate(trail):
            steps_before = len(trail) - i
            add(record(bsev, body, dict(sess), err_ts - steps_before * step_gap))

        # The error itself, with a request context on top of the session.
        route = random.choice(ROUTES)
        a = dict(sess)
        a.update({
            "exception.type": etype,
            "exception.message": msg,
            "exception.stacktrace": stack,
            "http.request.method": random.choice(["GET", "POST", "POST", "PUT"]),
            "url.path": route,
            "url.full": f"https://shop.example.com{route}",
            "http.response.status_code": str(random.choice([500, 500, 502, 400, 408])),
        })
        add(record(sev, msg, a, err_ts))
    flush()

for body, sev, count in INFO_LOGS:
    for _ in range(count):
        add(record(sev, body, session_attrs(), now_ns - random.randint(0, SPREAD)))
    flush()

print(f"published {sent} events across {len(TEMPLATES)} error types + {len(INFO_LOGS)} info streams")
