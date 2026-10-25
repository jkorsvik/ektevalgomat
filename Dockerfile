# ektevalgomat — two-stage build ending in FROM scratch: built directly
# on the host (`build: .` in docker-compose.yml), no cross-compilation,
# no registry push.
#
# Base image pinned by the same digest the other two public Go stacks
# already use. Bump deliberately, in all three Dockerfiles together:
# `docker pull golang:1.23-alpine && docker inspect --format
# '{{index .RepoDigests 0}}' golang:1.23-alpine`.
#
# Everything — statement/party/evidence YAML, HTML templates, CSS, JS —
# is go:embed'd into the binary at compile time (see internal/data/loader.go
# and web/embed.go), so the final image has nothing on disk but the one
# binary. No shell, no package manager, no libc, no "inside" to get a
# shell in even with code execution. The one external dependency
# (gopkg.in/yaml.v3, for parsing the evidence tree) is resolved from
# go.sum in the builder stage only — never present in the final image.

FROM golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
# Two tests read the contribution files rather than the code: the guard
# that every prefilled issue link names a template that exists
# (.github/ISSUE_TEMPLATE/*.md) and the guard that CONTRIBUTING.md names
# every gate. Without them here, `go test ./...` fails inside the image
# while passing in CI, which is how this line came to exist: the deploy
# on 2026-10-23 failed the build with "CONTRIBUTING.md is the front door
# for a contributor: no such file". They are copied for the test stage
# only; the final image is FROM scratch and carries the one binary.
COPY .github ./.github
COPY CONTRIBUTING.md ./
# The same trap, one level up: TestDocsStateTheRealCounts parses the
# numbers out of README.md and docs/roadmap.md so prose cannot drift from
# the dataset. It was written after finding the README claiming 613
# evidence files against 615, 53 fully coded statements against 54, and
# the roadmap two statements behind. Without these two files the test
# fails inside the image only, which is the failure this comment exists
# to stop repeating.
COPY README.md ./
COPY docs ./docs
# And the trap a third time, from the other direction: the programmes the
# site quotes from are kept as files, and TestArchivedProgrammesMatchYAML
# reads them to prove the hash published on /partier still matches the
# document the quotes came from. Without this line that test fails inside
# the image while passing in CI, which would have turned every deploy red
# after merge. ~23 MB in the builder stage only; the final image is FROM
# scratch and carries the one binary, so nothing here ships.
COPY research/raw/programs ./research/raw/programs
# Evidence YAML is parsed at package initialization — a malformed file is
# a build-time bug. Run the focused test before producing an image so the
# currently-running public site is never replaced by one that can't even
# load its own dataset.
RUN go test ./...
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM scratch
COPY --from=builder /out/server /server
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/server"]
