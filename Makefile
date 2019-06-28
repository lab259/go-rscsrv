COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out

EXAMPLES=$(shell ls ./examples/)

$(EXAMPLES): %:
	$(eval EXAMPLE=$*)
	@:

run:
	@if [ ! -z "$(EXAMPLE)" ]; then \
		go run ./examples/$(EXAMPLE); \
	else \
		echo "Usage: make [$(EXAMPLES)] run"; \
		echo "The environment variable \`EXAMPLE\` is not defined."; \
	fi

build:
	@test -d ./examples && $(foreach example,$(EXAMPLES),go build "-ldflags=$(LDFLAGS)" -o ./bin/$(example) -v ./examples/$(example) &&) :

vet:
	@go vet ./...

fmt:
	@go fmt ./...

test:
	@ginkgo --failFast ./...

test-watch:
	@ginkgo watch -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@ginkgo -r -covermode=count --cover --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find . -type f -name *.coverprofile -exec grep -h -v "^mode:" {} >> "${COVERAGEFILE}" \; -exec rm -f {} \;

coverage: coverage-ci
	@sed -i -e "s|_$(CURDIR)/|./|g" "${COVERAGEFILE}"
	@cp "${COVERAGEFILE}" coverage.txt

coverage-html: coverage
	@go tool cover -html="${COVERAGEFILE}" -o .cover/report.html
	@xdg-open .cover/report.html 2> /dev/null > /dev/null

.PHONY: $(EXAMPLES) run build vet fmt test test-watch coverage-ci coverage coverage-html