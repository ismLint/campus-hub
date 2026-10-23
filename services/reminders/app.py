import os
from datetime import datetime, timedelta, timezone

import requests
from flask import Flask, jsonify

app = Flask(__name__)

CORE_API = os.getenv("CORE_API_URL", "http://localhost:8080")
TG_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN", "")
TG_CHAT = os.getenv("TELEGRAM_CHAT_ID", "")


@app.get("/health")
def health():
    return {"status": "ok", "service": "reminders"}


def fetch_tasks():
    r = requests.get(f"{CORE_API}/tasks", timeout=3)
    r.raise_for_status()
    return r.json()


def parse_dt(s: str) -> datetime:
    if s.endswith("Z"):
        s = s[:-1] + "+00:00"
    dt = datetime.fromisoformat(s)
    return dt if dt.tzinfo else dt.replace(tzinfo=timezone.utc)


def send_telegram(text: str) -> bool:
    if not (TG_TOKEN and TG_CHAT):
        return False
    try:
        requests.post(
            f"https://api.telegram.org/bot{TG_TOKEN}/sendMessage",
            json={"chat_id": TG_CHAT, "text": text},
            timeout=5,
        )
        return True
    except Exception as e:
        print(f"[telegram] send failed: {e}")
        return False


@app.post("/check-reminders")
def check_reminders():
    try:
        tasks = fetch_tasks()
    except Exception as e:
        return jsonify(error="core-api unavailable", details=str(e)), 502

    if not isinstance(tasks, list):
        return jsonify(error="bad response from core-api", got=type(tasks).__name__), 502

    now = datetime.now(timezone.utc)
    horizon = now + timedelta(hours=24)
    reminders = []

    for t in tasks:
        if t.get("status") == "done":
            continue
        dl_str = t.get("deadline")
        if not dl_str:
            continue
        try:
            dl = parse_dt(dl_str)
        except Exception:
            continue
        if now <= dl <= horizon:
            hours_left = round((dl - now).total_seconds() / 3600, 1)
            text = (
                f"Напоминание: «{t['title']}» "
                f"(предмет: {t.get('subject', '—')}) — дедлайн через {hours_left} ч"
            )
            reminders.append(
                {
                    "task_id": t["id"],
                    "title": t["title"],
                    "deadline": dl.isoformat(),
                    "hours_left": hours_left,
                    "message": text,
                    "sent_to_telegram": send_telegram(text),
                }
            )

    for r in reminders:
        print(f"[reminder] {r['message']}")

    return jsonify(
        checked_at=now.isoformat(),
        horizon_hours=24,
        count=len(reminders),
        reminders=reminders,
    )



if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8083, debug=True)