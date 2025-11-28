# Makefile for iqtoolkit-analyzer

.PHONY: help setup sync-version check-version install test lint format clean hooks validate dev-check test-ollama update-requirements

# Python interpreter from virtual environment
PYTHON := .venv/bin/python
PIP := .venv/bin/pip
POETRY := $(shell command -v poetry 2> /dev/null)

# Default target
help:
	@echo "🚀 Iqtoolkit Analyzer - Development Commands"
	@echo ""
	@echo "⚠️  IMPORTANT: All commands require '.venv' directory in repo root!"
	@echo "   First time: make setup  (prefers Poetry, falls back to venv/pip)"
	@echo ""
	@echo "Setup & Installation:"
	@echo "  make validate     Check if environment is properly configured"
	@echo "  make setup        Install development dependencies and git hooks"
	@echo "  make hooks        Install git hooks only" 
	@echo "  make install      Install package in development mode"
	@echo ""
	@echo "Version Management:"
	@echo "  make sync-version Update all files to match VERSION file"
	@echo "  make check-version Verify all versions are consistent"
	@echo ""
	@echo "Dependency Management:"
	@echo "  make update-requirements  Regenerate requirements.txt from pyproject.toml"
	@echo ""
	@echo "Code Quality:"
	@echo "  make format       Format code with black"
	@echo "  make lint         Run linting (flake8, mypy)"
	@echo "  make test         Run tests with coverage"
	@echo "  make test-ollama  Test Ollama setup and integration"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean        Remove build artifacts and cache"

# Setup development environment
setup: hooks install
	@if [ -n "$(POETRY)" ]; then \
		echo "📦 Using Poetry to install (dev,test groups)..."; \
		poetry install --with dev,test; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		echo "📦 Installing development dependencies with pip..."; \
		.venv/bin/pip install -r requirements.txt; \
		.venv/bin/pip install -e .[dev]; \
	fi
	@echo "✅ Development environment ready!"

# Install git hooks
hooks:
	@chmod +x scripts/setup-hooks.sh
	@bash scripts/setup-hooks.sh

# Install package in development mode
install:
	@if [ -n "$(POETRY)" ]; then \
		echo "🚀 Installing with Poetry..."; \
		poetry install; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		echo "🐍 Installing with pip..."; \
		.venv/bin/pip install -r requirements.txt; \
		.venv/bin/pip install -e .; \
	fi

# Version management
sync-version:
	@echo "🔄 Synchronizing versions..."
	@if [ -n "$(POETRY)" ]; then \
		poetry run python scripts/propagate_version.py; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		.venv/bin/pip install -r requirements.txt > /dev/null 2>&1; \
		.venv/bin/python scripts/propagate_version.py; \
	fi

check-version:
	@echo "🔍 Checking version consistency..."
	@if [ -n "$(POETRY)" ]; then \
		poetry run python scripts/propagate_version.py --verify; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		.venv/bin/pip install -r requirements.txt > /dev/null 2>&1; \
		.venv/bin/python scripts/propagate_version.py --verify; \
	fi

# Dependency management
update-requirements:
	@if [ -n "$(POETRY)" ]; then \
		echo "📦 Exporting requirements.txt from Poetry..."; \
		poetry export --without-hashes -f requirements.txt -o requirements.txt; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		echo "📦 Updating requirements.txt using custom script..."; \
		.venv/bin/pip install -r requirements.txt > /dev/null 2>&1; \
		.venv/bin/python scripts/update_requirements.py; \
	fi

# Code formatting

format:
	@echo "🎨 Formatting code..."
	@if [ -n "$(POETRY)" ]; then \
		poetry run black iqtoolkit_analyzer tests scripts *.py; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		.venv/bin/pip install -r requirements.txt > /dev/null 2>&1; \
		.venv/bin/python -m black iqtoolkit_analyzer tests scripts *.py; \
	fi
	@echo "✅ Code formatted!"

# Linting
lint:
	@echo "🔍 Running linting..."
	@if [ -n "$(POETRY)" ]; then \
		poetry run flake8 . --max-line-length=88 --extend-ignore=E203,W503 --exclude=.venv,build,dist,*.egg-info,scripts/propagate_version.py; \
		poetry run mypy iqtoolkit_analyzer --ignore-missing-imports; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		.venv/bin/pip install -r requirements.txt > /dev/null 2>&1; \
		.venv/bin/python -m flake8 . --max-line-length=88 --extend-ignore=E203,W503 --exclude=.venv,build,dist,*.egg-info,scripts/propagate_version.py; \
		.venv/bin/python -m mypy iqtoolkit_analyzer --ignore-missing-imports; \
	fi
	@echo "✅ Linting passed!"

# Run tests
test:
	@echo "🧪 Running tests..."
	@if [ -n "$(POETRY)" ]; then \
		poetry run pytest tests/ --cov=iqtoolkit_analyzer --cov-report=term-missing --cov-report=html; \
	else \
		if [ ! -d ".venv" ]; then \
			echo "📦 Creating '.venv' with standard venv..."; \
			python -m venv .venv; \
		fi; \
		.venv/bin/pip install -r requirements.txt > /dev/null 2>&1; \
		.venv/bin/python -m pytest tests/ --cov=iqtoolkit_analyzer --cov-report=term-missing --cov-report=html; \
	fi
	@echo "✅ Tests completed!"

# Test Ollama setup
test-ollama:
	@echo "🤖 Testing Ollama setup..."
	@if [ -n "$(POETRY)" ]; then \
		poetry run python scripts/test_ollama.py; \
	else \
		.venv/bin/python scripts/test_ollama.py; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf build/ dist/ *.egg-info/
	@rm -rf .pytest_cache/ .coverage htmlcov/
	@find . -name "*.pyc" -delete
	@find . -name "__pycache__" -delete
	@echo "✅ Cleaned up!"

# Validate environment setup
validate:
	@chmod +x scripts/validate-environment.sh
	@bash scripts/validate-environment.sh

# Quick development workflow
dev-check: format lint test check-version
	@echo "✅ All development checks passed!"