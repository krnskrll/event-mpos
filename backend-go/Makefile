# POS backend — codegen + build, same layout as cuscos-backend/server.
# oapi-codegen is pinned by the go.mod tool directive (v2.4.1); never @latest.
OAPI := go tool oapi-codegen
SERVE_MODULES = health auth me event product catalog stock order backoffice device

.PHONY: api api-models api-serve generate build test

api: api-models api-serve  ## generate shared models + one gin server package per module

api-models:
	@mkdir -p generated/api/models
	$(OAPI) -config oapi-models.yaml api/components/schemas.yml

api-serve:
	@for m in $(SERVE_MODULES); do \
	  [ -f api/serve/$${m}.yml ] || continue ; \
	  mkdir -p generated/api/$${m} ; \
	  $(OAPI) -generate gin-server,models \
	    -package $${m}api \
	    -import-mapping ../components/schemas.yml:github.com/krnskrll/event-mpos/generated/api/models \
	    -o generated/api/$${m}/$${m}.gen.go \
	    api/serve/$${m}.yml || exit 1 ; \
	done

generate: api

build: generate
	go build ./...

test: generate
	go test ./...
