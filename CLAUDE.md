# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains mixins for Grafana and Prometheus monitoring stacks. The project is transitioning from Jsonnet-based mixins to using the modern Grafana Foundation SDK and grafanactl tools.

## Technology Stack

### Current Approach (for new mixins)
- **Grafana Foundation SDK**: https://grafana.github.io/grafana-foundation-sdk/
- **grafanactl**: https://grafana.github.io/grafanactl/

These are relatively new tools with limited documentation. Development will involve learning and documenting patterns as mixins are built.

### Legacy Approach (archived)
- **Jsonnet**: The `old-archived/` directory contains Jsonnet-based mixins that have been deprecated
- These older mixins used the ksonnet-util library pattern

## Repository Structure

- **otel-app-semantic/**: Directory for OpenTelemetry application semantic convention mixins (to be developed)
- **old-archived/**: Contains deprecated Jsonnet mixins
  - cert-manager/: Kubernetes cert-manager deployment mixin (Jsonnet)
  - contour/: Contour ingress controller mixin (Jsonnet)  
  - oauth2_proxy/: OAuth2 proxy authentication mixin (Jsonnet)

## Development Guidelines

### Working with Grafana Foundation SDK
The Grafana Foundation SDK provides a programmatic way to generate Grafana resources like dashboards, alerts, and recording rules. Specific patterns and commands will be documented as mixins are developed.

### Working with grafanactl
grafanactl is a CLI tool for managing Grafana resources. Usage patterns will be documented during development.

## Development Commands

### Running a mixin in development mode
```bash
# Run a mixin with hot reload
make dev MIXIN=otel-app-semantic

# This runs: grafanactl resources serve --script 'go run <mixin>/main.go' --watch './<mixin>'
```

## Important Notes

- This repository is in transition from Jsonnet to Grafana Foundation SDK
- Documentation for the new tools is sparse, so development will involve exploration and documentation
- Patterns and best practices will be established and documented as new mixins are created
- The CLAUDE.md file should be updated continuously as new patterns and commands are discovered

## Git Workflow

The repository uses the `master` branch as the main branch. Currently, there are staged deletions of old Jsonnet files that have been moved to the `old-archived/` directory.