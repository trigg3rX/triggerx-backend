# TriggerX Backend

The **TriggerX Backend** is a decentralized system designed to automate and manage task execution across blockchain networks. It consists of multiple microservices that work together to ensure efficient, reliable, and scalable task orchestration with decentralized validation and consensus.

<p align="center">
  <img src="https://github.com/trigg3rX/triggerx-backend/actions/workflows/build.yml/badge.svg" alt="Build" />
  <img src="https://github.com/trigg3rX/triggerx-backend/actions/workflows/go_lint.yml/badge.svg" alt="Lint" />
</p>

---

## Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) folder:

- **[Architecture](./docs/01-architecture/)** - System design and architecture
- **[Services](./docs/02-services/)** - Service-specific documentation
- **[Infrastructure](./docs/03-infrastructure/)** - Infrastructure setup and configuration
- **[Development](./docs/04-development/)** - Setup guides and development docs
- **[Features](./docs/05-features/)** - Platform features and capabilities

**Quick Links:**

- [System Overview](./docs/overview.md)
- [System Architecture](./docs/01-architecture/a_system_overview.md)
- [Services Overview](./docs/02-services/a_services.md)
- [Development Setup](./docs/04-development/setup/dependencies.md)
- [Data Flow](./docs/01-architecture/d_data_flow.md)

---

## Quick Start

### Prerequisites

- Go 1.25.5
- Node.js v22.6.0 (for Othentic CLI)
- Docker (with compose)
- Just command runner (`justfile`)

### Installation

1. **Clone the repository:**

   ```sh
   git clone https://github.com/trigg3rX/triggerx-backend.git
   cd triggerx-backend
   ```

2. **Install dependencies:**

   ```sh
   go mod tidy
   npm i -g @othentic/cli
   npm i -g @othentic/node
   ```

3. **Configure environment:**

   ```sh
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Set up the database:**

   ```sh
   just db-setup
   ```

### Running Services

Start services in the following order:

```sh
# 1. Start infrastructure
just start-othentic
just start-observability

# 2. Start core services
just start-db-server
just start-taskmonitor
just start-health
just start-time-scheduler
just start-condition-scheduler
just start-taskdispatcher
just start-eventmonitor

# 3. Start keeper nodes
just start-keeper
# Or using the keeper setup repo:
# git clone https://github.com/trigg3rX/triggerx-keeper-setup.git
# cd triggerx-keeper-setup && ./triggerx.sh start
```

**Note:** Ideally all services should be run via `just run-all` (WIP). For now, run them individually as shown above.

For detailed setup instructions, see [Development Setup](./docs/04-development/setup/dependencies.md).

---

## Documentation by Role

### For Developers

- [Development Setup](./docs/04-development/setup/)
- [Testing Guide](./docs/04-development/guides/testing.md)
- [Tracing Guide](./docs/04-development/guides/tracing.md)
- [Developer Notes](./docs/04-development/guides/dev-notes.md)

### For Operators

- [Infrastructure Setup](./docs/03-infrastructure/)
- [ScyllaDB Setup](./docs/03-infrastructure/f_scylla.md)
- [Observability Setup](./docs/03-infrastructure/g_observability.md)

### For Architects

- [System Architecture](./docs/01-architecture/a_system_overview.md)
- [Data Flow](./docs/01-architecture/d_data_flow.md)
- [Service Details](./docs/02-services/a_services.md)

---

## Contributing

We welcome contributions! Please check out our contribution guidelines and code of conduct.

- **Documentation**: See [Documentation Overview](./docs/overview.md)
- **Development**: See [Development Guides](./docs/04-development/)
- **Issues**: Report bugs and request features via GitHub Issues

---

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

---

## Links

- **GitHub**: [trigg3rX/triggerx-backend](https://github.com/trigg3rX/triggerx-backend)
- **Telegram**: [Join our community](https://t.me/triggerxnetwork)
- **Documentation**: [TriggerX Docs](https://triggerx.gitbook.io/triggerx-docs)
- **Keeper Setup**: [triggerx-keeper-setup](https://github.com/trigg3rX/triggerx-keeper-setup)
