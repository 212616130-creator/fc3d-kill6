#!/usr/bin/env python3
"""生成 golden.json 引擎基准（含个位自适应+回退队列状态机）
供 Go 引擎测试逐期断言一致性。仅开发期使用，不入库核心链路。
"""
import csv, json, random, sys

CSV = "fc3d-history.csv"
OUT = "golden.json"

# ── V9.3 全部公式 ─────────────────────
def kill_h(b, s, g):
    span = max(b, s, g) - min(b, s, g)
    if b % 2 == 0 and s % 2 == 0 and g % 2 == 0:  return (b + s + g + 1) % 10
    if b % 2 == 1 and s % 2 == 1 and g % 2 == 1:  return (b + s + g + 2) % 10
    if b == s:  return (3 * max(b, s, g)) % 10
    if b == g:  return (span + 1) % 10
    if s == g:  return (b + s + g + 8) % 10
    if span == 4:  return (b + s + g + 2) % 10
    if span >= 6:  return (b * g - s) % 10
    if (b + s + g) % 2 == 1:  return (b * b + s + g * g) % 10
    if b < g:  return (b + s + g + 2) % 10
    if b + s + g <= 12:  return (span + 3) % 10
    return (b + s + g + 1) % 10

def kill_t(b, s, g):
    if (b + s + g) % 2 == 1:
        if (b * b + s * s) % 10 == 0:  return (b + s + g + 2) % 10
        return (b * b + s * s + g) % 10
    if max(b, s, g) - min(b, s, g) >= 6:
        if b >= s and b >= g:  return ((b + s) * g) % 10
        return (3 * max(b, s, g)) % 10
    return (g * g + b) % 10

def _get_o_cond(b, s, g):
    sp = max(b, s, g) - min(b, s, g)
    if b % 2 == 1 and s % 2 == 1 and g % 2 == 1: return 'all_odd'
    if b == s: return 'b_eq_s'
    if b == g: return 'b_eq_g'
    if s == g: return 's_eq_g'
    if sp == 4: return 'span4'
    if sp == 2: return 'span2'
    if g == max(b, s, g): return 'g_max'
    if b > g: return 'b_gt_g'
    if b + s + g >= 15: return 'sum_hi'
    if (b + s + g) % 2 == 0: return 'sum_even'
    if (b + s + g) % 2 == 1: return 'sum_odd'
    return 'default'

O_BACKUP_FM = {
    'g_max': lambda b, s, g: (3 * max(b, s, g)) % 10,
    'b_gt_g': lambda b, s, g: (b * b + g) % 10,
    'sum_hi': lambda b, s, g: (b + s + g + 3) % 10,
    'sum_odd': lambda b, s, g: (b + s + g + 1) % 10,
    'default': lambda b, s, g: (b + s + g + 1) % 10,
}
O_FAIL_WIN = 5

def kill_o(b, s, g, fail_state=None, period_idx=None):
    span = max(b, s, g) - min(b, s, g)
    if b % 2 == 1 and s % 2 == 1 and g % 2 == 1:  pk = (b + s + g + 3) % 10
    elif b == s:  pk = (b + s + g + 6) % 10
    elif b == g:  pk = (b + s + g + 2) % 10
    elif s == g:  pk = (b + s + g + 1) % 10
    elif span == 4:  pk = (b * b + s * s + g) % 10
    elif span == 2:  pk = (s * g + b) % 10
    elif g == max(b, s, g):  pk = (s * g + b) % 10
    elif b > g:  pk = (s * g) % 10
    elif b + s + g >= 15:  pk = (b * s + s * g) % 10
    elif (b + s + g) % 2 == 0:  pk = (s * g + b) % 10
    elif (b + s + g) % 2 == 1:  pk = (g * g * s) % 10
    else:  pk = (s * g - b) % 10
    if fail_state is not None and period_idx is not None:
        cn = _get_o_cond(b, s, g)
        if cn in fail_state and period_idx - fail_state[cn] <= O_FAIL_WIN:
            if cn in O_BACKUP_FM:
                pk = O_BACKUP_FM[cn](b, s, g) % 10
    return pk

def kill_h2(b, s, g):
    span = max(b, s, g) - min(b, s, g)
    return (b - span + 9) % 10

def kill_t2(b, s, g):
    mid = sorted([b, s, g])[1]
    return (s - mid + 5) % 10

def kill_o2(b, s, g):
    return (g * g + abs(b - g)) % 10

H_FB = [lambda b, s, g: (b + s + g + 1) % 10, lambda b, s, g: (b * s) % 10]
T_FB = [lambda b, s, g: (g * g + b) % 10, lambda b, s, g: (b + s + g + 1) % 10,
        lambda b, s, g: max(b, s, g) - min(b, s, g), lambda b, s, g: (b * g) % 10,
        lambda b, s, g: (b + s) % 10, lambda b, s, g: (b * s) % 10]
O_FB = [lambda b, s, g: (b + s + g + 1) % 10, lambda b, s, g: (b * s) % 10]

def apply_fb(kill, prev, fb_list, b, s, g):
    if kill != prev:
        return kill
    for f in fb_list:
        alt = f(b, s, g) % 10
        if alt != prev:
            return alt
    return (kill + 1) % 10

# ── 读取 CSV ─────────────────────────────────────────
draws = []
with open(CSV, encoding="utf-8") as f:
    for r in csv.DictReader(f):
        draws.append({"issue": r["issue"].strip(), "b": int(r["hundreds"]),
                      "s": int(r["tens"]), "g": int(r["ones"])})

# ── 全量逐期状态机──
rows = []
phk = ptk = pok = None
o_fail = {}
for i in range(1, len(draws)):
    p = draws[i - 1]
    b, s, g = p["b"], p["s"], p["g"]
    phk = apply_fb(kill_h(b, s, g), phk, H_FB, b, s, g) if phk is not None else kill_h(b, s, g)
    ptk = apply_fb(kill_t(b, s, g), ptk, T_FB, b, s, g) if ptk is not None else kill_t(b, s, g)
    pok_raw = kill_o(b, s, g, o_fail, i)
    pok = apply_fb(pok_raw, pok, O_FB, b, s, g) if pok is not None else pok_raw
    if pok == draws[i]["g"]:
        cn = _get_o_cond(b, s, g)
        o_fail[cn] = i
    rows.append({
        "issue": draws[i]["issue"],
        "hK": phk, "tK": ptk, "oK": pok,
        "hK2": kill_h2(b, s, g), "tK2": kill_t2(b, s, g), "oK2": kill_o2(b, s, g),
    })

# ── 纯函数随机输入（无状态，覆盖全 1000 种组合 + 1000 随机）──
rnd = []
for b in range(10):
    for s in range(10):
        for g in range(10):
            rnd.append({"b": b, "s": s, "g": g,
                        "h": kill_h(b, s, g), "t": kill_t(b, s, g), "o": kill_o(b, s, g),
                        "h2": kill_h2(b, s, g), "t2": kill_t2(b, s, g), "o2": kill_o2(b, s, g)})

out = {
    "source": "fc3d-kill6 V9.3 引擎基准 (tools/gen_golden.py)",
    "total": len(rows),
    "rows": rows,
    "exhaustive": rnd,
}
with open(OUT, "w", encoding="utf-8") as f:
    json.dump(out, f, ensure_ascii=False)
print(f"golden.json 生成完成: {len(rows)} 期逐期 + {len(rnd)} 组全量穷举 (1000组合)")
