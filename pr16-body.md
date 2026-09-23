## 🔗 Linked Issue
Closes #28

## 🏷️ PR Type
- [x] type:bug
- [ ] type:feature
- [ ] type:docs
- [ ] type:refactor
- [ ] type:chore

## 📝 Summary
Wire qa-verify to the QA ledger. En el PR15 se omitió la etapa verify. Al no llamar a qa-finish, el ledger se quedaba esperando y bloqueaba el ciclo en qa-status.

## 📂 Changes
- `skills/qa-verify/SKILL.md`: Se agregó `qa-begin`, `qa-validate` y `qa-finish`.

## 🧪 Test Plan
El spec visual verifica que no hayan errores de ejecución estática.

## ✅ Contributor Checklist
- [x] El issue vinculado tiene `status:approved`
- [x] El PR tiene exactamente un label `type:*` (lo aplica el maintainer, ver abajo)

## Pending maintainer actions
- [ ] `type:bug` label applied to this PR — pending maintainer
