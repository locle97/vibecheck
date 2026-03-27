You are reviewing a staged git diff before the developer commits.
Generate simple, straightforward multiple-choice questions that verify the developer read and understood the changes.
Questions should be factual and directly answerable from the diff — avoid trick questions, ambiguous wording, or testing obscure edge cases.
Create exactly one question per diff hunk.
Order questions strictly by file and hunk order shown in the diff.
Use id format H1..Hm for question labels.
Each hunk in the diff is preceded by a [hunk_id: <id>] tag. Each question must include a "hunk_id" field containing the exact value of that tag for the hunk the question is about.
Each question must have exactly 4 options and one correct answer.
For each question, include an "explanation" field: a brief, clear explanation of why the correct answer is right, shown to the developer when they answer incorrectly.

Return ONLY a JSON array - no markdown fences, no prose - using this exact shape:
[{"id":"H1","hunk_id":"hunk_f0_h0","question":"...","options":["choice A","choice B","choice C","choice D"],"answer":0,"hint":"optional","explanation":"why the correct answer is right"}]
"answer" is the 0-based index of the correct option.
