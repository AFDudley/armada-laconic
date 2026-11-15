#!/usr/bin/env python3
"""Convert `def P … : Prop := body` predicates of an armada-lean checkout to
one-field structures, add `P.iff`, and repair the proofs that break.

See README.md beside this file for the design. Usage:

    python3 codemod.py --checkout <armada-lean worktree> --slot <1-4> \\
        --names RegistryConsistent,Adjudicator.IsDAG --report report.json

Steps, in order:

1. `lake build` the unchanged tree, then `scripts/type_lint.lean` (the
   "before" counts and finding list).
2. Discovery: `discover.lean` prints each predicate's declaration range
   and the pieces of its `def` syntax.
3. Guard: the predicate is not under `ArmadaLean/Spec/`, and
   `git grep "Spec\\.<name>\\b" -- ArmadaLean/Spec` finds nothing.
4. Edit: the `def` becomes `structure … where holds : body`; `P.iff`
   follows it.
5. Repair loop: `lake build`; for each failing module, the probe
   (`probe/Main.lean`, a Lean executable) elaborates the file and prints
   its messages with ranges, plus every numeric projection `t.1` on a
   converted type; `repairs.match` maps each error to one fixed pattern;
   the edits apply; repeat. An error that matches no pattern, or a repair
   that would repeat, stops the tool. The tool then changes nothing more
   and reports the message.
6. `docs/counterexamples/*.lean`: the probe on each, repaired the same
   way.
7. `scripts/type_lint.lean` again (the "after" list), `python3
   scripts/update_gate_baseline.py --hashes-only`, `bash scripts/check.sh
   --json`, and the JSON report.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import time
from dataclasses import asdict
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from repairs import Edit, Message, Predicate, Repair, Source, apply_edits, match  # noqa: E402

HERE = Path(__file__).resolve().parent
PROBE_DIR = HERE / "probe"
FORBIDDEN = ("sorry", "admit", "axiom", "native_decide")


class Stop(Exception):
    """The tool stops: the message says why. Nothing more changes."""


class Lean:
    """Runs Lean commands in one checkout under one `flock` slot."""

    def __init__(self, root: Path, slot: int, log_dir: Path) -> None:
        self.root = root
        self.slot = slot
        self.log_dir = log_dir
        self.lock = Path.home() / ".cache" / "lean-run" / "locks" / f"slot-{slot}.lock"
        self.env = dict(os.environ)
        self.env["PATH"] = f"{Path.home()}/.elan/bin:" + self.env.get("PATH", "")
        self.env.pop("CI", None)
        self.timings: list[dict] = []

    def run(self, args: list[str], log_name: str, cwd: Path | None = None) -> subprocess.CompletedProcess:
        cmd = ["flock", str(self.lock), "nice", *args]
        t0 = time.monotonic()
        proc = subprocess.run(cmd, cwd=cwd or self.root, env=self.env, capture_output=True, text=True)
        dt = round(time.monotonic() - t0, 1)
        self.timings.append({"command": " ".join(args), "seconds": dt, "exit": proc.returncode})
        (self.log_dir / log_name).write_text(proc.stdout + "\n--- stderr ---\n" + proc.stderr)
        print(f"[lean] {' '.join(args)[:80]} … exit {proc.returncode} in {dt}s", file=sys.stderr)
        return proc

    def build(self, tag: str) -> tuple[bool, str]:
        """`lake build`, then the linter modules the gate builds too."""
        p1 = self.run(["lake", "build"], f"build-{tag}.log")
        out = p1.stdout + p1.stderr
        if p1.returncode != 0:
            return False, out
        p2 = self.run(["lake", "build", "ArmadaLean.Lint.TypeDoctrine",
                       "ArmadaLean.Lint.TypeDoctrineTest"], f"build-lint-{tag}.log")
        return p2.returncode == 0, out + p2.stdout + p2.stderr

    def ensure_probe(self) -> Path:
        """Build `probe/` once per toolchain; refuse a toolchain mismatch."""
        want = (self.root / "lean-toolchain").read_text().strip()
        have = (PROBE_DIR / "lean-toolchain").read_text().strip()
        if want != have:
            raise Stop(f"probe/lean-toolchain is {have}; the checkout uses {want}. "
                       "Set probe/lean-toolchain to the checkout's toolchain and rerun.")
        exe = PROBE_DIR / ".lake" / "build" / "bin" / "probe"
        srcs = [PROBE_DIR / "Main.lean", PROBE_DIR / "lakefile.lean"]
        if not exe.exists() or any(s.stat().st_mtime > exe.stat().st_mtime for s in srcs):
            p = self.run(["lake", "build"], "probe-build.log", cwd=PROBE_DIR)
            if p.returncode != 0 or not exe.exists():
                raise Stop(f"the probe did not build (exit {p.returncode}):\n{p.stdout[-2000:]}\n"
                           f"{p.stderr[-2000:]}")
        return exe

    def probe(self, rel: str, preds: list[Predicate], tag: str) -> tuple[list[Message], list[dict]]:
        """Elaborate `rel`: its messages, and every `t.1` that resolves to a
        converted predicate's `holds` projection."""
        exe = self.ensure_probe()
        names = ",".join(p.full_name for p in preds)
        proc = self.run(["lake", "env", str(exe), rel, names], f"probe-{tag}.log")
        msgs: list[Message] = []
        projections: list[dict] = []
        summary = None
        for line in proc.stdout.splitlines():
            if not line.startswith("{"):
                continue
            d = json.loads(line)
            if d.get("summary"):
                summary = d
            elif d.get("projection"):
                projections.append(d)
            else:
                end = d.get("endPos") or {}
                msgs.append(Message(
                    file=rel, line=d["pos"]["line"], col=d["pos"]["column"],
                    end_line=end.get("line"), end_col=end.get("column"),
                    severity=d["severity"], text=d["data"]))
        if summary is None:
            raise Stop(f"the probe on {rel} exited {proc.returncode} without a summary:\n"
                       f"{proc.stdout[-2000:]}\n{proc.stderr[-2000:]}")
        return msgs, projections

    def type_lint(self, tag: str) -> tuple[list[str], dict]:
        proc = self.run(["lake", "env", "lean", "scripts/type_lint.lean"], f"lint-{tag}.log")
        findings = [ln for ln in proc.stdout.splitlines()
                    if ln.startswith("{") and not ln.startswith('{"summary"')]
        summaries = [ln for ln in proc.stdout.splitlines() if ln.startswith('{"summary"')]
        if proc.returncode != 0 or not summaries:
            raise Stop(f"scripts/type_lint.lean failed (exit {proc.returncode}):\n"
                       f"{proc.stdout[-2000:]}\n{proc.stderr[-2000:]}")
        (self.log_dir / f"lint-{tag}.jsonl").write_text("\n".join(findings) + "\n")
        return findings, json.loads(summaries[-1])


# --- lake output -------------------------------------------------------------

FAILED_RE = re.compile(r"^✖ \[\d+/\d+\] (?:Building|Compiling|Replaying) (\S+)", re.M)
ERROR_RE = re.compile(r"^error: (?:\./)*(\S+?\.lean):(\d+):(\d+): ", re.M)


def failing_files(build_out: str, root: Path) -> list[str]:
    files: list[str] = []
    for m in FAILED_RE.finditer(build_out):
        rel = m.group(1).replace(".", "/") + ".lean"
        if (root / rel).exists() and rel not in files:
            files.append(rel)
    for m in ERROR_RE.finditer(build_out):
        rel = m.group(1)
        if (root / rel).exists() and rel not in files:
            files.append(rel)
    return files


# --- discovery -----------------------------------------------------------------


def discover(lean: Lean, names: list[str], resume: bool) -> list[dict]:
    """The discovery rows. A resumed run cannot import the part-repaired
    tree, so it reads the rows the first run logged."""
    if resume:
        stdout = (lean.log_dir / "discover.log").read_text().split("\n--- stderr ---\n")[0]
        returncode = 0
    else:
        lean.env["CODEMOD_NAMES"] = ",".join(names)
        proc = lean.run(["lake", "env", "lean", str(HERE / "discover.lean")], "discover.log")
        lean.env.pop("CODEMOD_NAMES", None)
        stdout, returncode = proc.stdout, proc.returncode
    rows = [json.loads(ln) for ln in stdout.splitlines() if ln.startswith("{")]
    if returncode != 0 or not rows or not rows[-1].get("summary"):
        raise Stop(f"discover.lean failed (exit {returncode}):\n{stdout[-3000:]}")
    rows = rows[:-1]
    bad = [r for r in rows if "error" in r]
    if bad:
        raise Stop("discovery: " + json.dumps(bad))
    return rows


def byte_to_char(data: bytes, off: int) -> int:
    return len(data[:off].decode("utf-8"))


def reindent(body: str, cont_indent: str) -> str:
    """The body text with its continuation lines re-indented to `cont_indent`."""
    lines = body.split("\n")
    if len(lines) == 1:
        return body
    rest = [ln for ln in lines[1:] if ln.strip()]
    common = min(len(ln) - len(ln.lstrip()) for ln in rest) if rest else 0
    out = [lines[0]]
    for ln in lines[1:]:
        out.append(cont_indent + ln[common:] if ln.strip() else "")
    return "\n".join(out)


def conversion_edit(row: dict, data: bytes) -> tuple[Edit, str]:
    """The edit that turns the `def` of `row` into a structure plus `P.iff`."""
    syn = row["syntax"]
    def ch(b: int) -> int:
        return byte_to_char(data, b)
    text = data.decode("utf-8")
    start = ch(syn["defKeyword"]["start"])
    end = ch(syn["body"]["end"])
    decl_id = syn["declIdText"]
    binders = syn["binders"]
    if binders:
        b_text = " " + text[ch(binders[0]["span"]["start"]):ch(binders[-1]["span"]["end"])]
    else:
        b_text = ""
    if syn["typeSpec"] is not None and syn["typeSpec"]["text"] != "Prop":
        raise Stop(f"{row['name']}: type is `{syn['typeSpec']['text']}`, not `Prop`")
    body = text[ch(syn["body"]["start"]):end]
    explicit = [n for b in binders if b["facts"]["kind"] == "explicit" for n in b["facts"]["names"]]
    if any(n == "_" for n in explicit):
        raise Stop(f"{row['name']}: an explicit binder is `_`; `P.iff` cannot apply it")
    app = " ".join([decl_id, *explicit])
    structure = (f"structure {decl_id}{b_text} : Prop where\n"
                 f"  holds : {reindent(body, '    ')}")
    iff = (f"/-- `{decl_id}`'s structure carries exactly its old `def` body; a citing statement's\n"
           f"    meaning did not change when `{decl_id}` became a structure. -/\n"
           f"theorem {decl_id}.iff{b_text} :\n"
           f"    {app} ↔ {reindent(body, '      ')} :=\n"
           f"  ⟨fun h => h.holds, fun h => ⟨h⟩⟩")
    return Edit(start, end, structure + "\n\n" + iff), app


# --- the run -------------------------------------------------------------------


def git(root: Path, *args: str) -> subprocess.CompletedProcess:
    return subprocess.run(["git", *args], cwd=root, capture_output=True, text=True)


def conflicts(a: Edit, b: Edit) -> bool:
    """Two edits conflict when their ranges overlap or an insertion point
    touches the other edit."""
    if a.start == a.end and b.start == b.end:
        return a.start == b.start
    if a.start == a.end:
        return b.start <= a.start <= b.end
    if b.start == b.end:
        return a.start <= b.start <= a.end
    return max(a.start, b.start) < min(a.end, b.end)


def check_forbidden(before: str, after: str, where: str) -> None:
    for word in FORBIDDEN:
        if after.count(word) > before.count(word):
            raise Stop(f"{where}: the edit would insert `{word}`")


def finding_key(line: str) -> str:
    """A finding line without its `example` field. `repeatedPremise` picks the
    example from the group's uses, and that pick differs between two runs
    on the same tree; the shape and the declarations do not."""
    d = json.loads(line)
    d.pop("example", None)
    return json.dumps(d, ensure_ascii=False, sort_keys=True)


class Run:
    def __init__(self, args: argparse.Namespace) -> None:
        self.root = Path(args.checkout).resolve()
        self.names = [n.strip() for n in args.names.split(",") if n.strip()]
        self.log_dir = self.root / ".scratch" / "codemod"
        self.log_dir.mkdir(parents=True, exist_ok=True)
        self.lean = Lean(self.root, args.slot, self.log_dir)
        self.report_path = Path(args.report).resolve()
        self.resume = args.resume
        self.repairs: list[Repair] = []
        self.rounds: list[dict] = []
        self.seen: set[tuple] = set()
        self.preds: list[Predicate] = []
        self.predicate_rows: list[dict] = []
        self.report: dict = {
            "checkout": str(self.root),
            "head": git(self.root, "rev-parse", "HEAD").stdout.strip(),
            "names": self.names,
            "stopped": None,
            "stops": [],
            "prior_repairs": [],
            "prior_rounds": [],
            "prior_timings": [],
        }
        if self.resume and self.report_path.exists():
            # Keep the earlier segments of a stopped run in the same report:
            # its repairs, its rounds, its timings, and the message it stopped on.
            prior = json.loads(self.report_path.read_text())
            self.report["stops"] = prior.get("stops", []) + ([prior["stopped"]] if prior.get("stopped") else [])
            self.report["prior_repairs"] = prior.get("prior_repairs", []) + prior.get("repairs", [])
            self.report["prior_rounds"] = prior.get("prior_rounds", []) + prior.get("rounds", [])
            self.report["prior_timings"] = prior.get("prior_timings", []) + prior.get("timings", [])

    # -- helpers --

    def write_report(self) -> None:
        self.report["repairs"] = [asdict(r) for r in self.repairs]
        for r in self.report["repairs"]:
            r.pop("edits", None)
        self.report["rounds"] = self.rounds
        self.report["timings"] = self.lean.timings
        self.report["files_changed"] = sorted(
            ln[3:] for ln in git(self.root, "status", "--porcelain").stdout.splitlines())
        self.report_path.write_text(json.dumps(self.report, indent=2, ensure_ascii=False) + "\n")

    def read(self, rel: str) -> str:
        return (self.root / rel).read_text(encoding="utf-8")

    def write(self, rel: str, text: str) -> None:
        if rel.startswith("ArmadaLean/Spec/"):
            raise Stop(f"{rel}: the tool never edits ArmadaLean/Spec/")
        (self.root / rel).write_text(text, encoding="utf-8")

    # -- steps --

    def step_before(self) -> None:
        if self.resume:
            # A stopped run left the tree converted and part-repaired. The
            # "before" measurement is the one the first run recorded.
            log = (self.log_dir / "lint-before.log").read_text().split("\n--- stderr ---\n")[0]
            self.findings_before = [ln for ln in log.splitlines()
                                    if ln.startswith("{") and not ln.startswith('{"summary"')]
            summaries = [ln for ln in log.splitlines() if ln.startswith('{"summary"')]
            if not summaries:
                raise Stop("--resume: no lint-before.log summary in " + str(self.log_dir))
            self.report["lint_before"] = json.loads(summaries[-1])
            self.report["resumed"] = True
            return
        ok, out = self.lean.build("before")
        if not ok:
            raise Stop("the unchanged tree does not build:\n" + out[-3000:])
        self.findings_before, self.report["lint_before"] = self.lean.type_lint("before")

    def step_discover(self) -> None:
        rows = discover(self.lean, self.names, self.resume)
        for r in rows:
            if r["path"].startswith("ArmadaLean/Spec/"):
                raise Stop(f"{r['name']} is declared under ArmadaLean/Spec/; refused")
            short = r["name"].rsplit(".", 1)[-1]
            g = git(self.root, "grep", "-n", rf"Spec\.{short}\b", "--", "ArmadaLean/Spec")
            if g.returncode == 0 and g.stdout.strip():
                raise Stop(f"{r['name']}: an approved statement names Spec.{short}; refused:\n{g.stdout}")
            already = (r["kind"] == "structure" and r["holdsStructure"] and r["hasIff"]) or self.resume
            if not already:
                if r["kind"] != "def" or not r["isProp"]:
                    raise Stop(f"{r['name']} is a {r['kind']} with isProp={r['isProp']}; "
                               "only `def … : Prop` converts")
                if r["parseError"] is not None or r["syntax"] is None:
                    raise Stop(f"{r['name']}: the declaration did not parse as `def … := …`: "
                               f"{r['parseError']}")
                if r["hasIff"]:
                    raise Stop(f"{r['name']}.iff already exists; refused")
            head = (r.get("syntax") or {}).get("bodyHead") or {}
            self.preds.append(Predicate(
                full_name=r["name"],
                decl_id=(r.get("syntax") or {}).get("declIdText") or short,
                body_head=head.get("head"),
                body_is_structure=bool(head.get("whnfHeadIsStructure"))))
            r["already_converted"] = already
            self.predicate_rows.append(r)
        self.report["predicates"] = [
            {"name": r["name"], "path": r["path"], "line": r["range"]["startLine"],
             "already_converted": r["already_converted"]} for r in self.predicate_rows]

    def step_convert(self) -> None:
        by_file: dict[str, list[dict]] = {}
        for r in self.predicate_rows:
            if not r["already_converted"]:
                by_file.setdefault(r["path"], []).append(r)
        for rel, rows in by_file.items():
            data = (self.root / rel).read_bytes()
            text = data.decode("utf-8")
            edits = []
            for r in rows:
                edit, app = conversion_edit(r, data)
                edits.append(edit)
                for p in self.report["predicates"]:
                    if p["name"] == r["name"]:
                        p["iff"] = f"{app} ↔ …"
            new = apply_edits(text, edits)
            check_forbidden(text, new, rel)
            self.write(rel, new)

    def repair_file(self, rel: str, tag: str) -> None:
        """Elaborate `rel` with the probe and repair until it has no error and
        no numeric projection on a converted type."""
        n = 0
        while True:
            n += 1
            msgs, projections = self.lean.probe(rel, self.preds, f"{tag}-{rel.replace('/', '.')}-{n}")
            errors = [m for m in msgs if m.severity == "error"]
            if not errors and not projections:
                return
            text = self.read(rel)
            src = Source(text)
            chosen: list[Repair] = []
            taken: list[Edit] = []
            unmatched: list[dict] = []
            # `t.1` on a converted type elaborates to `P.holds t`, the whole
            # body, where `t.1` meant its first part. Lean reports that far
            # from `t.1`, or not at all. The probe finds every such `.1`.
            for p in projections:
                off = src.offset(p["pos"]["line"], p["pos"]["column"])
                if off == 0 or src.text[off - 1] != ".":
                    raise Stop(f"{rel}:{p['pos']['line']}:{p['pos']['column']}: the probe reports a "
                               "projection, but the source has no `.` before it")
                chosen.append(Repair(rel, p["pos"]["line"], p["pos"]["column"], "holds-proj",
                                     p["predicate"], "numeric projection on the converted type (probe)",
                                     (Edit(off - 1, off - 1, ".holds"),)))
                taken.extend(chosen[-1].edits)
            for m in sorted(errors, key=lambda m: (m.line, m.col), reverse=True):
                r = match(m, src, self.preds)
                if r is None:
                    unmatched.append({"line": m.line, "col": m.col, "message": m.text})
                    continue
                if any(conflicts(e, t) for e in r.edits for t in taken):
                    continue
                key = (rel, r.pattern, r.detail, src.line_text(r.line))
                if key in self.seen:
                    raise Stop(f"{rel}:{r.line}:{r.col}: repair `{r.pattern}` would repeat on "
                               f"`{src.line_text(r.line).strip()}`; the tool stops")
                chosen.append(r)
                taken.extend(r.edits)
            self.rounds.append({"file": rel, "round": n, "errors": len(errors),
                                "projections": len(projections), "repairs": len(chosen),
                                "unmatched": unmatched})
            if not chosen:
                u = unmatched[0]
                raise Stop(f"{rel}:{u['line']}:{u['col']}: no pattern matches:\n{u['message']}")
            edits = [e for r in chosen for e in r.edits]
            new = apply_edits(text, edits)
            check_forbidden(text, new, rel)
            self.write(rel, new)
            for r in chosen:
                self.seen.add((rel, r.pattern, r.detail, src.line_text(r.line)))
                self.repairs.append(r)
                print(f"[repair] {rel}:{r.line}:{r.col} {r.pattern}: {r.detail}", file=sys.stderr)
            self.write_report()

    def step_repair_loop(self) -> None:
        k = 0
        while True:
            k += 1
            ok, out = self.lean.build(f"round{k}")
            if ok:
                return
            files = failing_files(out, self.root)
            if not files:
                raise Stop("lake build failed with no module error:\n" + out[-3000:])
            for rel in files:
                self.repair_file(rel, f"round{k}")

    def step_counterexamples(self) -> None:
        for f in sorted((self.root / "docs" / "counterexamples").glob("*.lean")):
            self.repair_file(str(f.relative_to(self.root)), "counterexample")

    def step_after(self) -> None:
        self.findings_after, self.report["lint_after"] = self.lean.type_lint("after")
        before = set(map(finding_key, self.findings_before))
        after = set(map(finding_key, self.findings_after))
        self.report["findings_removed"] = sorted(before - after)
        self.report["findings_added"] = sorted(after - before)
        p = subprocess.run([sys.executable, "scripts/update_gate_baseline.py", "--hashes-only"],
                           cwd=self.root, capture_output=True, text=True)
        self.report["gate_baseline_hashes"] = {"exit": p.returncode, "stderr": p.stderr[-1000:]}
        proc = self.lean.run(["bash", "scripts/check.sh", "--json"], "check.log")
        try:
            check = json.loads(proc.stdout.strip().splitlines()[-1])
        except (ValueError, IndexError):
            check = {"error": proc.stdout[-2000:] + proc.stderr[-2000:]}
        self.report["check"] = {k: v for k, v in check.items() if k.endswith("_ok")} \
            if "error" not in check else check
        self.report["check_full"] = check

    def main(self) -> int:
        try:
            self.step_before()
            self.step_discover()
            self.step_convert()
            self.write_report()
            self.step_repair_loop()
            self.step_counterexamples()
            self.step_after()
        except Stop as e:
            self.report["stopped"] = str(e)
            self.write_report()
            print(f"STOPPED: {e}", file=sys.stderr)
            return 2
        self.write_report()
        return 0


def parse_args(argv: list[str]) -> argparse.Namespace:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--checkout", required=True, help="armada-lean worktree to edit")
    ap.add_argument("--names", required=True, help="comma-separated predicate names")
    ap.add_argument("--slot", type=int, required=True, help="lean-run lock slot (1-4)")
    ap.add_argument("--report", required=True, help="JSON report path")
    ap.add_argument("--resume", action="store_true",
                    help="continue a stopped run: skip the base build, reuse its lint-before.log")
    return ap.parse_args(argv)


if __name__ == "__main__":
    sys.exit(Run(parse_args(sys.argv[1:])).main())
