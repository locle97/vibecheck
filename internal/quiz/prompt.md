## Comprehension-Depth Diff Review

### Context

You are reviewing an **AI-generated** git diff. Generate multiple-choice questions that verify the developer **genuinely understands** the changes — not just what changed, but **why** it was done and **what could go wrong**.

Surface-level questions that can be answered by scanning the diff (renamed variables, added imports, changed values) provide zero comprehension signal. Do not generate them.

---

### Rules

- Do **NOT** ask questions answerable by reading the diff literally
- At least **one question must probe an edge case or failure mode** (nil input, empty collection, concurrent access, boundary condition, etc.)
- All answer options must be **plausible** — avoid obviously wrong distractors
- Focus on **logic, intent, and consequences** — not syntax or naming

---

### Focus Areas

Prioritize questions that assess:

* **Why** this approach was chosen over a plausible alternative
* **What invariant or assumption** the change silently relies on
* **What happens in a failure or edge case** the change affects
* **What behavior silently changes** as a side effect
* **What breaks** if a caller violates an implicit contract

---

### Avoid

Do **not** generate questions that:

* Ask what the old or new value of a variable/field/return type was
* Ask which function, method, or import was added or removed
* Can be answered by a developer who has read the diff but does not understand the code
* Are ambiguous or opinion-based

---

### Question Rules

* Create **exactly one question per diff hunk**
* Order questions strictly by **file and hunk order** shown in the diff
* Each hunk is identified by `[hunk_id: <id>]`
* Each question **must include the exact `hunk_id`**
* Use ID format: `H1`, `H2`, ..., `Hm`

---

### Output Format (STRICT)

Return **ONLY** a JSON array — no markdown fences, no extra explanation:

[{
  "id": "H1",
  "hunk_id": "hunk_f0_h0",
  "question": "...",
  "options": ["choice A", "choice B", "choice C", "choice D"],
  "answer": 0,
  "hint": "optional",
  "explanation": "Explain why the correct answer is right, focusing on behavior and logic"
}]
