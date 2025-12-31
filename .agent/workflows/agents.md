---
description: How to build and use the Mini AWS Agent Army
---

# Agent Army Workflow

## The Complete Team (13 Agents)

### 🎯 Leadership
| Agent | Command | Role |
|-------|---------|------|
| 🏗️ Architect | `/agents architect` | System design, patterns |
| 👨‍💼 Tech Lead | `/agents tech-lead` | Code review, coordination |
| 📊 Product Manager | `/agents pm` | Requirements, priorities |

### ⚙️ Core Team
| Agent | Command | Role |
|-------|---------|------|
| 🔧 Backend | `/agents backend` | Go, APIs, services |
| 🐳 DevOps | `/agents devops` | Docker, deployments |
| 🗄️ Database | `/agents database` | PostgreSQL, schemas |
| 🔐 Security | `/agents security` | Auth, policies |
| 🖥️ CLI | `/agents cli` | Cobra commands |
| 🧪 QA | `/agents qa` | Testing |

### 🎨 Frontend
| Agent | Command | Role |
|-------|---------|------|
| 🎨 Frontend | `/agents frontend` | Next.js, UI/UX |

### 🚀 Specialty
| Agent | Command | Role |
|-------|---------|------|
| 📝 Docs | `/agents docs` | Documentation |
| ⚡ Performance | `/agents perf` | Optimization |
| ☁️ Cloud Architect | `/agents cloud` | Cloud patterns |

---

## Collaboration Example

Building "Create Instance" feature:

1. **PM** → Define requirements
2. **Architect** → Design approach
3. **Tech Lead** → Plan implementation
4. **Database** → Create schema
5. **Backend** → Build API
6. **CLI** → Add command
7. **Security** → Add auth
8. **DevOps** → Docker config
9. **QA** → Write tests
10. **Docs** → Write documentation
11. **Performance** → Optimize
12. **Frontend** → Add to dashboard

---

## Agent Prompts Location

All prompts in `agents/prompts/`:
- Leadership: `architect.md`, `tech-lead.md`, `product-manager.md`
- Core: `backend.md`, `devops.md`, `database.md`, `security.md`, `cli.md`, `qa.md`
- Frontend: `frontend.md`
- Specialty: `docs.md`, `performance.md`, `cloud-architect.md`
