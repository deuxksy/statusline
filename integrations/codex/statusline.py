"""A Codex statusline provider using newline-delimited MCP over stdio."""

from datetime import datetime
from pathlib import Path
from threading import Lock, Thread, Timer
from time import time
import json
import re
import subprocess
import sys


URI = "ui://statusline-tickets/footer"
TOOL = {
    "name": "statusline",
    "title": "Statusline · 초기화 티켓",
    "description": "Show the current model and agent state in the Codex footer.",
    "inputSchema": {"type": "object", "properties": {}},
    "_meta": {
        "openai/ui": {"entrypoints": [{"type": "statusline", "resourceUri": URI}]}
    },
}


class TicketSnapshot:
    def __init__(self, load):
        self.load = load
        self.lock = Lock()
        self.started = False
        self.value = None

    def get(self):
        with self.lock:
            if not self.started:
                self.started = True
                Thread(target=self._update, daemon=True).start()
            return self.value

    def _update(self):
        try:
            value = self.load()
        except (OSError, ValueError, TypeError, subprocess.SubprocessError):
            return
        with self.lock:
            self.value = value


def load_tickets():
    requests = [
        {"id": 1, "method": "initialize", "params": {
            "clientInfo": {"name": "statusline-tickets", "version": "1.0"}
        }},
        {"method": "initialized"},
        {"id": 2, "method": "account/rateLimits/read", "params": {
            "excludeResetCreditDetails": False
        }},
    ]
    with subprocess.Popen(
        ["codex", "-c", "mcp_servers.statusline-tickets.enabled=false",
         "app-server", "--stdio"],
        stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL,
        text=True,
    ) as process:
        # App-server 응답을 받기 전에 stdin을 닫으면 요청이 취소될 수 있다.
        timeout = Timer(8, process.kill)
        timeout.start()
        try:
            process.stdin.write("".join(json.dumps(request) + "\n" for request in requests))
            process.stdin.flush()
            for line in process.stdout:
                reply = json.loads(line)
                if reply.get("id") == 2:
                    return reply.get("result", {}).get("rateLimitResetCredits")
        finally:
            timeout.cancel()
            process.kill()


tickets = TicketSnapshot(load_tickets)
BINARY = Path(__file__).resolve().parents[2] / "statusline"
SGR = re.compile(r"\x1b\[([0-9;]*)m")


def ticket_fields(summary, now):
    count = (summary or {}).get("availableCount", 0)
    expirations = [
        row["expiresAt"] for row in (summary or {}).get("credits") or []
        if row.get("status") == "available"
        and isinstance(row.get("expiresAt"), (int, float))
        and row["expiresAt"] > now
    ]
    if count > 0 and expirations:
        try:
            return count, datetime.fromtimestamp(min(expirations)).date().isoformat()
        except (OverflowError, OSError, ValueError):
            pass
    return count, ""


def permission_name(thread):
    permissions = thread.get("permissions") or {}
    approval = thread.get("approval") or {}
    name = permissions.get("display_name")
    if name == "Workspace" and approval.get("policy") == "on-request":
        return {"user": "Ask for Approval", "auto-review": "Approve for me"}.get(
            approval.get("reviewer"), name
        )
    return name


def render(state):
    thread = state.get("thread", {})
    model = thread.get("model", {})
    context = thread.get("usage", {}).get("context", {})
    limits = next((item for item in state.get("account", {}).get("rate_limits", [])
                   if item.get("limit_id") == "codex"), {})
    windows = {}
    for key in ("primary_limit", "secondary_limit"):
        window = limits.get(key) or {}
        name = {300: "primary", 10080: "secondary"}.get(window.get("window_minutes"))
        if name and "used_percent" in window:
            windows[name] = {
                "used_percent": window["used_percent"],
                "resets_at": (window.get("resets_at") or 0) * 1000,
            }
    count, expires = ticket_fields(tickets.get(), time())
    payload = {
        "statusline_snapshot": True,
        "model": model.get("display_name", "Codex"),
        "effort": model.get("reasoning_level"),
        "cwd": thread.get("workspace", {}).get("cwd"),
        "context_used": context.get("used_percentage"),
        "context_token_count": context.get("used_tokens"),
        "info": {
            "model_context_window": context.get("context_window_size"),
        },
        "rate_limits": windows,
        "permission": permission_name(thread),
        "speed": model.get("service_tier_display_name") or (
            "Fast" if model.get("service_tier") in ("fast", "priority") else "Standard"
        ),
        "reset_tickets": count,
        "reset_tickets_expire": expires,
    }
    try:
        result = subprocess.run(
            [str(BINARY), "--cli=codex"], input=json.dumps(payload),
            capture_output=True, text=True, timeout=0.5, check=True,
        )
    except (OSError, subprocess.SubprocessError):
        return "statusline: Go renderer unavailable"
    return ui_tree(result.stdout.rstrip("\n"))


def ui_tree(output):
    # Codex는 연속 ASCII 공백을 한 칸으로 줄이므로 고정 폭 패딩에는 NBSP를 쓴다.
    output = re.sub(r" {2,}", lambda match: "\u00a0" * (len(match[0]) - 1) + " ", output)
    # ANSI 스타일을 Codex UI 트리의 텍스트와 배경 행으로 옮긴다.
    children = []
    props = {}
    background = None

    def append_text(value):
        if background is None and children and children[-1]["type"] == "row" and value.startswith(" "):
            value = "\u00a0" + value[1:]
        node = {"type": "text", "props": props.copy(), "children": [value]}
        if background is None:
            children.append(node)
            return
        node["children"] = [value.replace(" ", "\u00a0")]
        if (not children or children[-1]["type"] != "row"
                or children[-1]["props"]["background"] != background):
            children.append({"type": "row", "props": {"gap": 0, "background": background},
                             "children": []})
        children[-1]["children"].append(node)

    offset = 0
    for match in SGR.finditer(output):
        if match.start() > offset:
            append_text(output[offset:match.start()])
        codes = (match.group(1) or "0").split(";")
        while codes:
            code = codes.pop(0)
            if code == "0":
                props.clear()
                background = None
            elif code == "1":
                props["weight"] = "bold"
            elif code == "22":
                props.pop("weight", None)
            elif code == "39":
                props.pop("color", None)
            elif code == "49":
                background = None
            elif code in ("38", "48") and len(codes) >= 4 and codes[0] == "2":
                color = "#{:02x}{:02x}{:02x}".format(*map(int, codes[1:4]))
                if code == "38":
                    props["color"] = color
                else:
                    background = color
                del codes[:4]
        offset = match.end()
    if offset < len(output):
        append_text(output[offset:])
    # 텍스트 조각을 한 흐름으로 묶어 스타일 경계의 공백을 유지한다.
    return {"type": "text", "children": children}


def respond(method, params):
    if method == "initialize":
        return {
            "protocolVersion": "2025-06-18",
            "capabilities": {"tools": {}, "resources": {}},
            "serverInfo": {"name": "statusline-tickets", "version": "1.0"},
        }
    if method == "ping":
        return {}
    if method == "tools/list":
        return {"tools": [TOOL]}
    if method == "tools/call":
        if params.get("name") != TOOL["name"]:
            raise ValueError("Unknown tool name")
        return {
            "content": [{"type": "text", "text": f"The statusline resource is {URI}."}]
        }
    if method == "resources/list":
        return {
            "resources": [
                {"name": "footer", "uri": URI, "mimeType": "application/json"}
            ]
        }
    if method == "resources/templates/list":
        return {"resourceTemplates": []}
    if method == "resources/read":
        if params.get("uri") != URI:
            raise ValueError("Unknown resource URI")
        state = params.get("_meta", {}).get("com.openai/codex-statusline-state", {})
        return {
            "contents": [
                {
                    "uri": URI,
                    "mimeType": "application/json",
                    "text": json.dumps(render(state), ensure_ascii=False),
                }
            ]
        }
    raise KeyError(method)


def main():
    for line in sys.stdin.buffer:
        reply = {"jsonrpc": "2.0", "id": None}
        try:
            request = json.loads(line)
            if not isinstance(request, dict) or not isinstance(
                request.get("method"), str
            ):
                reply["error"] = {"code": -32600, "message": "Invalid request"}
            elif "id" not in request:
                continue
            else:
                reply["id"] = request["id"]
                reply["result"] = respond(request["method"], request.get("params", {}))
        except json.JSONDecodeError:
            reply["error"] = {"code": -32700, "message": "Parse error"}
        except KeyError:
            reply["error"] = {"code": -32601, "message": "Method not found"}
        except (TypeError, ValueError, AttributeError):
            reply["error"] = {"code": -32602, "message": "Invalid parameters"}
        print(json.dumps(reply), flush=True)


if __name__ == "__main__":
    main()
