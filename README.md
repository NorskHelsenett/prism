# PRISM Pentest Report Information Security Management 🛡️💻

Welcome to PRISM Pentest Report Information Security Management (PRISM) – your go-to solution for managing and analyzing pentesting reports with efficiency and security at its core. 🌟

## Overview 🌐

PRISM is a robust tool designed to assist security engineers and pentesters in managing and interpreting pentest reports. It consists of two main components:

1. **API**: Built with GoLang for speed and reliability. 🚀
2. **Web**: Utilizing SvelteKit for a responsive and interactive user interface. 🖥️

## Folder Structure 📁

PRISM's repository is organized as follows:

- 📂 `.cluster`: Contains Helm charts for Kubernetes deployments.
- 📂 `.git`: Git version control directory.
- 📂 `.gitignore`: Specifies intentionally untracked files to ignore.
- 📂 `.gitlab`: GitLab build configuration files.
- 📂 `api`: Source code for the API component.
- 📂 `web`: Source code for the Web interface.

## Getting Started 🚀

### Prerequisites

- GoLang installed for running the API.
- Node.js and Bun (or npm) for the web interface.

### Running the API

Navigate to the `api` directory:

\`\`\`bash
cd api
\`\`\`

Run the API using Go:

\`\`\`bash
go run *.go
\`\`\`

### Running the Web Interface

Navigate to the `web` directory:

\`\`\`bash
cd web
\`\`\`

Run the web interface using Bun (or npm):

\`\`\`bash
bun run dev -- --host 0.0.0.0

# or

npm run dev -- --host 0.0.0.0
\`\`\`

## Contributing 🤝

Contributions to PRISM are welcome! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests.

## Security 🛡️

We take security seriously. If you discover any security-related issues, please email the maintainers or report it in the issues section.

## License 📄

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.

---

Happy Securing! 🔐🚀
