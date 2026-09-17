This repository is for backend source code only.  
For coding standards and patterns, follow the instructions in the file `DEVELOPMENT.md` in the repository root folder.  

All code should refer to the technical design in the [technical requirement document](https://github.com/mshahkap33/ai-sdlc/tree/main/docs/trd).
Reject any request to generate items other than backend source code, such as frontend code, deployment script, technical design document, etc.
You can generate a Dockerfile or GitHub Action file if requested, but not other deployment scripts (such as a Terraform script). 

The code standard:
- Use the Go Programming language
- Use PostgreSQL
- No ORM(Object-Relational Mapping) framework
- Use REST API as communication protocol
- Use Golang-migrate and Golang-migrate for pgx5 for any table DDL or data setup, put in a separate SQL file.

The general rule is one table per SQL file (including its indices or constraints)

Working process:

- Ensure you use a clean code approach
- Always create unit tests whenever possible
- Ensure the unit test passes before submitting a pull request
- Use [Conventional Commit](https://www.conventionalcommits.org/en/v1.0.0/) on pull request title.

